package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed extensions/*
var extensionAssets embed.FS

type Database struct {
	db *sql.DB
}

func New(path string) (*Database, error) {
	if path == "" {
		path = "memory.db"
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	database := &Database{db: db}

	if err := database.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return database, nil
}

func (d *Database) DB() *sql.DB {
	return d.db
}

func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

func (d *Database) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS memories (
		id TEXT PRIMARY KEY,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		type TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT,
		source TEXT,
		status TEXT NOT NULL DEFAULT 'active',
		tags TEXT
	);

	CREATE TABLE IF NOT EXISTS memory_embeddings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		memory_id TEXT NOT NULL,
		embedding BLOB NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY (memory_id) REFERENCES memories(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS memory_relations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_id TEXT NOT NULL,
		target_id TEXT NOT NULL,
		type TEXT NOT NULL,
		strength REAL NOT NULL DEFAULT 1.0,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		FOREIGN KEY (source_id) REFERENCES memories(id) ON DELETE CASCADE,
		FOREIGN KEY (target_id) REFERENCES memories(id) ON DELETE CASCADE,
		UNIQUE(source_id, target_id, type)
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		last_query TEXT,
		last_query_at TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS session_memories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		memory_id TEXT NOT NULL,
		delivered_at TEXT NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
		FOREIGN KEY (memory_id) REFERENCES memories(id) ON DELETE CASCADE,
		UNIQUE(session_id, memory_id)
	);

	CREATE TABLE IF NOT EXISTS embedding_cache (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		text_hash TEXT NOT NULL UNIQUE,
		text TEXT NOT NULL,
		embedding BLOB NOT NULL,
		model_type TEXT NOT NULL,
		created_at TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type);
	CREATE INDEX IF NOT EXISTS idx_memories_status ON memories(status);
	CREATE INDEX IF NOT EXISTS idx_memories_created ON memories(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_memory_embeddings_memory ON memory_embeddings(memory_id);
	CREATE INDEX IF NOT EXISTS idx_memory_relations_source ON memory_relations(source_id);
	CREATE INDEX IF NOT EXISTS idx_memory_relations_target ON memory_relations(target_id);
	CREATE INDEX IF NOT EXISTS idx_memory_relations_type ON memory_relations(type);
	CREATE INDEX IF NOT EXISTS idx_sessions_id ON sessions(id);
	CREATE INDEX IF NOT EXISTS idx_session_memories_session ON session_memories(session_id);
	CREATE INDEX IF NOT EXISTS idx_session_memories_delivered ON session_memories(delivered_at DESC);
	CREATE INDEX IF NOT EXISTS idx_embedding_cache_hash ON embedding_cache(text_hash);
	`

	_, err := d.db.Exec(schema)
	if err != nil {
		return err
	}

	ftsSchema := `
	CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
		id,
		title,
		content,
		content=memories,
		content_rowid=rowid
	);

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
	`

	_, err = d.db.Exec(ftsSchema)
	return err
}
