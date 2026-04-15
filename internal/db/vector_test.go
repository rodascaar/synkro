package db

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	return db
}

func TestVectorFunctions_Unavailable(t *testing.T) {
	orig := vecAvailable
	vecAvailable = false
	defer func() { vecAvailable = orig }()

	db := openTestDB(t)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	err := InsertVector(ctx, db, "mem-1", []float32{0.1, 0.2})
	assert.Error(t, err)

	err = DeleteVector(ctx, db, "mem-1")
	assert.Error(t, err)

	err = UpdateVector(ctx, db, "mem-1", []float32{0.1, 0.2})
	assert.Error(t, err)

	results, err := SearchVectors(ctx, db, []float32{0.1, 0.2}, 5)
	assert.Nil(t, results)
	assert.Error(t, err)

	results, err = SearchVectorsWithMetadata(ctx, db, []float32{0.1, 0.2}, 5)
	assert.Nil(t, results)
	assert.Error(t, err)
}

func TestVectorFunctions_Available(t *testing.T) {
	if !vecAvailable {
		t.Skip("sqlite-vec not available in this environment")
	}

	orig := vecAvailable
	vecAvailable = true
	defer func() { vecAvailable = orig }()

	db := openTestDB(t)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `CREATE VIRTUAL TABLE IF NOT EXISTS memory_vec USING vec0(memory_id TEXT PRIMARY KEY, embedding float[2])`)
	if err != nil {
		t.Skipf("vec0 not available: %v", err)
	}

	err = InsertVector(ctx, db, "mem-1", []float32{0.1, 0.2})
	assert.NoError(t, err)
}
