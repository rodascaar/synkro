package embeddings

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCacheDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS embedding_cache (
		text_hash TEXT PRIMARY KEY,
		text TEXT NOT NULL,
		embedding BLOB NOT NULL,
		model_type TEXT,
		created_at TEXT
	)`)
	require.NoError(t, err)

	return db, func() { _ = db.Close() }
}

func TestCache_GetMiss(t *testing.T) {
	sqlDB, cleanup := setupCacheDB(t)
	defer cleanup()

	cache := NewCache(sqlDB, 10)

	_, ok := cache.Get(context.Background(), "nonexistent")
	assert.False(t, ok)
}

func TestCache_SetAndGet(t *testing.T) {
	sqlDB, cleanup := setupCacheDB(t)
	defer cleanup()

	cache := NewCache(sqlDB, 10)
	embedding := []float32{0.1, 0.2, 0.3}

	err := cache.Set(context.Background(), "hello world", embedding, "tfidf")
	require.NoError(t, err)

	got, ok := cache.Get(context.Background(), "hello world")
	assert.True(t, ok)
	assert.Equal(t, embedding, got)
}

func TestCache_Eviction(t *testing.T) {
	sqlDB, cleanup := setupCacheDB(t)
	defer cleanup()

	cache := NewCache(sqlDB, 3)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		err := cache.Set(ctx, string(rune('a'+i)), []float32{float32(i)}, "tfidf")
		require.NoError(t, err)
	}

	_, ok := cache.Get(ctx, "a")
	assert.False(t, ok, "oldest item should be evicted")

	_, ok = cache.Get(ctx, "b")
	assert.False(t, ok, "second oldest should be evicted")

	got, ok := cache.Get(ctx, "c")
	assert.True(t, ok, "third item should remain")
	assert.Equal(t, []float32{2.0}, got)

	got, ok = cache.Get(ctx, "e")
	assert.True(t, ok, "newest item should remain")
	assert.Equal(t, []float32{4.0}, got)
}

func TestCache_LRUMoveToFront(t *testing.T) {
	sqlDB, cleanup := setupCacheDB(t)
	defer cleanup()

	cache := NewCache(sqlDB, 3)
	ctx := context.Background()

	require.NoError(t, cache.Set(ctx, "a", []float32{1}, "tfidf"))
	require.NoError(t, cache.Set(ctx, "b", []float32{2}, "tfidf"))
	require.NoError(t, cache.Set(ctx, "c", []float32{3}, "tfidf"))

	_, ok := cache.Get(ctx, "a")
	assert.True(t, ok)

	require.NoError(t, cache.Set(ctx, "d", []float32{4}, "tfidf"))

	_, ok = cache.Get(ctx, "b")
	assert.False(t, ok, "b should be evicted, a was accessed recently")

	got, ok := cache.Get(ctx, "a")
	assert.True(t, ok, "a should remain after LRU promotion")
	assert.Equal(t, []float32{1}, got)
}

func TestCache_Clear(t *testing.T) {
	sqlDB, cleanup := setupCacheDB(t)
	defer cleanup()

	cache := NewCache(sqlDB, 10)
	ctx := context.Background()

	require.NoError(t, cache.Set(ctx, "test", []float32{1}, "tfidf"))

	err := cache.Clear(ctx)
	require.NoError(t, err)

	_, ok := cache.Get(ctx, "test")
	assert.False(t, ok)
}

func TestCache_Persistence(t *testing.T) {
	sqlDB, cleanup := setupCacheDB(t)
	defer cleanup()

	cache := NewCache(sqlDB, 10)
	ctx := context.Background()

	embedding := []float32{0.5, 0.6, 0.7}
	require.NoError(t, cache.Set(ctx, "persist me", embedding, "tfidf"))

	cache2 := NewCache(sqlDB, 10)
	got, ok := cache2.Get(ctx, "persist me")
	assert.True(t, ok)
	assert.Equal(t, embedding, got)
}

func TestCache_DifferentTextsSameHash(t *testing.T) {
	sqlDB, cleanup := setupCacheDB(t)
	defer cleanup()

	cache := NewCache(sqlDB, 10)
	ctx := context.Background()

	require.NoError(t, cache.Set(ctx, "text one", []float32{1}, "tfidf"))
	require.NoError(t, cache.Set(ctx, "text two", []float32{2}, "tfidf"))

	got1, ok1 := cache.Get(ctx, "text one")
	assert.True(t, ok1)
	assert.Equal(t, []float32{1}, got1)

	got2, ok2 := cache.Get(ctx, "text two")
	assert.True(t, ok2)
	assert.Equal(t, []float32{2}, got2)
}

func TestCache_EmptyEmbedding(t *testing.T) {
	sqlDB, cleanup := setupCacheDB(t)
	defer cleanup()

	cache := NewCache(sqlDB, 10)
	ctx := context.Background()

	require.NoError(t, cache.Set(ctx, "empty", []float32{}, "tfidf"))

	got, ok := cache.Get(ctx, "empty")
	assert.True(t, ok)
	assert.Empty(t, got)
}

func TestCache_UpdateExisting(t *testing.T) {
	sqlDB, cleanup := setupCacheDB(t)
	defer cleanup()

	cache := NewCache(sqlDB, 10)
	ctx := context.Background()

	require.NoError(t, cache.Set(ctx, "update", []float32{1}, "tfidf"))
	require.NoError(t, cache.Set(ctx, "update", []float32{2}, "tfidf"))

	got, ok := cache.Get(ctx, "update")
	assert.True(t, ok)
	assert.Equal(t, []float32{2}, got)
}

func TestCache_Prune(t *testing.T) {
	sqlDB, cleanup := setupCacheDB(t)
	defer cleanup()

	cache := NewCache(sqlDB, 10)
	ctx := context.Background()

	require.NoError(t, cache.Set(ctx, "old", []float32{1}, "tfidf"))

	err := cache.Prune(ctx)
	require.NoError(t, err)

	_, ok := cache.Get(ctx, "old")
	assert.False(t, ok, "old entry should be pruned")
}
