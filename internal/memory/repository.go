package memory

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rodascaar/synkro/internal/embeddings"
)

type Repository struct {
	db                 *sql.DB
	embeddingGenerator embeddings.EmbeddingGenerator
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SetEmbeddingGenerator(generator embeddings.EmbeddingGenerator) {
	r.embeddingGenerator = generator
}

func (r *Repository) Create(ctx context.Context, mem *Memory) error {
	id := mem.ID
	if id == "" {
		id = fmt.Sprintf("mem-%s-%04d", time.Now().UTC().Format("20060102150405"), uuid.New().ID()%9000+1000)
	}

	now := time.Now().UTC()
	if mem.CreatedAt.IsZero() {
		mem.CreatedAt = now
	}
	if mem.UpdatedAt.IsZero() {
		mem.UpdatedAt = now
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO memories (id, created_at, updated_at, type, title, content, source, status, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, mem.CreatedAt.Format(time.RFC3339), mem.UpdatedAt.Format(time.RFC3339), mem.Type, mem.Title, mem.Content, mem.Source, mem.Status, strings.Join(mem.Tags, "|"))
	if err != nil {
		return err
	}

	mem.ID = id

	if err := tx.Commit(); err != nil {
		return err
	}

	if r.embeddingGenerator != nil {
		embedding, err := r.embeddingGenerator.Generate(ctx, mem.Title+" "+mem.Content)
		if err == nil {
			_ = r.saveEmbedding(ctx, mem.ID, embedding)
		}
	}

	return nil
}

func (r *Repository) saveEmbedding(ctx context.Context, memoryID string, embedding []float32) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO memory_embeddings (memory_id, embedding, created_at)
		VALUES (?, ?, ?)
	`, memoryID, embedding, time.Now().UTC().Format(time.RFC3339))
	return err
}

func (r *Repository) Get(ctx context.Context, id string) (*Memory, error) {
	mem := &Memory{}
	var tagsStr string
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, created_at, updated_at, type, title, content, source, status, tags
		FROM memories WHERE id = ?
	`, id).Scan(&mem.ID, &createdAtStr, &updatedAtStr, &mem.Type, &mem.Title, &mem.Content, &mem.Source, &mem.Status, &tagsStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	mem.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	mem.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	if tagsStr != "" {
		mem.Tags = strings.Split(tagsStr, "|")
	}

	return mem, nil
}

func (r *Repository) Search(ctx context.Context, query string, filter MemoryFilter) ([]*Memory, error) {
	if r.embeddingGenerator != nil && query != "" {
		queryEmbedding, err := r.embeddingGenerator.Generate(ctx, query)
		if err == nil {
			return r.vectorialSearch(ctx, queryEmbedding, filter)
		}
	}

	where := "WHERE 1=1"
	args := []interface{}{}

	if filter.Type != "" {
		where += " AND m.type = ?"
		args = append(args, filter.Type)
	}
	if filter.Status != "" {
		where += " AND m.status = ?"
		args = append(args, filter.Status)
	}

	sqlQuery := fmt.Sprintf("SELECT m.id, m.created_at, m.updated_at, m.type, m.title, m.content, m.source, m.status, m.tags FROM memories m %s ORDER BY m.created_at DESC", where)
	if filter.Limit > 0 {
		sqlQuery += " LIMIT ?"
		args = append(args, filter.Limit)
	} else {
		sqlQuery += " LIMIT 50"
		args = append(args, 50)
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		mem := &Memory{}
		var tagsStr string
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(&mem.ID, &createdAtStr, &updatedAtStr, &mem.Type, &mem.Title, &mem.Content, &mem.Source, &mem.Status, &tagsStr); err != nil {
			return nil, err
		}

		mem.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		mem.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		if tagsStr != "" {
			mem.Tags = strings.Split(tagsStr, "|")
		}
		memories = append(memories, mem)
	}

	return memories, nil
}

func (r *Repository) vectorialSearch(ctx context.Context, queryEmbedding []float32, filter MemoryFilter) ([]*Memory, error) {
	where := "WHERE 1=1"
	args := []interface{}{}

	if filter.Type != "" {
		where += " AND m.type = ?"
		args = append(args, filter.Type)
	}
	if filter.Status != "" {
		where += " AND m.status = ?"
		args = append(args, filter.Status)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	sqlQuery := fmt.Sprintf(`
		SELECT m.id, m.created_at, m.updated_at, m.type, m.title, m.content, m.source, m.status, m.tags
		FROM memories m
		LEFT JOIN memory_embeddings e ON m.id = e.memory_id
		%s
		ORDER BY m.created_at DESC
		LIMIT ?
	`, where)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type resultWithEmbedding struct {
		Memory    *Memory
		Embedding []float32
	}

	var results []resultWithEmbedding
	for rows.Next() {
		mem := &Memory{}
		var tagsStr string
		var createdAtStr, updatedAtStr string
		var embeddingBytes []byte

		if err := rows.Scan(&mem.ID, &createdAtStr, &updatedAtStr, &mem.Type, &mem.Title, &mem.Content, &mem.Source, &mem.Status, &tagsStr, &embeddingBytes); err != nil {
			continue
		}

		mem.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		mem.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		if tagsStr != "" {
			mem.Tags = strings.Split(tagsStr, "|")
		}

		var embedding []float32
		if embeddingBytes != nil {
			embedding = embeddings.DeserializeEmbedding(embeddingBytes)
		}

		results = append(results, resultWithEmbedding{
			Memory:    mem,
			Embedding: embedding,
		})
	}

	type scoredResult struct {
		Memory     *Memory
		Similarity float64
	}

	var scored []scoredResult
	for _, result := range results {
		if result.Embedding != nil {
			similarity := float64(embeddings.CosineSimilarity(queryEmbedding, result.Embedding))
			scored = append(scored, scoredResult{
				Memory:     result.Memory,
				Similarity: similarity,
			})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Similarity > scored[j].Similarity
	})

	var memories []*Memory
	for i, result := range scored {
		if i >= limit {
			break
		}
		memories = append(memories, result.Memory)
	}

	return memories, nil
}

func (r *Repository) HybridSearch(ctx context.Context, query string, k int, filter HybridSearchFilter) ([]*HybridSearchResult, error) {
	if r.embeddingGenerator == nil {
		return nil, fmt.Errorf("embedding generator not configured")
	}

	queryEmbedding, err := r.embeddingGenerator.Generate(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	fts5Results, fts5Err := r.searchWithBM25(ctx, query, filter)

	vectorResults, err := r.searchByVector(ctx, queryEmbedding, filter, k*3)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	var combinedResults []*HybridSearchResult
	if fts5Err == nil && len(fts5Results) > 0 {
		combinedResults = r.mergeSearchResults(ctx, fts5Results, vectorResults, queryEmbedding, k)
	} else {
		combinedResults = r.vectorOnlyResults(vectorResults, k)
	}

	return combinedResults, nil
}

func (r *Repository) vectorOnlyResults(vectorResults map[string]*VectorResult, k int) []*HybridSearchResult {
	results := make([]*HybridSearchResult, 0, len(vectorResults))

	for _, vector := range vectorResults {
		results = append(results, &HybridSearchResult{
			Memory:        vector.Memory,
			VectorScore:   vector.Score,
			FTS5Score:     0.0,
			CombinedScore: vector.Score,
			MatchType:     "vector",
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CombinedScore > results[j].CombinedScore
	})

	if len(results) > k {
		results = results[:k]
	}

	return results
}

func (r *Repository) searchWithBM25(ctx context.Context, query string, filter HybridSearchFilter) (map[string]*FTS5Result, error) {
	where := "WHERE 1=1"
	args := []interface{}{}

	if filter.Type != "" {
		where += " AND m.type = ?"
		args = append(args, filter.Type)
	}
	if filter.Status != "" {
		where += " AND m.status = ?"
		args = append(args, filter.Status)
	}

	sqlQuery := fmt.Sprintf(`
		SELECT m.id, m.created_at, m.updated_at, m.type, m.title, m.content, m.source, m.status, m.tags, rank
		FROM memories m
		INNER JOIN memories_fts f ON m.id = f.id
		%s AND memories_fts MATCH ?
		ORDER BY rank
		LIMIT 50
	`, where)
	args = append(args, query)

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]*FTS5Result)
	for rows.Next() {
		mem := &Memory{}
		var tagsStr string
		var createdAtStr, updatedAtStr string
		var bm25Rank float64

		if err := rows.Scan(&mem.ID, &createdAtStr, &updatedAtStr, &mem.Type, &mem.Title, &mem.Content, &mem.Source, &mem.Status, &tagsStr, &bm25Rank); err != nil {
			return nil, err
		}

		mem.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		mem.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		if tagsStr != "" {
			mem.Tags = strings.Split(tagsStr, "|")
		}

		results[mem.ID] = &FTS5Result{
			Memory: mem,
			Rank:   bm25Rank,
			Score:  calculateBM25Score(bm25Rank),
		}
	}

	return results, nil
}

func (r *Repository) searchByVector(ctx context.Context, queryEmbedding []float32, filter HybridSearchFilter, limit int) (map[string]*VectorResult, error) {
	where := "WHERE 1=1"
	args := []interface{}{}

	if filter.Type != "" {
		where += " AND m.type = ?"
		args = append(args, filter.Type)
	}
	if filter.Status != "" {
		where += " AND m.status = ?"
		args = append(args, filter.Status)
	}

	sqlQuery := fmt.Sprintf(`
		SELECT m.id, m.created_at, m.updated_at, m.type, m.title, m.content, m.source, m.status, m.tags, e.embedding
		FROM memories m
		LEFT JOIN memory_embeddings e ON m.id = e.memory_id
		%s
		ORDER BY m.created_at DESC
		LIMIT ?
	`, where)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]*VectorResult)
	for rows.Next() {
		mem := &Memory{}
		var tagsStr string
		var createdAtStr, updatedAtStr string
		var embeddingBytes []byte

		if err := rows.Scan(&mem.ID, &createdAtStr, &updatedAtStr, &mem.Type, &mem.Title, &mem.Content, &mem.Source, &mem.Status, &tagsStr, &embeddingBytes); err != nil {
			continue
		}

		mem.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		mem.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		if tagsStr != "" {
			mem.Tags = strings.Split(tagsStr, "|")
		}

		var vectorScore float64
		if embeddingBytes != nil {
			memEmbedding := embeddings.DeserializeEmbedding(embeddingBytes)
			if memEmbedding != nil {
				vectorScore = float64(embeddings.CosineSimilarity(queryEmbedding, memEmbedding))
			}
		}

		results[mem.ID] = &VectorResult{
			Memory: mem,
			Score:  vectorScore,
		}
	}

	return results, nil
}

