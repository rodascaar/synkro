package memory

import "time"

type Memory struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Source    *string   `json:"source"`
	Status    string    `json:"status"`
	Tags      []string  `json:"tags"`
	TopicKey  string    `json:"topic_key,omitempty"`
	Pinned    bool      `json:"pinned"`
}

type HybridSearchFilter struct {
	Type   string
	Status string
	Limit  int
}

// MemoryFilter es un alias de HybridSearchFilter; ambos representan el mismo
// filtro de tipo/estado/límite.
type MemoryFilter = HybridSearchFilter

type MemoryUpdate struct {
	Title   *string
	Content *string
	Status  *string
	Tags    []string
}

type MemoryRelation struct {
	SourceID  string    `json:"source_id"`
	TargetID  string    `json:"target_id"`
	Type      string    `json:"type"`     // extends, depends_on, conflicts_with, example_of, part_of, related_to
	Strength  float64   `json:"strength"` // 0.0 a 1.0
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RelationType string

const (
	RelationExtends       RelationType = "extends"
	RelationDependsOn     RelationType = "depends_on"
	RelationConflictsWith RelationType = "conflicts_with"
	RelationExampleOf     RelationType = "example_of"
	RelationPartOf        RelationType = "part_of"
	RelationRelatedTo     RelationType = "related_to"
)

type HybridSearchResult struct {
	Memory        *Memory `json:"memory"`
	VectorScore   float64 `json:"vector_score"`   // Puntaje vectorial (0.0-1.0)
	FTS5Score     float64 `json:"fts5_score"`     // Puntaje FTS5 (BM25 normalizado 0.0-1.0)
	CombinedScore float64 `json:"combined_score"` // Puntaje combinado (0.0-1.0)
	MatchType     string  `json:"match_type"`     // "vector", "fts5", "both"
}

type FTS5Result struct {
	Memory *Memory
	Rank   float64 // BM25 rank de FTS5 (negativo; menor = mejor)
	Score  float64 // Score normalizado (0.0-1.0)
}

type VectorResult struct {
	Memory *Memory
	Score  float64 // Coseno (0.0-1.0)
}