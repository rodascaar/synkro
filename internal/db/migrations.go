package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
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
			Version: 1,
			Name:    "initial_schema",
			Up: func(ctx context.Context, db Executor) error {
				return nil
			},
		},
		{
			Version: 2,
			Name:    "memory_tags_table",
			Up: func(ctx context.Context, db Executor) error {
				_, err := db.ExecContext(ctx, `
					CREATE TABLE IF NOT EXISTS memory_tags (
						memory_id TEXT NOT NULL,
						tag TEXT NOT NULL,
						PRIMARY KEY (memory_id, tag),
						FOREIGN KEY (memory_id) REFERENCES memories(id) ON DELETE CASCADE
					);
					CREATE INDEX IF NOT EXISTS idx_memory_tags_tag ON memory_tags(tag);
				`)
				if err != nil {
					return fmt.Errorf("failed to create memory_tags table: %w", err)
				}

				rows, err := db.QueryContext(ctx, `SELECT id, tags FROM memories WHERE tags IS NOT NULL AND tags != ''`)
				if err != nil {
					return fmt.Errorf("failed to query memories for tag migration: %w", err)
				}
				defer rows.Close()

				type memTags struct {
					id   string
					tags string
				}
				var toMigrate []memTags
				for rows.Next() {
					var mt memTags
					if err := rows.Scan(&mt.id, &mt.tags); err != nil {
						continue
					}
					toMigrate = append(toMigrate, mt)
				}

				for _, mt := range toMigrate {
					tags := splitTags(mt.tags)
					for _, tag := range tags {
						if tag == "" {
							continue
						}
						_, err := db.ExecContext(ctx,
							`INSERT OR IGNORE INTO memory_tags (memory_id, tag) VALUES (?, ?)`,
							mt.id, tag,
						)
						if err != nil {
							log.Printf("Warning: failed to migrate tag %q for memory %s: %v", tag, mt.id, err)
						}
					}
				}

				return nil
			},
		},
		{
			Version: 3,
			Name:    "vec0_vectors_table",
			Up: func(ctx context.Context, db Executor) error {
				_, err := db.ExecContext(ctx, `
					CREATE VIRTUAL TABLE IF NOT EXISTS memory_vec USING vec0(
						memory_id TEXT PRIMARY KEY,
						embedding float[384]
					);
				`)
				if err != nil {
					log.Printf("Warning: vec0 not available, falling back to in-memory vector search: %v", err)
					return nil
				}

				rows, err := db.QueryContext(ctx, `
					SELECT me.memory_id, me.embedding
					FROM memory_embeddings me
				`)
				if err != nil {
					if err == sql.ErrNoRows {
						return nil
					}
					return fmt.Errorf("failed to query existing embeddings: %w", err)
				}
				defer rows.Close()

				type existingEmbedding struct {
					memoryID  string
					embedding []byte
				}
				var toMigrate []existingEmbedding
				for rows.Next() {
					var e existingEmbedding
					if err := rows.Scan(&e.memoryID, &e.embedding); err != nil {
						continue
					}
					toMigrate = append(toMigrate, e)
				}

				for _, e := range toMigrate {
					if e.embedding == nil {
						continue
					}
					_, err := db.ExecContext(ctx,
						`INSERT OR IGNORE INTO memory_vec(memory_id, embedding) VALUES (?, ?)`,
						e.memoryID, e.embedding,
					)
					if err != nil {
						log.Printf("Warning: failed to migrate vector for memory %s: %v", e.memoryID, err)
					}
				}

				return nil
			},
		},
	}
}

func splitTags(tagsStr string) []string {
	if tagsStr == "" {
		return nil
	}
	parts := strings.Split(tagsStr, "|")
	var tags []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tags = append(tags, p)
		}
	}
	return tags
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
			tx.Rollback()
			return fmt.Errorf("migration %d (%s) failed: %w", m.Version, m.Name, err)
		}

		_, err = tx.ExecContext(ctx, `INSERT INTO _migrations (version, name) VALUES (?, ?)`, m.Version, m.Name)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", m.Version, err)
		}

		log.Printf("Migration %d: %s applied successfully", m.Version, m.Name)
	}

	return nil
}