func (r *Repository) mergeSearchResults(ctx context.Context, fts5Results map[string]*FTS5Result, vectorResults map[string]*VectorResult, queryEmbedding []float32, k int) []*HybridSearchResult {
	seen := make(map[string]bool)
	merged := make([]*HybridSearchResult, 0)

	normalizer := func(score, min, max float64) float64 {
		if max == min {
			return 0.5
		}
		return (score - min) / (max - min)
	}

	var maxFTS5, minFTS5 float64
	for _, result := range fts5Results {
		if result.Score > maxFTS5 {
			maxFTS5 = result.Score
		}
		if result.Score < minFTS5 {
			minFTS5 = result.Score
		}
	}

	var maxVector, minVector float64
	for _, result := range vectorResults {
		if result.Score > maxVector {
			maxVector = result.Score
		}
		if result.Score < minVector {
			minVector = result.Score
		}
	}

	for id, fts5 := range fts5Results {
		if seen[id] {
			continue
		}
		seen[id] = true

		normalizedFTS5 := normalizer(fts5.Score, minFTS5, maxFTS5)

		vector, vectorExists := vectorResults[id]
		normalizedVector := 0.0
		if vectorExists {
			normalizedVector = normalizer(vector.Score, minVector, maxVector)
		}

		combinedScore := normalizedFTS5*0.6 + normalizedVector*0.4
		matchType := "fts5"
		if vectorExists {
			matchType = "both"
			combinedScore = normalizedFTS5*0.5 + normalizedVector*0.5
		}

		merged = append(merged, &HybridSearchResult{
			Memory:        fts5.Memory,
			VectorScore:   normalizedVector,
			FTS5Score:     normalizedFTS5,
			CombinedScore: combinedScore,
			MatchType:     matchType,
		})
	}

	for id, vector := range vectorResults {
		if seen[id] {
			continue
		}
		seen[id] = true

		normalizedVector := normalizer(vector.Score, minVector, maxVector)

		merged = append(merged, &HybridSearchResult{
			Memory:        vector.Memory,
			VectorScore:   normalizedVector,
			FTS5Score:     0.0,
			CombinedScore: normalizedVector * 0.8,
			MatchType:     "vector",
		})
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].CombinedScore > merged[j].CombinedScore
	})

	if len(merged) > k {
		merged = merged[:k]
	}

	return merged
}

