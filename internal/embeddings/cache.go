package embeddings

import (
	"container/list"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type cacheEntry struct {
	key       string
	value     []float32
	timestamp time.Time
}

type Cache struct {
	db      *sql.DB
	mu      sync.Mutex
	items   map[string]*list.Element
	order   *list.List
	maxSize int
}

func NewCache(db *sql.DB, maxSize int) *Cache {
	c := &Cache{
		db:      db,
		items:   make(map[string]*list.Element),
		order:   list.New(),
		maxSize: maxSize,
	}

	c.loadFromDB(context.Background())
	return c
}

func (c *Cache) loadFromDB(ctx context.Context) {
	rows, err := c.db.QueryContext(ctx, "SELECT text, embedding FROM embedding_cache")
	if err != nil {
		return
	}
	defer rows.Close()

	c.mu.Lock()
	defer c.mu.Unlock()

	for rows.Next() {
		var text string
		var embeddingBytes []byte
		if err := rows.Scan(&text, &embeddingBytes); err != nil {
			continue
		}

		embedding := DeserializeEmbedding(embeddingBytes)
		if embedding != nil {
			hash := c.hashText(text)
			entry := &cacheEntry{key: hash, value: embedding, timestamp: time.Now().UTC()}
			elem := c.order.PushFront(entry)
			c.items[hash] = elem
		}
	}
}

func (c *Cache) Get(_ context.Context, text string) ([]float32, bool) {
	hash := c.hashText(text)

	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[hash]; ok {
		c.order.MoveToFront(elem)
		return elem.Value.(*cacheEntry).value, true
	}

	return nil, false
}

func (c *Cache) Set(ctx context.Context, text string, embedding []float32, modelType string) error {
	hash := c.hashText(text)
	serialized := SerializeEmbedding(embedding)
	now := time.Now().UTC()

	_, err := c.db.ExecContext(ctx, `
		INSERT INTO embedding_cache (text_hash, text, embedding, model_type, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(text_hash) DO UPDATE SET
			embedding = excluded.embedding,
			model_type = excluded.model_type
	`, hash, text, serialized, modelType, now.Format(time.RFC3339))

	if err != nil {
		return fmt.Errorf("failed to save embedding to cache: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[hash]; ok {
		c.order.MoveToFront(elem)
		elem.Value.(*cacheEntry).value = embedding
		elem.Value.(*cacheEntry).timestamp = now
	} else {
		entry := &cacheEntry{key: hash, value: embedding, timestamp: now}
		elem := c.order.PushFront(entry)
		c.items[hash] = elem
		c.evict()
	}

	return nil
}

func (c *Cache) Prune(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM embedding_cache WHERE created_at < datetime('now', -30 days)")
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -30)
	var toRemove []string
	for e := c.order.Back(); e != nil; e = e.Prev() {
		entry := e.Value.(*cacheEntry)
		if entry.timestamp.Before(cutoff) {
			toRemove = append(toRemove, entry.key)
		} else {
			break
		}
	}
	for _, key := range toRemove {
		if elem, ok := c.items[key]; ok {
			c.order.Remove(elem)
			delete(c.items, key)
		}
	}

	return nil
}

func (c *Cache) Clear(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM embedding_cache")
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.order.Init()
	return nil
}

func (c *Cache) evict() {
	for c.order.Len() > c.maxSize {
		oldest := c.order.Back()
		if oldest == nil {
			break
		}
		entry := c.order.Remove(oldest).(*cacheEntry)
		delete(c.items, entry.key)
	}
}

func (c *Cache) hashText(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}
