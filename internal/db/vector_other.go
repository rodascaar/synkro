//go:build !windows

package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/rodascaar/synkro/internal/embeddings"
	synkroerrors "github.com/rodascaar/synkro/internal/errors"
)

var vecAvailable bool

func init() {
	defer func() {
		if r := recover(); r != nil {
			vecAvailable = false
			log.Printf("Warning: sqlite-vec not available, vector search will use in-memory fallback: %v", r)
		}
	}()
	sqlite_vec.Auto()
	vecAvailable = true
}

func InsertVector(ctx context.Context, db *sql.DB, memoryID string, embedding []float32) error {
	if !vecAvailable {
		return synkroerrors.ErrVecNotAvailable
	}

	vec, err := sqlite_vec.SerializeFloat32(embedding)
	if err != nil {
		return fmt.Errorf("failed to serialize embedding: %w", err)
	}

	_, err = db.ExecContext(ctx, `INSERT INTO memory_vec(memory_id, embedding) VALUES (?, ?)`, memoryID, vec)
	if err != nil {
		return fmt.Errorf("failed to insert vector: %w", err)
	}

	return nil
}

func DeleteVector(ctx context.Context, db *sql.DB, memoryID string) error {
	if !vecAvailable {
		return synkroerrors.ErrVecNotAvailable
	}

	_, err := db.ExecContext(ctx, `DELETE FROM memory_vec WHERE memory_id = ?`, memoryID)
	return err
}

func UpdateVector(ctx context.Context, exec Executor, memoryID string, embedding []float32) error {
	if !vecAvailable {
		return synkroerrors.ErrVecNotAvailable
	}

	vec, err := sqlite_vec.SerializeFloat32(embedding)
	if err != nil {
		return fmt.Errorf("failed to serialize embedding: %w", err)
	}

	_, err = exec.ExecContext(ctx, `
		INSERT INTO memory_vec(memory_id, embedding) VALUES (?, ?)
		ON CONFLICT(memory_id) DO UPDATE SET embedding = excluded.embedding
	`, memoryID, vec)
	return err
}

type VectorSearchResult struct {
	MemoryID  string
	Distance  float64
	Embedding []float32
}

func SearchVectors(ctx context.Context, db *sql.DB, queryEmbedding []float32, k int) ([]*VectorSearchResult, error) {
	if !vecAvailable {
		return nil, synkroerrors.ErrVecNotAvailable
	}

	vec, err := sqlite_vec.SerializeFloat32(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query embedding: %w", err)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT memory_id, distance
		FROM memory_vec
		WHERE embedding MATCH ?
		ORDER BY distance
		LIMIT ?
	`, vec, k)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []*VectorSearchResult
	for rows.Next() {
		var r VectorSearchResult
		if err := rows.Scan(&r.MemoryID, &r.Distance); err != nil {
			continue
		}
		results = append(results, &r)
	}

	return results, nil
}

func SearchVectorsWithMetadata(ctx context.Context, db *sql.DB, queryEmbedding []float32, k int) ([]*VectorSearchResult, error) {
	if !vecAvailable {
		return nil, synkroerrors.ErrVecNotAvailable
	}

	vec, err := sqlite_vec.SerializeFloat32(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query embedding: %w", err)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT v.memory_id, v.distance, e.embedding
		FROM memory_vec v
		LEFT JOIN memory_embeddings e ON v.memory_id = e.memory_id
		WHERE v.embedding MATCH ?
		ORDER BY v.distance
		LIMIT ?
	`, vec, k)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []*VectorSearchResult
	for rows.Next() {
		var r VectorSearchResult
		var embeddingBytes []byte
		if err := rows.Scan(&r.MemoryID, &r.Distance, &embeddingBytes); err != nil {
			continue
		}
		if embeddingBytes != nil {
			r.Embedding = embeddings.DeserializeEmbedding(embeddingBytes)
		}
		results = append(results, &r)
	}

	return results, nil
}
