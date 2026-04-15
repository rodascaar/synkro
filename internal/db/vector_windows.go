//go:build windows

package db

import (
	"context"
	"database/sql"

	synkroerrors "github.com/rodascaar/synkro/internal/errors"
)

var vecAvailable bool

func InsertVector(ctx context.Context, db *sql.DB, memoryID string, embedding []float32) error {
	return synkroerrors.ErrVecNotAvailable
}

func DeleteVector(ctx context.Context, db *sql.DB, memoryID string) error {
	return synkroerrors.ErrVecNotAvailable
}

func UpdateVector(ctx context.Context, exec Executor, memoryID string, embedding []float32) error {
	return synkroerrors.ErrVecNotAvailable
}

type VectorSearchResult struct {
	MemoryID  string
	Distance  float64
	Embedding []float32
}

func SearchVectors(ctx context.Context, db *sql.DB, queryEmbedding []float32, k int) ([]*VectorSearchResult, error) {
	return nil, synkroerrors.ErrVecNotAvailable
}

func SearchVectorsWithMetadata(ctx context.Context, db *sql.DB, queryEmbedding []float32, k int) ([]*VectorSearchResult, error) {
	return nil, synkroerrors.ErrVecNotAvailable
}
