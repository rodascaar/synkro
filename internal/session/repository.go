package session

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(ctx context.Context, session *Session) error {
	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (id, last_query, last_query_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			last_query = excluded.last_query,
			last_query_at = excluded.last_query_at,
			updated_at = excluded.updated_at
	`, session.ID, session.LastQuery, session.LastQueryAt.Format(time.RFC3339), session.CreatedAt.Format(time.RFC3339), now.Format(time.RFC3339))

	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

func (r *Repository) Get(ctx context.Context, sessionID string) (*Session, error) {
	var lastQuery, lastQueryAt, createdAt, updatedAt sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT last_query, last_query_at, created_at, updated_at
		FROM sessions WHERE id = ?
	`, sessionID).Scan(&lastQuery, &lastQueryAt, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	createdTime, _ := time.Parse(time.RFC3339, createdAt.String)
	updatedTime, _ := time.Parse(time.RFC3339, updatedAt.String)

	session := &Session{
		ID:                sessionID,
		CreatedAt:         createdTime,
		UpdatedAt:         updatedTime,
		DeliveredMemories: make(map[string]*DeliveredMemory),
	}

	if lastQuery.Valid {
		session.LastQuery = lastQuery.String
	}
	if lastQueryAt.Valid {
		session.LastQueryAt, _ = time.Parse(time.RFC3339, lastQueryAt.String)
	}

	deliveries, err := r.getDeliveries(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries: %w", err)
	}

	for _, delivery := range deliveries {
		session.DeliveredMemories[delivery.MemoryID] = delivery
	}

	return session, nil
}

func (r *Repository) UpdateLastQuery(ctx context.Context, sessionID, query string) error {
	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx, `
		UPDATE sessions 
		SET last_query = ?, last_query_at = ?, updated_at = ?
		WHERE id = ?
	`, query, now.Format(time.RFC3339), now.Format(time.RFC3339), sessionID)

	if err != nil {
		return fmt.Errorf("failed to update last query: %w", err)
	}

	return nil
}

func (r *Repository) MarkDelivered(ctx context.Context, sessionID, memoryID string) error {
	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO session_memories (session_id, memory_id, delivered_at)
		VALUES (?, ?, ?)
		ON CONFLICT(session_id, memory_id) DO UPDATE SET
			delivered_at = excluded.delivered_at
	`, sessionID, memoryID, now.Format(time.RFC3339))

	if err != nil {
		return fmt.Errorf("failed to mark delivered: %w", err)
	}

	return nil
}

func (r *Repository) GetRecentDeliveries(ctx context.Context, sessionID string, limit int) ([]*DeliveredMemory, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT memory_id, delivered_at
		FROM session_memories
		WHERE session_id = ?
		ORDER BY delivered_at DESC
		LIMIT ?
	`, sessionID, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to get recent deliveries: %w", err)
	}
	defer func() { _ = rows.Close() }()

	deliveries := make([]*DeliveredMemory, 0)
	for rows.Next() {
		var memoryID, deliveredAt string
		if err := rows.Scan(&memoryID, &deliveredAt); err != nil {
			return nil, err
		}

		deliveryTime, _ := time.Parse(time.RFC3339, deliveredAt)
		deliveries = append(deliveries, &DeliveredMemory{
			MemoryID:    memoryID,
			DeliveredAt: deliveryTime,
		})
	}

	return deliveries, nil
}

func (r *Repository) getDeliveries(ctx context.Context, sessionID string) ([]*DeliveredMemory, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT memory_id, delivered_at
		FROM session_memories
		WHERE session_id = ?
		ORDER BY delivered_at DESC
	`, sessionID)

	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries: %w", err)
	}
	defer func() { _ = rows.Close() }()

	deliveries := make([]*DeliveredMemory, 0)
	for rows.Next() {
		var memoryID, deliveredAt string
		if err := rows.Scan(&memoryID, &deliveredAt); err != nil {
			return nil, err
		}

		deliveryTime, _ := time.Parse(time.RFC3339, deliveredAt)
		deliveries = append(deliveries, &DeliveredMemory{
			MemoryID:    memoryID,
			DeliveredAt: deliveryTime,
		})
	}

	return deliveries, nil
}
