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

	cleanup := func() { _ = d.Close() }
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

func TestRepository_HybridSearch_MultipleResults(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())
	repo.SetEmbeddingGenerator(embeddings.NewTFIDFEmbeddingGenerator(nil))

	memories := []*memory.Memory{
		{Type: "note", Title: "Go Concurrency", Content: "Goroutines and channels for concurrent programming", Status: "active"},
		{Type: "note", Title: "Python Async", Content: "Asyncio and await for asynchronous programming", Status: "active"},
		{Type: "note", Title: "Rust Ownership", Content: "Ownership and borrowing in Rust", Status: "active"},
	}
	for _, mem := range memories {
		require.NoError(t, repo.Create(context.Background(), mem))
	}

	results, err := repo.HybridSearch(context.Background(), "concurrent programming", 5, memory.HybridSearchFilter{Status: "active"})
	assert.NoError(t, err)
	assert.Greater(t, len(results), 0)

	for _, r := range results {
		assert.Equal(t, "active", r.Memory.Status)
	}
}

func TestRepository_GetByTag(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())

	mem1 := &memory.Memory{Type: "note", Title: "Tagged Note", Content: "Content", Status: "active", Tags: []string{"go", "programming"}}
	mem2 := &memory.Memory{Type: "note", Title: "Untagged Note", Content: "Content", Status: "active", Tags: []string{"python"}}
	require.NoError(t, repo.Create(context.Background(), mem1))
	require.NoError(t, repo.Create(context.Background(), mem2))

	results, err := repo.GetByTag(context.Background(), "go", memory.MemoryFilter{})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "Tagged Note", results[0].Title)

	results, err = repo.GetByTag(context.Background(), "nonexistent", memory.MemoryFilter{})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(results))
}

func TestRepository_GetByTag_WithTypeFilter(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())

	mem1 := &memory.Memory{Type: "decision", Title: "Decision A", Content: "Content", Status: "active", Tags: []string{"important"}}
	mem2 := &memory.Memory{Type: "note", Title: "Note B", Content: "Content", Status: "active", Tags: []string{"important"}}
	require.NoError(t, repo.Create(context.Background(), mem1))
	require.NoError(t, repo.Create(context.Background(), mem2))

	results, err := repo.GetByTag(context.Background(), "important", memory.MemoryFilter{Type: "decision"})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "Decision A", results[0].Title)
}

func TestRepository_GetByTag_WithLimit(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())

	for i := 0; i < 5; i++ {
		mem := &memory.Memory{Type: "note", Title: "Tag " + string(rune('A'+i)), Content: "Content", Status: "active", Tags: []string{"limited"}}
		require.NoError(t, repo.Create(context.Background(), mem))
	}

	results, err := repo.GetByTag(context.Background(), "limited", memory.MemoryFilter{Limit: 3})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(results))
}

func TestRepository_Get_NotFound(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())
	mem, err := repo.Get(context.Background(), "nonexistent-id")
	assert.NoError(t, err)
	assert.Nil(t, mem)
}

func TestRepository_Create_EmptyTags(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())

	mem := &memory.Memory{
		Type:    "task",
		Title:   "No Tags",
		Content: "Content",
		Source:  "test",
		Status:  "active",
		Tags:    nil,
	}

	require.NoError(t, repo.Create(context.Background(), mem))
	assert.NotEmpty(t, mem.ID)

	fetched, err := repo.Get(context.Background(), mem.ID)
	require.NoError(t, err)
	assert.Equal(t, "task", fetched.Type)
	assert.Empty(t, fetched.Tags)
}

func TestRepository_Update_MultipleFields(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())

	mem := &memory.Memory{
		Type: "note", Title: "Original", Content: "Original content",
		Source: "test", Status: "active",
	}
	require.NoError(t, repo.Create(context.Background(), mem))

	newTitle := "Updated Title"
	newContent := "Updated content"
	newStatus := "archived"
	update := &memory.MemoryUpdate{
		Title:   &newTitle,
		Content: &newContent,
		Status:  &newStatus,
	}

	require.NoError(t, repo.Update(context.Background(), mem.ID, update))

	fetched, err := repo.Get(context.Background(), mem.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", fetched.Title)
	assert.Equal(t, "Updated content", fetched.Content)
	assert.Equal(t, "archived", fetched.Status)
}

func TestRepository_Update_WithTags(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())

	mem := &memory.Memory{
		Type: "note", Title: "Tag Update Test", Content: "Content",
		Source: "test", Status: "active", Tags: []string{"old-tag"},
	}
	require.NoError(t, repo.Create(context.Background(), mem))

	update := &memory.MemoryUpdate{
		Tags: []string{"new-tag-1", "new-tag-2"},
	}
	require.NoError(t, repo.Update(context.Background(), mem.ID, update))

	fetched, err := repo.Get(context.Background(), mem.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"new-tag-1", "new-tag-2"}, fetched.Tags)
}

func TestRepository_Search_WithFilters(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())

	require.NoError(t, repo.Create(context.Background(), &memory.Memory{
		Type: "decision", Title: "DB Decision", Content: "Use PostgreSQL", Status: "active",
	}))
	require.NoError(t, repo.Create(context.Background(), &memory.Memory{
		Type: "note", Title: "DB Note", Content: "Connection pooling", Status: "archived",
	}))
	require.NoError(t, repo.Create(context.Background(), &memory.Memory{
		Type: "task", Title: "DB Task", Content: "Migrate schema", Status: "active",
	}))

	memories, err := repo.Search(context.Background(), "", memory.MemoryFilter{Type: "decision"})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(memories))
	assert.Equal(t, "DB Decision", memories[0].Title)

	memories, err = repo.Search(context.Background(), "", memory.MemoryFilter{Status: "archived"})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(memories))

	memories, err = repo.Search(context.Background(), "", memory.MemoryFilter{Type: "decision", Status: "active"})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(memories))

	memories, err = repo.Search(context.Background(), "", memory.MemoryFilter{Limit: 2})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(memories))
}

func TestRepository_Delete_NotFound(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())

	err := repo.Delete(context.Background(), "nonexistent-id")
	assert.NoError(t, err)
}

func TestRepository_Create_AssignsID(t *testing.T) {
	d, cleanup := setupTestDB(t)
	defer cleanup()

	repo := memory.NewRepository(d.DB())

	mem := &memory.Memory{Type: "note", Title: "Auto ID", Content: "Content", Status: "active"}
	require.NoError(t, repo.Create(context.Background(), mem))
	assert.True(t, len(mem.ID) > 0)
	assert.Contains(t, mem.ID, "mem-")

	mem2 := &memory.Memory{Type: "note", Title: "Auto ID 2", Content: "Content", Status: "active"}
	require.NoError(t, repo.Create(context.Background(), mem2))
	assert.NotEqual(t, mem.ID, mem2.ID)
}
