package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Migration struct {
	Version int
	Name    string
	Up      func(ctx context.Context, db Executor) error
}

func getMigrations() []Migration {
	return []Migration{
		{
			Version: 4,
			Name:    "fts5_create_and_backfill",
			Up: func(ctx context.Context, db Executor) error {
				_, err := db.ExecContext(ctx, `
					CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
						id,
						title,
						content,
						content=memories,
						content_rowid=rowid
					);
				`)
				if err != nil {
					log.Printf("Warning: FTS5 not available, skipping migration: %v", err)
					return nil
				}

				_, err = db.ExecContext(ctx, `
					INSERT INTO memories_fts(rowid, id, title, content)
					SELECT rowid, id, title, content FROM memories;
				`)
				if err != nil {
					log.Printf("Warning: failed to backfill FTS5: %v", err)
				}

				_, err = db.ExecContext(ctx, `
					CREATE TRIGGER IF NOT EXISTS memories_ai AFTER INSERT ON memories BEGIN
						INSERT INTO memories_fts(rowid, id, title, content)
						VALUES (new.rowid, new.id, new.title, new.content);
					END;

					CREATE TRIGGER IF NOT EXISTS memories_ad AFTER DELETE ON memories BEGIN
						INSERT INTO memories_fts(memories_fts, rowid, id, title, content)
						VALUES ('delete', old.rowid, old.id, old.title, old.content);
					END;

					CREATE TRIGGER IF NOT EXISTS memories_au AFTER UPDATE ON memories BEGIN
						INSERT INTO memories_fts(memories_fts, rowid, id, title, content)
						VALUES ('delete', old.rowid, old.id, old.title, old.content);
						INSERT INTO memories_fts(rowid, id, title, content)
						VALUES (new.rowid, new.id, new.title, new.content);
					END;
				`)
				if err != nil {
					log.Printf("Warning: failed to create FTS5 triggers: %v", err)
				}

				return nil
			},
		},
		{
			Version: 5,
			Name:    "dedupe_embeddings_and_drop_memory_tags",
			Up: func(ctx context.Context, db Executor) error {
				_, err := db.ExecContext(ctx, `
					DELETE FROM memory_embeddings
					WHERE id NOT IN (SELECT MAX(id) FROM memory_embeddings GROUP BY memory_id);
				`)
				if err != nil {
					return fmt.Errorf("failed to dedupe embeddings: %w", err)
				}

				_, err = db.ExecContext(ctx, `
					DROP INDEX IF EXISTS idx_memory_embeddings_memory;
					CREATE UNIQUE INDEX IF NOT EXISTS idx_memory_embeddings_memory ON memory_embeddings(memory_id);
				`)
				if err != nil {
					return fmt.Errorf("failed to make memory_embeddings.memory_id unique: %w", err)
				}

				_, err = db.ExecContext(ctx, `DROP TABLE IF EXISTS memory_tags`)
				if err != nil {
					return fmt.Errorf("failed to drop memory_tags table: %w", err)
				}
				return nil
			},
		},
	}
}

func (d *Database) runMigrations() error {
	ctx := context.Background()

	_, err := d.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS _migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	var currentVersion int
	err = d.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM _migrations`).Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	migrations := getMigrations()
	for _, m := range migrations {
		if m.Version <= currentVersion {
			continue
		}

		log.Printf("Applying migration %d: %s", m.Version, m.Name)

		tx, err := d.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %d: %w", m.Version, err)
		}

		if err := m.Up(ctx, tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %d (%s) failed: %w", m.Version, m.Name, err)
		}

		_, err = tx.ExecContext(ctx, `INSERT INTO _migrations (version, name) VALUES (?, ?)`, m.Version, m.Name)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", m.Version, err)
		}

		log.Printf("Migration %d: %s applied successfully", m.Version, m.Name)
	}

	return nil
}
