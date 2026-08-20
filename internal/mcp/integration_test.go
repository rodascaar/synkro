package mcp_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rodascaar/synkro/internal/db"
	"github.com/rodascaar/synkro/internal/embeddings"
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

func TestHandlers_SearchMemories_FiltersLowSimilarity(t *testing.T) {
	server, memRepo := setupTestServer(t)
	memRepo.SetEmbeddingGenerator(embeddings.NewTFIDFEmbeddingGenerator(nil))

	ctx := context.Background()
	_ = memRepo.Create(ctx, &memory.Memory{
		Type: "note", Title: "Database Design", Content: "PostgreSQL architecture and indexing patterns", Status: "active",
	})
	_ = memRepo.Create(ctx, &memory.Memory{
		Type: "note", Title: "Cooking Recipe", Content: "How to bake a chocolate cake with frosting", Status: "active",
	})

	var buf mcp.BufferWriter
	err := server.SearchMemory(context.Background(), mcp.SearchMemoryInput{Query: "database architecture", Limit: 10}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))

	memories := response["memories"].([]interface{})
	assert.GreaterOrEqual(t, len(memories), 1)

	for _, m := range memories {
		title := m.(map[string]interface{})["title"].(string)
		assert.NotEqual(t, "Cooking Recipe", title, "unrelated low-similarity memory should be filtered")
	}
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
	memRepo.SetEmbeddingGenerator(embeddings.NewTFIDFEmbeddingGenerator(nil))
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

func TestHandlers_SearchMemories_DefaultActive(t *testing.T) {
	server, memRepo := setupTestServer(t)
	ctx := context.Background()

	_ = memRepo.Create(ctx, &memory.Memory{
		Type: "note", Title: "Database Active", Content: "database indexing", Status: "active",
	})
	_ = memRepo.Create(ctx, &memory.Memory{
		Type: "note", Title: "Database Archived", Content: "database backup", Status: "archived",
	})

	var buf mcp.BufferWriter
	err := server.SearchMemory(context.Background(), mcp.SearchMemoryInput{Query: "database", Limit: 10}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))

	memories := response["memories"].([]interface{})
	assert.Len(t, memories, 1)
	assert.Equal(t, "Database Active", memories[0].(map[string]interface{})["title"])
}

func TestHandlers_ActivateContext_CountsTokens(t *testing.T) {
	server, memRepo := setupTestServerWithGraph(t)
	memRepo.SetEmbeddingGenerator(embeddings.NewTFIDFEmbeddingGenerator(nil))
	ctx := context.Background()

	mem := &memory.Memory{Type: "note", Title: "Token Counting", Content: "This is a test content with several words to count", Status: "active"}
	require.NoError(t, memRepo.Create(ctx, mem))

	var buf mcp.BufferWriter
	err := server.ActivateContext(ctx, mcp.ActivateContextInput{
		Query:     "token counting words",
		SessionID: "token-session",
	}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))

	totalTokens, ok := response["total_tokens"].(float64)
	require.True(t, ok)
	assert.Greater(t, totalTokens, float64(0))
}

