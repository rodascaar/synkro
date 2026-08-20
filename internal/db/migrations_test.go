package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errModuleMissing = errors.New("no such module: fts5")
	errCreateFTS     = errors.New("SQL logic error: syntax error near \"VIRTUAL\"")
	errBackfill      = errors.New("no such table: memories_fts")
	errTrigger       = errors.New("SQL logic error: syntax error near \"TRIGGER\"")
)

type fakeExecutor struct {
	execFunc func(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (f *fakeExecutor) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.execFunc(ctx, query, args...)
}

func (f *fakeExecutor) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	panic("unexpected QueryContext call")
}

func (f *fakeExecutor) QueryRowContext(context.Context, string, ...any) *sql.Row {
	panic("unexpected QueryRowContext call")
}

// scriptedExec returns an ExecContext that returns each error in errs, in
// order, and panics if called more times than scripted.
func scriptedExec(errs ...error) func(ctx context.Context, query string, args ...any) (sql.Result, error) {
	i := 0
	return func(ctx context.Context, query string, args ...any) (sql.Result, error) {
		if i >= len(errs) {
			panic(fmt.Sprintf("unexpected ExecContext call #%d: %s", i+1, query))
		}
		e := errs[i]
		i++
		return nil, e
	}
}

func migrationV4(t *testing.T) Migration {
	t.Helper()
	for _, m := range getMigrations() {
		if m.Version == 4 {
			return m
		}
	}
	t.Fatalf("migration 4 not found")
	return Migration{}
}

func TestMigrationV4_Success(t *testing.T) {
	f := &fakeExecutor{execFunc: scriptedExec(nil, nil, nil)}
	err := migrationV4(t).Up(context.Background(), f)
	require.NoError(t, err)
}

func TestMigrationV4_CreateModuleMissing_Degrades(t *testing.T) {
	f := &fakeExecutor{execFunc: scriptedExec(errModuleMissing)}
	err := migrationV4(t).Up(context.Background(), f)
	require.NoError(t, err)
}

func TestMigrationV4_CreateFailsOther_ReturnsError(t *testing.T) {
	f := &fakeExecutor{execFunc: scriptedExec(errCreateFTS)}
	err := migrationV4(t).Up(context.Background(), f)
	require.Error(t, err)
	assert.ErrorIs(t, err, errCreateFTS)
	assert.Contains(t, err.Error(), "failed to create FTS5 virtual table")
}

func TestMigrationV4_BackfillFails_ReturnsError(t *testing.T) {
	f := &fakeExecutor{execFunc: scriptedExec(nil, errBackfill)}
	err := migrationV4(t).Up(context.Background(), f)
	require.Error(t, err)
	assert.ErrorIs(t, err, errBackfill)
	assert.Contains(t, err.Error(), "failed to backfill FTS5")
}

func TestMigrationV4_TriggerFails_ReturnsError(t *testing.T) {
	f := &fakeExecutor{execFunc: scriptedExec(nil, nil, errTrigger)}
	err := migrationV4(t).Up(context.Background(), f)
	require.Error(t, err)
	assert.ErrorIs(t, err, errTrigger)
	assert.Contains(t, err.Error(), "failed to create FTS5 triggers")
}

func TestMigrationV4_CreatesFTS5InRealDB(t *testing.T) {
	database, err := New(t.TempDir() + "/test.db")
	require.NoError(t, err)
	defer func() { _ = database.Close() }()

	var count int
	err = database.DB().QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='memories_fts'",
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	var triggers int
	err = database.DB().QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='trigger' AND name IN ('memories_ai','memories_ad','memories_au')",
	).Scan(&triggers)
	require.NoError(t, err)
	assert.Equal(t, 3, triggers)
}

func TestIsFTS5ModuleMissing(t *testing.T) {
	assert.True(t, isFTS5ModuleMissing(errors.New("no such module: fts5")))
	assert.True(t, isFTS5ModuleMissing(errors.New("SQL logic error: no such module: fts5 (1)")))
	assert.False(t, isFTS5ModuleMissing(nil))
	assert.False(t, isFTS5ModuleMissing(errCreateFTS))
	assert.False(t, isFTS5ModuleMissing(errBackfill))
}
