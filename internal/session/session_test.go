package session

import (
	"context"
	"testing"
	"time"

	"github.com/rodascaar/synkro/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func now() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

func setupTestDB(t *testing.T) *db.Database {
	t.Helper()
	tmpFile := t.TempDir() + "/test.db"
	database, err := db.New(tmpFile)
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })
	return database
}

func TestRepository_SaveAndGet(t *testing.T) {
	database := setupTestDB(t)
	repo := NewRepository(database.DB())
	ctx := context.Background()

	session := &Session{
		ID:        "test-session-1",
		LastQuery: "hello world",
		CreatedAt: now(),
		UpdatedAt: now(),
		DeliveredMemories: map[string]*DeliveredMemory{
			"mem1": {MemoryID: "mem1", DeliveredAt: now()},
		},
	}

	err := repo.Save(ctx, session)
	require.NoError(t, err)

	got, err := repo.Get(ctx, "test-session-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "test-session-1", got.ID)
	assert.Equal(t, "hello world", got.LastQuery)
}

func TestRepository_Get_NotFound(t *testing.T) {
	database := setupTestDB(t)
	repo := NewRepository(database.DB())

	got, err := repo.Get(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestRepository_UpdateLastQuery(t *testing.T) {
	database := setupTestDB(t)
	repo := NewRepository(database.DB())
	ctx := context.Background()

	session := &Session{ID: "sess1", CreatedAt: now(), UpdatedAt: now()}
	repo.Save(ctx, session)

	err := repo.UpdateLastQuery(ctx, "sess1", "new query")
	require.NoError(t, err)

	got, _ := repo.Get(ctx, "sess1")
	assert.Equal(t, "new query", got.LastQuery)
}

func TestRepository_MarkDelivered(t *testing.T) {
	database := setupTestDB(t)
	repo := NewRepository(database.DB())
	ctx := context.Background()

	session := &Session{ID: "sess1", CreatedAt: now(), UpdatedAt: now()}
	repo.Save(ctx, session)

	database.DB().Exec(`INSERT OR IGNORE INTO memories (id, created_at, updated_at, type, title, content, status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"mem1", now().Format(time.RFC3339Nano), now().Format(time.RFC3339Nano), "note", "test", "test", "active")

	err := repo.MarkDelivered(ctx, "sess1", "mem1")
	require.NoError(t, err)

	database.DB().Exec(`INSERT OR IGNORE INTO memories (id, created_at, updated_at, type, title, content, status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"mem2", now().Format(time.RFC3339Nano), now().Format(time.RFC3339Nano), "note", "test", "test", "active")

	err = repo.MarkDelivered(ctx, "sess1", "mem2")
	require.NoError(t, err)

	deliveries, err := repo.GetRecentDeliveries(ctx, "sess1", 10)
	require.NoError(t, err)
	assert.Len(t, deliveries, 2)
}

func TestRepository_GetRecentDeliveries_Limit(t *testing.T) {
	database := setupTestDB(t)
	repo := NewRepository(database.DB())
	ctx := context.Background()

	session := &Session{ID: "sess1", CreatedAt: now(), UpdatedAt: now()}
	repo.Save(ctx, session)

	for i := 0; i < 5; i++ {
		memID := "mem" + string(rune('a'+i))
		database.DB().Exec(`INSERT OR IGNORE INTO memories (id, created_at, updated_at, type, title, content, status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			memID, now().Format(time.RFC3339Nano), now().Format(time.RFC3339Nano), "note", "test", "test", "active")
		repo.MarkDelivered(ctx, "sess1", memID)
	}

	deliveries, err := repo.GetRecentDeliveries(ctx, "sess1", 3)
	require.NoError(t, err)
	assert.Len(t, deliveries, 3)
}

func TestRepository_Save_Upsert(t *testing.T) {
	database := setupTestDB(t)
	repo := NewRepository(database.DB())
	ctx := context.Background()

	session := &Session{ID: "sess1", LastQuery: "first", CreatedAt: now(), UpdatedAt: now()}
	repo.Save(ctx, session)

	session.LastQuery = "second"
	repo.Save(ctx, session)

	got, _ := repo.Get(ctx, "sess1")
	assert.Equal(t, "second", got.LastQuery)
}

func TestSessionTracker_InMemory(t *testing.T) {
	tracker := NewSessionTracker(nil)
	ctx := context.Background()

	session := tracker.GetOrCreate(ctx, "s1")
	assert.Equal(t, "s1", session.ID)
	assert.Equal(t, "", session.LastQuery)

	session.LastQuery = "query1"
	session.LastQueryAt = time.Now()
	assert.True(t, tracker.IsDuplicateQuery("s1", "query1"))
	assert.False(t, tracker.IsDuplicateQuery("s1", "query2"))
	assert.False(t, tracker.IsDuplicateQuery("s2", "query1"))

	session.DeliveredMemories["mem1"] = &DeliveredMemory{MemoryID: "mem1", DeliveredAt: time.Now()}
	session.DeliveredMemories["mem2"] = &DeliveredMemory{MemoryID: "mem2", DeliveredAt: time.Now()}

	deliveries := tracker.GetRecentDeliveries(ctx, "s1", 10)
	assert.Len(t, deliveries, 2)

	deliveries = tracker.GetRecentDeliveries(ctx, "nonexistent", 10)
	assert.Empty(t, deliveries)
}

func TestSessionTracker_WithDB(t *testing.T) {
	database := setupTestDB(t)
	repo := NewRepository(database.DB())
	tracker := NewSessionTracker(repo)
	ctx := context.Background()

	session := tracker.GetOrCreate(ctx, "s1")
	session.LastQuery = "test query"
	session.LastQueryAt = time.Now()
	session.DeliveredMemories["mem1"] = &DeliveredMemory{MemoryID: "mem1", DeliveredAt: time.Now()}

	err := tracker.Persist(ctx)
	require.NoError(t, err)

	tracker2 := NewSessionTracker(repo)
	got := tracker2.GetOrCreate(ctx, "s1")
	assert.Equal(t, "test query", got.LastQuery)
}

func TestSessionTracker_GetOrCreate_Idempotent(t *testing.T) {
	tracker := NewSessionTracker(nil)
	ctx := context.Background()

	s1 := tracker.GetOrCreate(ctx, "s1")
	s2 := tracker.GetOrCreate(ctx, "s1")
	assert.Same(t, s1, s2)
}

func TestSessionTracker_GetRecentDeliveries_Order(t *testing.T) {
	tracker := NewSessionTracker(nil)
	ctx := context.Background()

	session := tracker.GetOrCreate(ctx, "order-test")
	base := time.Now().UTC()

	session.DeliveredMemories["first"] = &DeliveredMemory{
		MemoryID:    "first",
		DeliveredAt: base.Add(-3 * time.Minute),
	}
	session.DeliveredMemories["second"] = &DeliveredMemory{
		MemoryID:    "second",
		DeliveredAt: base.Add(-1 * time.Minute),
	}
	session.DeliveredMemories["third"] = &DeliveredMemory{
		MemoryID:    "third",
		DeliveredAt: base.Add(-5 * time.Minute),
	}
	session.DeliveredMemories["fourth"] = &DeliveredMemory{
		MemoryID:    "fourth",
		DeliveredAt: base.Add(-2 * time.Second),
	}

	deliveries := tracker.GetRecentDeliveries(ctx, "order-test", 10)
	require.Len(t, deliveries, 4)
	assert.Equal(t, "fourth", deliveries[0], "most recent should be first")
	assert.Equal(t, "second", deliveries[1], "second most recent")
	assert.Equal(t, "first", deliveries[2], "third most recent")
	assert.Equal(t, "third", deliveries[3], "oldest should be last")

	deliveries = tracker.GetRecentDeliveries(ctx, "order-test", 2)
	require.Len(t, deliveries, 2)
	assert.Equal(t, "fourth", deliveries[0])
	assert.Equal(t, "second", deliveries[1])
}
