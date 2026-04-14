package graph

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/rodascaar/synkro/internal/db"
	"github.com/rodascaar/synkro/internal/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGraphTestDB(t *testing.T) (*db.Database, *memory.Repository, *Repository) {
	t.Helper()
	tmpFile := t.TempDir() + "/test.db"
	database, err := db.New(tmpFile)
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })

	memRepo := memory.NewRepository(database.DB())
	graphRepo := NewRepository(database.DB())

	return database, memRepo, graphRepo
}

func insertMemory(t *testing.T, db *sql.DB, id string) {
	t.Helper()
	db.Exec(`INSERT OR IGNORE INTO memories (id, created_at, updated_at, type, title, content, status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339), "note", "test", "test content", "active")
}

func TestRepository_AddAndGet(t *testing.T) {
	database, _, graphRepo := setupGraphTestDB(t)
	ctx := context.Background()

	insertMemory(t, database.DB(), "mem1")
	insertMemory(t, database.DB(), "mem2")

	rel := &memory.MemoryRelation{
		SourceID:  "mem1",
		TargetID:  "mem2",
		Type:      "related_to",
		Strength:  0.8,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := graphRepo.Add(ctx, rel)
	require.NoError(t, err)

	relations, err := graphRepo.Get(ctx, "mem1")
	require.NoError(t, err)
	require.Len(t, relations, 1)
	assert.Equal(t, "mem2", relations[0].TargetID)
	assert.Equal(t, "related_to", relations[0].Type)
	assert.Equal(t, 0.8, relations[0].Strength)
}

func TestRepository_Get_BothDirections(t *testing.T) {
	database, _, graphRepo := setupGraphTestDB(t)
	ctx := context.Background()

	insertMemory(t, database.DB(), "a")
	insertMemory(t, database.DB(), "b")

	rel := &memory.MemoryRelation{
		SourceID: "a", TargetID: "b", Type: "depends_on",
		Strength: 0.5, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	graphRepo.Add(ctx, rel)

	fromA, _ := graphRepo.Get(ctx, "a")
	assert.Len(t, fromA, 1)

	fromB, _ := graphRepo.Get(ctx, "b")
	assert.Len(t, fromB, 1)
}

func TestRepository_Add_Upsert(t *testing.T) {
	database, _, graphRepo := setupGraphTestDB(t)
	ctx := context.Background()

	insertMemory(t, database.DB(), "a")
	insertMemory(t, database.DB(), "b")

	rel := &memory.MemoryRelation{
		SourceID: "a", TargetID: "b", Type: "related_to",
		Strength: 0.5, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	graphRepo.Add(ctx, rel)

	rel.Strength = 0.9
	graphRepo.Add(ctx, rel)

	relations, _ := graphRepo.Get(ctx, "a")
	assert.Equal(t, 0.9, relations[0].Strength)
}

func TestRepository_Delete(t *testing.T) {
	database, _, graphRepo := setupGraphTestDB(t)
	ctx := context.Background()

	insertMemory(t, database.DB(), "a")
	insertMemory(t, database.DB(), "b")

	rel := &memory.MemoryRelation{
		SourceID: "a", TargetID: "b", Type: "related_to",
		Strength: 0.5, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	graphRepo.Add(ctx, rel)

	err := graphRepo.Delete(ctx, "a", "b")
	require.NoError(t, err)

	relations, _ := graphRepo.Get(ctx, "a")
	assert.Empty(t, relations)
}

func TestRepository_UpdateStrength(t *testing.T) {
	database, _, graphRepo := setupGraphTestDB(t)
	ctx := context.Background()

	insertMemory(t, database.DB(), "a")
	insertMemory(t, database.DB(), "b")

	rel := &memory.MemoryRelation{
		SourceID: "a", TargetID: "b", Type: "related_to",
		Strength: 0.5, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	graphRepo.Add(ctx, rel)

	rel.Strength = 1.0
	err := graphRepo.UpdateStrength(ctx, rel)
	require.NoError(t, err)

	relations, _ := graphRepo.Get(ctx, "a")
	assert.Equal(t, 1.0, relations[0].Strength)
}

func TestRepository_LoadAll(t *testing.T) {
	database, _, graphRepo := setupGraphTestDB(t)
	ctx := context.Background()

	insertMemory(t, database.DB(), "a")
	insertMemory(t, database.DB(), "b")
	insertMemory(t, database.DB(), "c")

	graphRepo.Add(ctx, &memory.MemoryRelation{
		SourceID: "a", TargetID: "b", Type: "related_to",
		Strength: 0.5, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	graphRepo.Add(ctx, &memory.MemoryRelation{
		SourceID: "b", TargetID: "c", Type: "depends_on",
		Strength: 0.8, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})

	all, err := graphRepo.LoadAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestGraph_AddRelation(t *testing.T) {
	database, memRepo, graphRepo := setupGraphTestDB(t)
	ctx := context.Background()

	insertMemory(t, database.DB(), "a")
	insertMemory(t, database.DB(), "b")

	g := NewGraph(memRepo, graphRepo)

	rel := &memory.MemoryRelation{
		SourceID: "a", TargetID: "b", Type: "related_to",
		Strength: 0.7, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}

	err := g.AddRelation(ctx, rel)
	require.NoError(t, err)

	relations, err := g.GetRelations(ctx, "a")
	require.NoError(t, err)
	assert.Len(t, relations, 1)
}

func TestGraph_FindPath_Direct(t *testing.T) {
	database, memRepo, graphRepo := setupGraphTestDB(t)
	ctx := context.Background()

	insertMemory(t, database.DB(), "a")
	insertMemory(t, database.DB(), "b")

	g := NewGraph(memRepo, graphRepo)
	g.AddRelation(ctx, &memory.MemoryRelation{
		SourceID: "a", TargetID: "b", Type: "related_to",
		Strength: 0.5, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})

	path, err := g.FindPath(ctx, "a", "b")
	require.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestGraph_FindPath_SameNode(t *testing.T) {
	_, memRepo, graphRepo := setupGraphTestDB(t)
	ctx := context.Background()

	g := NewGraph(memRepo, graphRepo)

	path, err := g.FindPath(ctx, "a", "a")
	require.NoError(t, err)
	assert.Equal(t, []string{"a"}, path)
}

func TestGraph_FindPath_NoPath(t *testing.T) {
	_, memRepo, graphRepo := setupGraphTestDB(t)
	ctx := context.Background()

	g := NewGraph(memRepo, graphRepo)

	_, err := g.FindPath(ctx, "a", "z")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no path found")
}

func TestGraph_GetStats(t *testing.T) {
	database, memRepo, graphRepo := setupGraphTestDB(t)
	ctx := context.Background()

	insertMemory(t, database.DB(), "a")
	insertMemory(t, database.DB(), "b")
	insertMemory(t, database.DB(), "c")

	g := NewGraph(memRepo, graphRepo)

	g.AddRelation(ctx, &memory.MemoryRelation{
		SourceID: "a", TargetID: "b", Type: "related_to",
		Strength: 0.5, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	g.AddRelation(ctx, &memory.MemoryRelation{
		SourceID: "b", TargetID: "c", Type: "related_to",
		Strength: 0.5, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})

	stats, err := g.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, stats["related_to"])
}

func TestGraph_NewGraph_NilRepo(t *testing.T) {
	g := NewGraph(nil, nil)
	require.NotNil(t, g)

	ctx := context.Background()
	relations, err := g.GetRelations(ctx, "anything")
	require.NoError(t, err)
	assert.Empty(t, relations)
}

func TestGraph_AddRelation_NilRepo(t *testing.T) {
	g := NewGraph(nil, nil)
	ctx := context.Background()

	rel := &memory.MemoryRelation{
		SourceID: "a", TargetID: "b", Type: "related_to",
		Strength: 0.5, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}

	err := g.AddRelation(ctx, rel)
	assert.NoError(t, err)
}
