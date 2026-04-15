package graph

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rodascaar/synkro/internal/memory"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Add(ctx context.Context, rel *memory.MemoryRelation) error {
	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO memory_relations (source_id, target_id, type, strength, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(source_id, target_id, type) DO UPDATE SET
			strength = excluded.strength,
			updated_at = excluded.updated_at
	`, rel.SourceID, rel.TargetID, rel.Type, rel.Strength, now.Format(time.RFC3339), now.Format(time.RFC3339))

	if err != nil {
		return fmt.Errorf("failed to add relation: %w", err)
	}

	return nil
}

func (r *Repository) Get(ctx context.Context, memoryID string) ([]*memory.MemoryRelation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT source_id, target_id, type, strength, created_at, updated_at
		FROM memory_relations
		WHERE source_id = ? OR target_id = ?
		ORDER BY strength DESC
	`, memoryID, memoryID)

	if err != nil {
		return nil, fmt.Errorf("failed to get relations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var relations []*memory.MemoryRelation
	for rows.Next() {
		var rel memory.MemoryRelation
		var createdAt, updatedAt string

		if err := rows.Scan(&rel.SourceID, &rel.TargetID, &rel.Type, &rel.Strength, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		rel.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		rel.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		relations = append(relations, &rel)
	}

	return relations, nil
}

func (r *Repository) Delete(ctx context.Context, sourceID, targetID string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM memory_relations
		WHERE source_id = ? AND target_id = ?
	`, sourceID, targetID)

	if err != nil {
		return fmt.Errorf("failed to delete relation: %w", err)
	}

	return nil
}

func (r *Repository) UpdateStrength(ctx context.Context, rel *memory.MemoryRelation) error {
	rel.UpdatedAt = time.Now().UTC()

	_, err := r.db.ExecContext(ctx, `
		UPDATE memory_relations
		SET strength = ?, updated_at = ?
		WHERE source_id = ? AND target_id = ? AND type = ?
	`, rel.Strength, rel.UpdatedAt.Format(time.RFC3339), rel.SourceID, rel.TargetID, rel.Type)

	if err != nil {
		return fmt.Errorf("failed to update relation strength: %w", err)
	}

	return nil
}

func (r *Repository) LoadAll(ctx context.Context) ([]*memory.MemoryRelation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT source_id, target_id, type, strength, created_at, updated_at
		FROM memory_relations
		ORDER BY created_at DESC
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to load all relations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var relations []*memory.MemoryRelation
	for rows.Next() {
		var rel memory.MemoryRelation
		var createdAt, updatedAt string

		if err := rows.Scan(&rel.SourceID, &rel.TargetID, &rel.Type, &rel.Strength, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		rel.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		rel.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		relations = append(relations, &rel)
	}

	return relations, nil
}
