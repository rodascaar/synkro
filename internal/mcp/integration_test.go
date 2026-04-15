package mcp_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rodascaar/synkro/internal/db"
	"github.com/rodascaar/synkro/internal/graph"
	"github.com/rodascaar/synkro/internal/mcp"
	"github.com/rodascaar/synkro/internal/memory"
	"github.com/rodascaar/synkro/internal/pruner"
	"github.com/rodascaar/synkro/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) (*mcp.Server, *memory.Repository) {
	t.Helper()
	tmpFile := t.TempDir() + "/test.db"
	database, err := db.New(tmpFile)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	memRepo := memory.NewRepository(database.DB())
	server := mcp.NewServer(memRepo, nil, nil, nil)

	return server, memRepo
}

func setupTestServerWithGraph(t *testing.T) (*mcp.Server, *memory.Repository) {
	t.Helper()
	tmpFile := t.TempDir() + "/test.db"
	database, err := db.New(tmpFile)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	memRepo := memory.NewRepository(database.DB())
	graphRepo := graph.NewRepository(database.DB())
	g := graph.NewGraph(memRepo, graphRepo)
	sessionRepo := session.NewRepository(database.DB())
	st := session.NewSessionTracker(sessionRepo)
	cp := pruner.NewContextPruner()

	server := mcp.NewServer(memRepo, g, st, cp)
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
	_ = memRepo.Create(ctx, &memory.Memory{
		Type: "note", Title: "Database Design", Content: "PostgreSQL architecture patterns", Status: "active",
	})
	_ = memRepo.Create(ctx, &memory.Memory{
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

func TestHandlers_AddRelation(t *testing.T) {
	server, memRepo := setupTestServerWithGraph(t)
	ctx := context.Background()

	mem1 := &memory.Memory{Type: "note", Title: "Memory 1", Content: "Content 1", Status: "active"}
	mem2 := &memory.Memory{Type: "note", Title: "Memory 2", Content: "Content 2", Status: "active"}
	require.NoError(t, memRepo.Create(ctx, mem1))
	require.NoError(t, memRepo.Create(ctx, mem2))

	var buf mcp.BufferWriter
	err := server.AddRelation(ctx, mcp.AddRelationInput{
		SourceID: mem1.ID,
		TargetID: mem2.ID,
		Type:     "related_to",
		Strength: 0.8,
	}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	assert.Equal(t, true, response["success"])
	assert.Equal(t, mem1.ID, response["source_id"])
	assert.Equal(t, mem2.ID, response["target_id"])
	assert.Equal(t, "related_to", response["type"])
}

func TestHandlers_AddRelation_InvalidType(t *testing.T) {
	server, _ := setupTestServerWithGraph(t)

	var buf mcp.BufferWriter
	err := server.AddRelation(context.Background(), mcp.AddRelationInput{
		SourceID: "mem-1",
		TargetID: "mem-2",
		Type:     "invalid_type",
	}, &buf)
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "invalid relation type")
}

func TestHandlers_AddRelation_MissingIDs(t *testing.T) {
	server, _ := setupTestServerWithGraph(t)

	var buf mcp.BufferWriter
	err := server.AddRelation(context.Background(), mcp.AddRelationInput{
		SourceID: "",
		TargetID: "mem-2",
		Type:     "related_to",
	}, &buf)
	assert.Error(t, err)
}

func TestHandlers_GetRelations(t *testing.T) {
	server, memRepo := setupTestServerWithGraph(t)
	ctx := context.Background()

	mem1 := &memory.Memory{Type: "note", Title: "Memory 1", Content: "Content 1", Status: "active"}
	mem2 := &memory.Memory{Type: "note", Title: "Memory 2", Content: "Content 2", Status: "active"}
	require.NoError(t, memRepo.Create(ctx, mem1))
	require.NoError(t, memRepo.Create(ctx, mem2))

	var buf mcp.BufferWriter
	err := server.AddRelation(ctx, mcp.AddRelationInput{
		SourceID: mem1.ID, TargetID: mem2.ID, Type: "extends", Strength: 0.9,
	}, &buf)
	require.NoError(t, err)

	buf.Reset()
	err = server.GetRelations(ctx, mcp.GetRelationsInput{MemoryID: mem1.ID}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	assert.Equal(t, float64(1), response["count"])

	relations := response["relations"].([]interface{})
	rel := relations[0].(map[string]interface{})
	assert.Equal(t, "extends", rel["type"])
}

func TestHandlers_GetRelations_Empty(t *testing.T) {
	server, memRepo := setupTestServerWithGraph(t)
	ctx := context.Background()

	mem := &memory.Memory{Type: "note", Title: "Solo", Content: "No relations", Status: "active"}
	require.NoError(t, memRepo.Create(ctx, mem))

	var buf mcp.BufferWriter
	err := server.GetRelations(ctx, mcp.GetRelationsInput{MemoryID: mem.ID}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	assert.Equal(t, float64(0), response["count"])
}

func TestHandlers_DeleteRelation(t *testing.T) {
	server, memRepo := setupTestServerWithGraph(t)
	ctx := context.Background()

	mem1 := &memory.Memory{Type: "note", Title: "Memory 1", Content: "Content 1", Status: "active"}
	mem2 := &memory.Memory{Type: "note", Title: "Memory 2", Content: "Content 2", Status: "active"}
	require.NoError(t, memRepo.Create(ctx, mem1))
	require.NoError(t, memRepo.Create(ctx, mem2))

	var buf mcp.BufferWriter
	err := server.AddRelation(ctx, mcp.AddRelationInput{
		SourceID: mem1.ID, TargetID: mem2.ID, Type: "depends_on",
	}, &buf)
	require.NoError(t, err)

	buf.Reset()
	err = server.DeleteRelation(ctx, mcp.DeleteRelationInput{
		SourceID: mem1.ID, TargetID: mem2.ID,
	}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	assert.Equal(t, true, response["success"])

	buf.Reset()
	err = server.GetRelations(ctx, mcp.GetRelationsInput{MemoryID: mem1.ID}, &buf)
	require.NoError(t, err)

	var getResp map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &getResp))
	assert.Equal(t, float64(0), getResp["count"])
}

func TestHandlers_FindPath(t *testing.T) {
	server, memRepo := setupTestServerWithGraph(t)
	ctx := context.Background()

	mem1 := &memory.Memory{Type: "note", Title: "A", Content: "Node A", Status: "active"}
	mem2 := &memory.Memory{Type: "note", Title: "B", Content: "Node B", Status: "active"}
	mem3 := &memory.Memory{Type: "note", Title: "C", Content: "Node C", Status: "active"}
	require.NoError(t, memRepo.Create(ctx, mem1))
	require.NoError(t, memRepo.Create(ctx, mem2))
	require.NoError(t, memRepo.Create(ctx, mem3))

	var buf mcp.BufferWriter
	require.NoError(t, server.AddRelation(ctx, mcp.AddRelationInput{SourceID: mem1.ID, TargetID: mem2.ID, Type: "related_to"}, &buf))
	buf.Reset()
	require.NoError(t, server.AddRelation(ctx, mcp.AddRelationInput{SourceID: mem2.ID, TargetID: mem3.ID, Type: "related_to"}, &buf))

	buf.Reset()
	err := server.FindPath(ctx, mcp.FindPathInput{FromID: mem1.ID, ToID: mem3.ID}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	assert.Equal(t, true, response["found"])
	path := response["path"].([]interface{})
	assert.Len(t, path, 3)
}

func TestHandlers_FindPath_NotFound(t *testing.T) {
	server, _ := setupTestServerWithGraph(t)

	var buf mcp.BufferWriter
	err := server.FindPath(context.Background(), mcp.FindPathInput{
		FromID: "nonexistent-1", ToID: "nonexistent-2",
	}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	assert.Equal(t, false, response["found"])
}

func TestHandlers_AddMemory_InvalidType(t *testing.T) {
	server, _ := setupTestServer(t)

	var buf mcp.BufferWriter
	err := server.AddMemoryWithWriter(context.Background(), mcp.AddMemoryInput{
		Type:  "invalid_type",
		Title: "Bad Type",
	}, &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid type")
}

func TestHandlers_ActivateContext_WithDedup(t *testing.T) {
	server, memRepo := setupTestServerWithGraph(t)
	ctx := context.Background()

	mem := &memory.Memory{Type: "note", Title: "Dedup Test", Content: "Testing deduplication", Status: "active"}
	require.NoError(t, memRepo.Create(ctx, mem))

	var buf mcp.BufferWriter
	err := server.ActivateContext(ctx, mcp.ActivateContextInput{
		Query:     "dedup test",
		SessionID: "dedup-session",
		MaxTokens: 4000,
	}, &buf)
	if err != nil {
		assert.Contains(t, err.Error(), "embedding")
		return
	}

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	assert.Equal(t, false, response["duplicate_detected"])
}

func TestHandlers_ActivateContext_LowSimilarity(t *testing.T) {
	server, memRepo := setupTestServerWithGraph(t)
	ctx := context.Background()

	mem := &memory.Memory{Type: "note", Title: "Random Note", Content: "Something completely unrelated", Status: "active"}
	require.NoError(t, memRepo.Create(ctx, mem))

	var buf mcp.BufferWriter
	err := server.ActivateContext(ctx, mcp.ActivateContextInput{
		Query:     "quantum physics entanglement",
		SessionID: "low-sim-session",
	}, &buf)
	if err != nil {
		return
	}

	assert.Contains(t, buf.String(), "Low similarity")
}
