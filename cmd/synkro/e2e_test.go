package main

import (
	"context"
	"testing"

	"github.com/rodascaar/synkro/internal/db"
	"github.com/rodascaar/synkro/internal/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_CreateSearchDelete(t *testing.T) {
	d, err := db.New(":memory:")
	require.NoError(t, err)
	defer func() { _ = d.Close() }()

	repo := memory.NewRepository(d.DB())

	mem := &memory.Memory{
		Type:    "decision",
		Title:   "Use Go for backend",
		Content: "Go is fast and has great concurrency",
		Status:  "active",
		Tags:    []string{"go", "backend"},
	}

	err = repo.Create(context.Background(), mem)
	require.NoError(t, err)
	assert.NotEmpty(t, mem.ID)
	assert.Equal(t, []string{"go", "backend"}, mem.Tags)

	fetched, err := repo.Get(context.Background(), mem.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, "Use Go for backend", fetched.Title)

	results, err := repo.Search(context.Background(), "", memory.MemoryFilter{Type: "decision"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, mem.ID, results[0].ID)

	tagResults, err := repo.GetByTag(context.Background(), "go", memory.MemoryFilter{})
	require.NoError(t, err)
	require.Len(t, tagResults, 1)
	assert.Equal(t, mem.ID, tagResults[0].ID)

	err = repo.Delete(context.Background(), mem.ID)
	require.NoError(t, err)

	deleted, err := repo.Get(context.Background(), mem.ID)
	require.NoError(t, err)
	assert.Nil(t, deleted)
}

func TestE2E_UpdateWithTags(t *testing.T) {
	d, err := db.New(":memory:")
	require.NoError(t, err)
	defer func() { _ = d.Close() }()

	repo := memory.NewRepository(d.DB())

	mem := &memory.Memory{
		Type:    "note",
		Title:   "Original",
		Content: "Content",
		Status:  "active",
		Tags:    []string{"old"},
	}
	require.NoError(t, repo.Create(context.Background(), mem))

	newTitle := "Updated"
	newContent := "New content"
	newTags := []string{"new1", "new2"}
	err = repo.Update(context.Background(), mem.ID, &memory.MemoryUpdate{
		Title:   &newTitle,
		Content: &newContent,
		Tags:    newTags,
	})
	require.NoError(t, err)

	fetched, err := repo.Get(context.Background(), mem.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", fetched.Title)
	assert.Equal(t, "New content", fetched.Content)
	assert.Equal(t, newTags, fetched.Tags)
}

func TestE2E_SearchEmptyReturnsAll(t *testing.T) {
	d, err := db.New(":memory:")
	require.NoError(t, err)
	defer func() { _ = d.Close() }()

	repo := memory.NewRepository(d.DB())

	for i := 0; i < 5; i++ {
		mem := &memory.Memory{
			Type:    "note",
			Title:   "Note " + string(rune('0'+i)),
			Content: "Content",
			Status:  "active",
		}
		require.NoError(t, repo.Create(context.Background(), mem))
	}

	memories, err := repo.Search(context.Background(), "", memory.MemoryFilter{Limit: 100})
	require.NoError(t, err)
	assert.Len(t, memories, 5)

	memories, err = repo.Search(context.Background(), "", memory.MemoryFilter{Limit: 2})
	require.NoError(t, err)
	assert.Len(t, memories, 2)
}

func TestE2E_DeleteNonexistent(t *testing.T) {
	d, err := db.New(":memory:")
	require.NoError(t, err)
	defer func() { _ = d.Close() }()

	repo := memory.NewRepository(d.DB())

	err = repo.Delete(context.Background(), "nonexistent-id")
	assert.NoError(t, err, "deleting nonexistent should not error")
}
