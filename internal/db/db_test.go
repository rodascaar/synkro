package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"

	database, err := New(tmpFile)
	require.NoError(t, err)
	require.NotNil(t, database)
	defer database.Close()

	db := database.DB()
	require.NotNil(t, db)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM memories").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	var mode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&mode)
	require.NoError(t, err)
	assert.Equal(t, "wal", mode)

	var fk int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fk)
	require.NoError(t, err)
	assert.Equal(t, 1, fk)
}

func TestNew_DefaultPath(t *testing.T) {
	database, err := New("memory.db")
	if err != nil {
		assert.NoError(t, err)
		return
	}
	require.NotNil(t, database)
	database.Close()
}

func TestNew_EmptyPath(t *testing.T) {
	database, err := New("")
	require.NoError(t, err)
	require.NotNil(t, database)
	database.Close()
}

func TestNew_Subdirectory(t *testing.T) {
	tmpDir := t.TempDir()
	path := tmpDir + "/sub/dir/test.db"

	database, err := New(path)
	require.NoError(t, err)
	require.NotNil(t, database)
	database.Close()
}

func TestDB(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"

	database, err := New(tmpFile)
	require.NoError(t, err)
	defer database.Close()

	db := database.DB()
	assert.NotNil(t, db)

	err = db.Ping()
	assert.NoError(t, err)
}

func TestClose(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"

	database, err := New(tmpFile)
	require.NoError(t, err)

	err = database.Close()
	assert.NoError(t, err)
}

func TestClose_NilDB(t *testing.T) {
	database := &Database{}
	err := database.Close()
	assert.NoError(t, err)
}

func TestInitSchema_Idempotent(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"

	database, err := New(tmpFile)
	require.NoError(t, err)

	err = database.initSchema()
	assert.NoError(t, err)

	err = database.initSchema()
	assert.NoError(t, err)

	database.Close()
}

func TestInitSchema_TablesCreated(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"

	database, err := New(tmpFile)
	require.NoError(t, err)
	defer database.Close()

	tables := []string{
		"memories",
		"memory_embeddings",
		"memory_relations",
		"sessions",
		"session_memories",
		"embedding_cache",
	}

	for _, table := range tables {
		var count int
		err := database.DB().QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&count)
		require.NoError(t, err, "table %s should exist", table)
		assert.Equal(t, 1, count, "table %s should exist", table)
	}
}