func calculateBM25Score(rank float64) float64 {
	if rank == 0 {
		return 1.0
	}
	return 1.0 / rank
}

func (r *Repository) getEmbedding(ctx context.Context, memoryID string) ([]float32, error) {
	var embeddingBytes []byte
	err := r.db.QueryRowContext(ctx, "SELECT embedding FROM memory_embeddings WHERE memory_id = ?", memoryID).Scan(&embeddingBytes)
	if err != nil {
		return nil, err
	}

	if embeddingBytes == nil {
		return nil, nil
	}

	return embeddings.DeserializeEmbedding(embeddingBytes), nil
}

func (r *Repository) Update(ctx context.Context, id string, update *MemoryUpdate) error {
	sets := []string{}
	args := []interface{}{}

	if update.Title != nil {
		sets = append(sets, "title = ?")
		args = append(args, *update.Title)
	}
	if update.Content != nil {
		sets = append(sets, "content = ?")
		args = append(args, *update.Content)
	}
	if update.Status != nil {
		sets = append(sets, "status = ?")
		args = append(args, *update.Status)
	}
	if update.Tags != nil {
		sets = append(sets, "tags = ?")
		args = append(args, strings.Join(update.Tags, "|"))
	}
	if len(sets) == 0 {
		return nil
	}

	sets = append(sets, "updated_at = ?")
	args = append(args, time.Now().UTC().Format(time.RFC3339))
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, fmt.Sprintf("UPDATE memories SET %s WHERE id = ?", strings.Join(sets, ", ")), args...)
	return err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM memories WHERE id = ?", id)
	return err
}
