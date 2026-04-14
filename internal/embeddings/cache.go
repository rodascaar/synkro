package embeddings

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	db      *sql.DB
	lru     *sync.Map
	maxSize int
}

func NewCache(db *sql.DB, maxSize int) *Cache {
	c := &Cache{
		db:      db,
		lru:     &sync.Map{},
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

	for rows.Next() {
		var text string
		var embeddingBytes []byte
		if err := rows.Scan(&text, &embeddingBytes); err != nil {
			continue
		}

		embedding := DeserializeEmbedding(embeddingBytes)
		if embedding != nil {
			c.lru.Store(c.hashText(text), embedding)
		}
	}
}

func (c *Cache) Get(ctx context.Context, text string) ([]float32, bool) {
	hash := c.hashText(text)

	if value, ok := c.lru.Load(hash); ok {
		if embedding, ok := value.([]float32); ok {
			return embedding, true
		}
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

	c.lru.Store(hash, embedding)

	c.pruneIfNeeded()

	return nil
}

func (c *Cache) Prune(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM embedding_cache WHERE created_at < datetime('now', -30 days)")
	return err
}

func (c *Cache) Clear(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM embedding_cache")
	if err != nil {
		return err
	}

	c.lru = &sync.Map{}
	return nil
}

func (c *Cache) pruneIfNeeded() {
	count := 0
	c.lru.Range(func(_, _ interface{}) bool {
		count++
		return count < c.maxSize
	})

	if count >= c.maxSize {
		c.lru = &sync.Map{}
	}
}

func (c *Cache) hashText(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}
