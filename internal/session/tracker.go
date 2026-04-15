package session

import (
	"context"
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
		return
	}
	defer rows.Close()

	for rows.Next() {
		var sessionID string
		if err := rows.Scan(&sessionID); err != nil {
			continue
		}

		session, err := st.repo.Get(ctx, sessionID)
		if err == nil && session != nil {
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
		_ = st.repo.Save(context.Background(), session)
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
			_ = st.repo.Save(context.Background(), session)
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

func (st *SessionTracker) Persist(ctx context.Context) error {
	st.mu.Lock()
	sessions := make([]*Session, 0, len(st.sessions))
	for _, s := range st.sessions {
		sessions = append(sessions, s)
	}
	st.mu.Unlock()

	if st.repo == nil {
		return nil
	}

	for _, session := range sessions {
		if err := st.repo.Save(ctx, session); err != nil {
			return err
		}
	}

	return nil
}
