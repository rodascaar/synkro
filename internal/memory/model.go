package memory

import (
	"time"
)

type Memory struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Type      string
	Title     string
	Content   string
	Source    *string
	Status    string
	Tags      []string

	// V2: Engrama (Embedding vectorial)
	Embedding []float32 `json:"-"`
}

type MemoryFilter struct {
	Type     string
	Status   string
	Source   string
	Tags     []string
	FromTime time.Time
	ToTime   time.Time
	Limit    int
	Offset   int
}

type MemoryUpdate struct {
	Title     *string
	Content   *string
	Status    *string
	Tags      []string
	Embedding *[]float32 `json:"embedding,omitempty"`
}

// =====================================================
// V2: MemoryRelation - Grafo de Asociaciones
// =====================================================

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

// =====================================================
// V2: EmbeddedMemory - Resultado de búsqueda vectorial
// =====================================================

type EmbeddedMemory struct {
	Memory     *Memory `json:"memory"`
	Similarity float64 `json:"similarity"` // 0.0 a 1.0
	Distance   float64 `json:"distance"`   // Distancia coseno (0.0 = idéntico, 2.0 = opuesto)
}

// =====================================================
// V2: Hybrid Search - Búsqueda combinada FTS5 + Vectorial
// =====================================================

type HybridSearchResult struct {
	Memory        *Memory `json:"memory"`
	VectorScore   float64 `json:"vector_score"`   // Puntaje vectorial (0.0-1.0)
	FTS5Score     float64 `json:"fts5_score"`     // Puntaje FTS5 (BM25)
	CombinedScore float64 `json:"combined_score"` // Puntaje combinado (0.0-1.0)
	MatchType     string  `json:"match_type"`     // "vectorial", "fts5", "both"
}

type FTS5Result struct {
	Memory *Memory
	Rank   float64 // BM25 rank from FTS5
	Score  float64 // Normalized score (0.0-1.0)
}

type VectorResult struct {
	Memory *Memory
	Score  float64 // Cosine similarity (0.0-1.0)
}

type HybridSearchFilter struct {
	Type          string
	Status        string
	Source        string
	Tags          []string
	Limit         int
	MinSimilarity *float64 `json:"min_similarity,omitempty"`
}

// =====================================================
// V2: Contexto Enriquecido (con grafo)
// =====================================================

type EnrichedContext struct {
	Memory          *Memory           `json:"memory"`
	RelatedMemories []*Memory         `json:"related_memories"`
	Relations       []*MemoryRelation `json:"relations"`
	Similarity      float64           `json:"similarity"`
}

type ContextWindow struct {
	PrimaryMemory   *Memory            `json:"primary_memory"`
	RelatedMemories []*EnrichedContext `json:"related_memories"`
	TotalTokens     int                `json:"total_tokens"`
}
