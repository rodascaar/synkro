package mcp_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rodascaar/synkro/internal/db"
	"github.com/rodascaar/synkro/internal/mcp"
	"github.com/rodascaar/synkro/internal/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) (*mcp.Server, *memory.Repository) {
	t.Helper()
	tmpFile := t.TempDir() + "/test.db"
	database, err := db.New(tmpFile)
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })

	memRepo := memory.NewRepository(database.DB())
	server := mcp.NewServer(memRepo, nil, nil, nil)

	return server, memRepo
}

func TestHandlers_AddAndGetMemory(t *testing.T) {
	server, _ := setupTestServer(t)

	var buf mcp.BufferWriter
	err := server.AddMemoryWithWriter(context.Background(), mcp.AddMemoryInput{
		Type:    "note",
		Title:   "Test Note",
		Content: "Test content here",
		Source:  "test",
		Tags:    []string{"tag1", "tag2"},
	}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	assert.Equal(t, true, response["success"])
	assert.NotEmpty(t, response["memory_id"])

	memID := response["memory_id"].(string)

	buf.Reset()
	err = server.GetMemory(context.Background(), mcp.GetMemoryInput{ID: memID}, &buf)
	require.NoError(t, err)

	var getResult map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &getResult))
	mem := getResult["memory"].(map[string]interface{})
	assert.Equal(t, "Test Note", mem["title"])
	assert.Equal(t, "Test content here", mem["content"])
	assert.Equal(t, "note", mem["type"])
}

func TestHandlers_ListMemories(t *testing.T) {
	server, memRepo := setupTestServer(t)

	for i := 0; i < 3; i++ {
		ctx := context.Background()
		mem := &memory.Memory{
			Type:    "note",
			Title:   "Note " + string(rune('A'+i)),
			Content: "Content " + string(rune('A'+i)),
			Status:  "active",
		}
		require.NoError(t, memRepo.Create(ctx, mem))
	}

	var buf mcp.BufferWriter
	err := server.ListMemory(context.Background(), mcp.ListMemoryInput{Limit: 10}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	assert.Equal(t, float64(3), response["count"])
}

func TestHandlers_SearchMemories(t *testing.T) {
	server, memRepo := setupTestServer(t)

	ctx := context.Background()
	memRepo.Create(ctx, &memory.Memory{
		Type: "note", Title: "Database Design", Content: "PostgreSQL architecture patterns", Status: "active",
	})
	memRepo.Create(ctx, &memory.Memory{
		Type: "note", Title: "Cooking Recipe", Content: "How to bake a cake", Status: "active",
	})

	var buf mcp.BufferWriter
	err := server.SearchMemory(context.Background(), mcp.SearchMemoryInput{Query: "database", Limit: 10}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	assert.True(t, response["count"].(float64) >= 1)
}

func TestHandlers_UpdateMemory(t *testing.T) {
	server, memRepo := setupTestServer(t)

	ctx := context.Background()
	mem := &memory.Memory{
		Type: "note", Title: "Original", Content: "Original content", Status: "active",
	}
	require.NoError(t, memRepo.Create(ctx, mem))

	var buf mcp.BufferWriter
	err := server.UpdateMemory(context.Background(), mcp.UpdateMemoryInput{
		ID:      mem.ID,
		Title:   "Updated",
		Content: "New content",
	}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	assert.Equal(t, true, response["success"])

	updated, err := memRepo.Get(ctx, mem.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Title)
}

func TestHandlers_ArchiveMemory(t *testing.T) {
	server, memRepo := setupTestServer(t)

	ctx := context.Background()
	mem := &memory.Memory{
		Type: "note", Title: "To Archive", Content: "Will be archived", Status: "active",
	}
	require.NoError(t, memRepo.Create(ctx, mem))

	var buf mcp.BufferWriter
	err := server.ArchiveMemory(context.Background(), mcp.ArchiveMemoryInput{ID: mem.ID}, &buf)
	require.NoError(t, err)

	archived, err := memRepo.Get(ctx, mem.ID)
	require.NoError(t, err)
	assert.Equal(t, "archived", archived.Status)
}

func TestHandlers_GetMemory_NotFound(t *testing.T) {
	server, _ := setupTestServer(t)

	var buf mcp.BufferWriter
	err := server.GetMemory(context.Background(), mcp.GetMemoryInput{ID: "nonexistent"}, &buf)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "MEM_NOT_FOUND"))
}

func TestHandlers_ActivateContext_Empty(t *testing.T) {
	server, _ := setupTestServer(t)

	var buf mcp.BufferWriter
	err := server.ActivateContext(context.Background(), mcp.ActivateContextInput{
		Query:     "nonexistent query that matches nothing",
		SessionID: "test-session",
	}, &buf)
	if err != nil {
		assert.Contains(t, err.Error(), "embedding generator")
		return
	}
	assert.Contains(t, buf.String(), "No memories found")
}
