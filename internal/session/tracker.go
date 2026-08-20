package session

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"
)

type Session struct {
	ID                string
	LastQuery         string
	LastQueryAt       time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeliveredMemories map[string]*DeliveredMemory
}

type DeliveredMemory struct {
	MemoryID    string
	DeliveredAt time.Time
}

type SessionTracker struct {
	mu         sync.RWMutex
	sessions   map[string]*Session
	repo       *Repository
	maxHistory int
}

func NewSessionTracker(repo *Repository) *SessionTracker {
	st := &SessionTracker{
		sessions:   make(map[string]*Session),
		repo:       repo,
		maxHistory: 20,
	}

	if repo != nil {
		st.loadFromDB(context.Background())
	}

	return st
}

func (st *SessionTracker) loadFromDB(ctx context.Context) {
	rows, err := st.repo.db.QueryContext(ctx, "SELECT id FROM sessions")
	if err != nil {
		log.Printf("warning: failed to load sessions from DB: %v", err)
		return
	}

	// Collect all IDs first and close the rows before issuing per-session
	// queries. Loading sessions while rows are still open would deadlock on a
	// single-connection pool (database/sql reuses the held connection).
	var ids []string
	for rows.Next() {
		var sessionID string
		if err := rows.Scan(&sessionID); err != nil {
			log.Printf("warning: failed to scan session row: %v", err)
			continue
		}
		ids = append(ids, sessionID)
	}
	if err := rows.Close(); err != nil {
		log.Printf("warning: failed to close sessions rows: %v", err)
	}
	if err := rows.Err(); err != nil {
		log.Printf("warning: error reading sessions: %v", err)
		return
	}

	for _, sessionID := range ids {
		session, err := st.repo.Get(ctx, sessionID)
		if err != nil {
			log.Printf("warning: failed to load session %q from DB: %v", sessionID, err)
			continue
		}
		if session != nil {
			st.mu.Lock()
			st.sessions[sessionID] = session
			st.mu.Unlock()
		}
	}
}

func (st *SessionTracker) getOrCreateLocked(sessionID string) *Session {
	session, exists := st.sessions[sessionID]
	if !exists {
		now := time.Now()
		session = &Session{
			ID:                sessionID,
			CreatedAt:         now,
			UpdatedAt:         now,
			DeliveredMemories: make(map[string]*DeliveredMemory),
		}
		st.sessions[sessionID] = session
	}
	return session
}

func (st *SessionTracker) GetOrCreate(ctx context.Context, sessionID string) *Session {
	st.mu.Lock()
	session := st.getOrCreateLocked(sessionID)
	st.mu.Unlock()

	if st.repo != nil {
		_ = st.repo.Save(ctx, session)
	}

	return session
}

func (st *SessionTracker) MarkAsDelivered(ctx context.Context, sessionID, memoryID string) {
	st.mu.Lock()
	_, existed := st.sessions[sessionID]
	session := st.getOrCreateLocked(sessionID)
	session.DeliveredMemories[memoryID] = &DeliveredMemory{
		MemoryID:    memoryID,
		DeliveredAt: time.Now(),
	}
	session.UpdatedAt = time.Now()
	st.mu.Unlock()

	if st.repo != nil {
		if !existed {
			_ = st.repo.Save(ctx, session)
		}
		_ = st.repo.MarkDelivered(ctx, sessionID, memoryID)
	}
}

func (st *SessionTracker) GetRecentDeliveries(_ context.Context, sessionID string, limit int) []string {
	st.mu.RLock()
	defer st.mu.RUnlock()

	session, exists := st.sessions[sessionID]
	if !exists {
		return []string{}
	}

	deliveries := make([]*DeliveredMemory, 0, len(session.DeliveredMemories))
	for _, d := range session.DeliveredMemories {
		deliveries = append(deliveries, d)
	}

	sort.Slice(deliveries, func(i, j int) bool {
		return deliveries[i].DeliveredAt.After(deliveries[j].DeliveredAt)
	})

	result := make([]string, 0, limit)
	for i, d := range deliveries {
		if i >= limit {
			break
		}
		result = append(result, d.MemoryID)
	}

	return result
}

func (st *SessionTracker) UpdateLastQuery(ctx context.Context, sessionID, query string) {
	st.mu.Lock()
	session := st.getOrCreateLocked(sessionID)
	session.LastQuery = query
	session.LastQueryAt = time.Now()
	session.UpdatedAt = time.Now()
	st.mu.Unlock()

	if st.repo != nil {
		_ = st.repo.UpdateLastQuery(ctx, sessionID, query)
	}
}

func (st *SessionTracker) IsDuplicateQuery(sessionID, query string) bool {
	st.mu.RLock()
	defer st.mu.RUnlock()

	session, exists := st.sessions[sessionID]
	if !exists {
		return false
	}

	return session.LastQuery == query
}