func TestHandlers_DetectConflicts(t *testing.T) {
	server, memRepo := setupTestServer(t)
	memRepo.SetEmbeddingGenerator(embeddings.NewTFIDFEmbeddingGenerator(nil))
	ctx := context.Background()

	require.NoError(t, memRepo.Create(ctx, &memory.Memory{
		Type: "decision", Title: "Use SQLite", Content: "We decided to use SQLite as the database for the project", Status: "active",
	}))

	var buf mcp.BufferWriter
	err := server.DetectConflicts(ctx, mcp.DetectConflictsInput{Text: "Use SQLite We decided to use SQLite as the database for the project"}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	assert.Equal(t, true, response["conflict_detected"])
	conflicts := response["potential_conflicts"].([]interface{})
	assert.NotEmpty(t, conflicts)
	assert.Equal(t, "Use SQLite", conflicts[0].(map[string]interface{})["title"])
}

func TestHandlers_DetectConflicts_NoMatch(t *testing.T) {
	server, memRepo := setupTestServer(t)
	memRepo.SetEmbeddingGenerator(embeddings.NewTFIDFEmbeddingGenerator(nil))
	ctx := context.Background()

	require.NoError(t, memRepo.Create(ctx, &memory.Memory{
		Type: "note", Title: "Coffee", Content: "I prefer espresso in the morning", Status: "active",
	}))

	var buf mcp.BufferWriter
	err := server.DetectConflicts(ctx, mcp.DetectConflictsInput{Text: "The cat sat on the mat under a tree"}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	assert.Equal(t, false, response["conflict_detected"])
	assert.Equal(t, float64(0), response["count"])
}

func TestHandlers_AddMemory_AttachesConflicts(t *testing.T) {
	server, memRepo := setupTestServer(t)
	memRepo.SetEmbeddingGenerator(embeddings.NewTFIDFEmbeddingGenerator(nil))
	ctx := context.Background()

	require.NoError(t, memRepo.Create(ctx, &memory.Memory{
		Type: "decision", Title: "DB Choice", Content: "We chose PostgreSQL for the primary storage", Status: "active",
	}))

	var buf mcp.BufferWriter
	err := server.AddMemoryWithWriter(ctx, mcp.AddMemoryInput{
		Type:    "decision",
		Title:   "DB Choice Revisited",
		Content: "We chose PostgreSQL for the primary storage solution",
	}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	require.Equal(t, true, response["success"])

	conflictDetected, ok := response["conflict_detected"].(bool)
	require.True(t, ok)
	// El umbral por defecto 0.8 debería detectar el casi-duplicado.
	assert.True(t, conflictDetected)
}

func TestHandlers_JudgeConflict_NotConflict(t *testing.T) {
	server, memRepo := setupTestServerWithGraph(t)
	ctx := context.Background()

	mem1 := &memory.Memory{Type: "decision", Title: "SQLite DB", Content: "Use SQLite for local storage", Status: "active"}
	mem2 := &memory.Memory{Type: "decision", Title: "Postgres DB", Content: "Use PostgreSQL for remote storage", Status: "active"}
	require.NoError(t, memRepo.Create(ctx, mem1))
	require.NoError(t, memRepo.Create(ctx, mem2))

	var buf mcp.BufferWriter
	err := server.JudgeConflict(ctx, mcp.JudgeConflictInput{
		MemoryID:    mem1.ID,
		CandidateID: mem2.ID,
		Verdict:     "not_conflict",
		Reasoning:   "Different scopes",
	}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	assert.Equal(t, false, response["relation_created"])
}

func TestHandlers_JudgeConflict_Supersedes(t *testing.T) {
	server, memRepo := setupTestServerWithGraph(t)
	ctx := context.Background()

	mem1 := &memory.Memory{Type: "decision", Title: "Old Approach", Content: "Use v1 architecture", Status: "active"}
	mem2 := &memory.Memory{Type: "decision", Title: "New Approach", Content: "Use v2 architecture", Status: "active"}
	require.NoError(t, memRepo.Create(ctx, mem1))
	require.NoError(t, memRepo.Create(ctx, mem2))

	var buf mcp.BufferWriter
	err := server.JudgeConflict(ctx, mcp.JudgeConflictInput{
		MemoryID:    mem2.ID,
		CandidateID: mem1.ID,
		Verdict:     "supersedes",
		Reasoning:   "v2 replaces v1",
	}, &buf)
	require.NoError(t, err)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &response))
	assert.Equal(t, true, response["relation_created"])
	assert.Equal(t, "supersedes", response["relation_type"])

	buf.Reset()
	err = server.GetRelations(ctx, mcp.GetRelationsInput{MemoryID: mem2.ID}, &buf)
	require.NoError(t, err)
	var relations map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(buf.String()), &relations))
	rels := relations["relations"].([]interface{})
	assert.Len(t, rels, 1)
	assert.Equal(t, "supersedes", rels[0].(map[string]interface{})["type"])
}

func TestHandlers_JudgeConflict_InvalidVerdict(t *testing.T) {
	server, _ := setupTestServerWithGraph(t)
	ctx := context.Background()

	var buf mcp.BufferWriter
	err := server.JudgeConflict(ctx, mcp.JudgeConflictInput{
		MemoryID:    "mem-1",
		CandidateID: "mem-2",
		Verdict:     "invalid_verdict",
	}, &buf)
	require.Error(t, err)
}
