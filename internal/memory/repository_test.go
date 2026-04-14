package memory_test

import (
	"context"
	"testing"

	"github.com/rodascaar/synkro/internal/db"
	"github.com/rodascaar/synkro/internal/embeddings"
	"github.com/rodascaar/synkro/internal/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*db.Database, func()) {
	d, err := db.New(":memory:")
	require.NoError(t, err)

	cleanup := func() { d.Close() }
	return d, cleanup
}

func TestRepository_Create(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())

	mem := &memory.Memory{
		Type:    "note",
		Title:   "Test Title",
		Content: "Test Content",
		Source:  "test",
		Status:  "active",
		Tags:    []string{"tag1", "tag2"},
	}

	err := repo.Create(context.Background(), mem)
	assert.NoError(t, err)
	assert.NotEmpty(t, mem.ID)
	assert.False(t, mem.CreatedAt.IsZero())
}

func TestRepository_Search(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())

	for i := 0; i < 5; i++ {
		mem := &memory.Memory{
			Type:    "note",
			Title:   "Title " + string(rune('0'+i)),
			Content: "Content " + string(rune('0'+i)),
			Source:  "test",
			Status:  "active",
		}
		require.NoError(t, repo.Create(context.Background(), mem))
	}

	memories, err := repo.Search(context.Background(), "", memory.MemoryFilter{})
	assert.NoError(t, err)
	assert.Len(t, memories, 5)

	memories, err = repo.Search(context.Background(), "Title", memory.MemoryFilter{})
	assert.NoError(t, err)
	assert.Greater(t, len(memories), 0)
	assert.Contains(t, memories[0].Title, "Title")
}

func TestRepository_Get(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())

	mem := &memory.Memory{
		Type:    "note",
		Title:   "Test Get",
		Content: "Test Content",
		Source:  "test",
		Status:  "active",
	}

	require.NoError(t, repo.Create(context.Background(), mem))

	fetched, err := repo.Get(context.Background(), mem.ID)
	assert.NoError(t, err)
	assert.NotNil(t, fetched)
	assert.Equal(t, mem.Title, fetched.Title)
}

func TestRepository_Update(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())

	mem := &memory.Memory{
		Type:    "note",
		Title:   "Test Update",
		Content: "Test Content",
		Source:  "test",
		Status:  "active",
	}

	require.NoError(t, repo.Create(context.Background(), mem))

	newTitle := "Updated Title"
	update := &memory.MemoryUpdate{
		Title: &newTitle,
	}

	err := repo.Update(context.Background(), mem.ID, update)
	assert.NoError(t, err)

	fetched, err := repo.Get(context.Background(), mem.ID)
	assert.NoError(t, err)
	assert.Equal(t, newTitle, fetched.Title)
}

func TestRepository_Delete(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())

	mem := &memory.Memory{
		Type:    "note",
		Title:   "Test Delete",
		Content: "Test Content",
		Source:  "test",
		Status:  "active",
	}

	require.NoError(t, repo.Create(context.Background(), mem))

	err := repo.Delete(context.Background(), mem.ID)
	assert.NoError(t, err)

	fetched, err := repo.Get(context.Background(), mem.ID)
	assert.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestRepository_HybridSearch(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())
	repo.SetEmbeddingGenerator(embeddings.NewTFIDFEmbeddingGenerator(nil))

	mem := &memory.Memory{
		Type:    "decision",
		Title:   "Use SQLite",
		Content: "Decided to use SQLite with FTS5 for Synkro",
		Source:  "test",
		Status:  "active",
	}

	require.NoError(t, repo.Create(context.Background(), mem))

	results, err := repo.HybridSearch(context.Background(), "SQLite", 10, memory.HybridSearchFilter{
		Type:   "decision",
		Status: "active",
	})
	assert.NoError(t, err)
	assert.Greater(t, len(results), 0)
}
