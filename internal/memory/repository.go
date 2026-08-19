package memory

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rodascaar/synkro/internal/db"
	"github.com/rodascaar/synkro/internal/embeddings"
)

const (
	defaultLimit = 50
	// searchFetchMultiplier amplía la ventana de candidatos previo al corte final.
	searchFetchMultiplier = 3

	weightFTS5Only   = 0.6 // Match solo FTS5 (sin embedding).
	weightVectorOnly = 0.8 // Match solo vectorial (sin FTS5).
	weightBoth       = 0.5 // Peso de cada componente cuando hay match híbrido.
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
		id = fmt.Sprintf("mem-%s", uuid.New().String()[:8])
	}

	now := time.Now().UTC()
	if mem.CreatedAt.IsZero() {
		mem.CreatedAt = now
	}
	if mem.UpdatedAt.IsZero() {
		mem.UpdatedAt = now
	}

	pinnedInt := 0
	if mem.Pinned {
		pinnedInt = 1
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO memories (id, created_at, updated_at, type, title, content, source, status, tags, topic_key, pinned)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, mem.CreatedAt.Format(time.RFC3339), mem.UpdatedAt.Format(time.RFC3339), mem.Type, mem.Title, mem.Content, sourcePtrVal(mem.Source), mem.Status, strings.Join(mem.Tags, "|"), mem.TopicKey, pinnedInt)
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
			if saveErr := r.saveEmbedding(ctx, r.db, mem.ID, embedding); saveErr != nil {
				log.Printf("warning: failed to save embedding for %s: %v", mem.ID, saveErr)
			}
		} else {
			log.Printf("warning: failed to generate embedding for %s: %v", mem.ID, err)
		}
	}

	return nil
}

// Upsert crea una memoria nueva o actualiza una existente por topic_key.
// Si mem.TopicKey no está vacío y existe una memoria activa con esa clave,
// la actualiza; en caso contrario, crea una nueva.
func (r *Repository) Upsert(ctx context.Context, mem *Memory) error {
	if mem.TopicKey != "" {
		if existing, err := r.GetByTopicKey(ctx, mem.TopicKey); err != nil {
			return err
		} else if existing != nil {
			mem.ID = existing.ID
			return r.UpdateMemory(ctx, mem)
		}
	}
	return r.Create(ctx, mem)
}

func (r *Repository) saveEmbedding(ctx context.Context, exec db.Executor, memoryID string, embedding []float32) error {
	_, err := exec.ExecContext(ctx, `
		INSERT INTO memory_embeddings (memory_id, embedding, created_at)
		VALUES (?, ?, ?)
		ON CONFLICT(memory_id) DO UPDATE SET embedding = excluded.embedding, created_at = excluded.created_at
	`, memoryID, embeddings.SerializeEmbedding(embedding), time.Now().UTC().Format(time.RFC3339))
	return err
}

func (r *Repository) Get(ctx context.Context, id string) (*Memory, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, created_at, updated_at, type, title, content, COALESCE(source, ''), status, COALESCE(tags, ''), COALESCE(topic_key, ''), pinned
		FROM memories WHERE id = ?
	`, id)
	mem, err := scanMemory(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mem, nil
}

// GetMany devuelve las memorias cuyos IDs están en ids, preservando el orden
// de entrada. Los IDs inexistentes se omiten.
func (r *Repository) GetMany(ctx context.Context, ids []string) ([]*Memory, error) {
	if len(ids) == 0 {
		return []*Memory{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	seen := make(map[string]bool)
	j := 0
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		placeholders[j] = "?"
		args[j] = id
		j++
	}
	if j == 0 {
		return []*Memory{}, nil
	}

	query := `
		SELECT id, created_at, updated_at, type, title, content, COALESCE(source, ''), status, COALESCE(tags, ''), COALESCE(topic_key, ''), pinned
		FROM memories WHERE id IN (` + strings.Join(placeholders[:j], ",") + `)`

	rows, err := r.db.QueryContext(ctx, query, args[:j]...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	byID := make(map[string]*Memory, j)
	for rows.Next() {
		mem, err := scanMemory(rows)
		if err != nil {
			return nil, err
		}
		byID[mem.ID] = mem
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	memories := make([]*Memory, 0, len(ids))
	added := make(map[string]bool, len(ids))
	for _, id := range ids {
		if added[id] {
			continue
		}
		if mem, ok := byID[id]; ok {
			added[id] = true
			memories = append(memories, mem)
		}
	}
	return memories, nil
}

// Search devuelve memorias según query. Con query vacío lista las más recientes;
// con query no vacío usa el pipeline unificado FTS5 BM25 + re-rank vectorial.
func (r *Repository) Search(ctx context.Context, query string, filter MemoryFilter) ([]*Memory, error) {
	if query == "" {
		return r.list(ctx, filter)
	}

	results, err := r.searchRanked(ctx, query, filter)
	if err != nil {
		return nil, err
	}

	memories := make([]*Memory, 0, len(results))
	for _, res := range results {
		memories = append(memories, res.Memory)
	}
	return memories, nil
}

// HybridSearch es el mismo pipeline unificado que Search pero con scores por resultado.
func (r *Repository) HybridSearch(ctx context.Context, query string, k int, filter HybridSearchFilter) ([]*HybridSearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}
	return r.searchRanked(ctx, query, filter)
}

func (r *Repository) searchRanked(ctx context.Context, query string, filter HybridSearchFilter) ([]*HybridSearchResult, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultLimit
	}

	fts5Results, fts5Err := r.searchWithBM25(ctx, query, filter, limit*searchFetchMultiplier)

	var vectorResults map[string]*VectorResult
	if r.embeddingGenerator != nil {
		if queryEmbedding, err := r.embeddingGenerator.Generate(ctx, query); err == nil {
			vectorResults, _ = r.scoreByVector(ctx, queryEmbedding, filter, limit*searchFetchMultiplier)
		}
	}

	if fts5Err != nil {
		log.Printf("warning: FTS5 search failed, using vector-only results: %v", fts5Err)
	}

	if len(fts5Results) == 0 || fts5Err != nil {
		return vectorOnlyResults(vectorResults, limit), nil
	}
	return mergeSearchResults(fts5Results, vectorResults, limit), nil
}

func (r *Repository) searchWithBM25(ctx context.Context, query string, filter HybridSearchFilter, limit int) (map[string]*FTS5Result, error) {
	where, args := buildFilterWhere(filter)

	sqlQuery := fmt.Sprintf(`
		SELECT m.id, m.created_at, m.updated_at, m.type, m.title, m.content, COALESCE(m.source, ''), m.status, COALESCE(m.tags, ''), COALESCE(m.topic_key, ''), m.pinned, rank
		FROM memories m
		INNER JOIN memories_fts f ON m.id = f.id
		%s AND memories_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, where)
	args = append(args, sanitizeFTS5Query(query), limit)

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	results := make(map[string]*FTS5Result)
	for rows.Next() {
		mem, bm25Rank, err := scanMemoryWithRank(rows)
		if err != nil {
			return nil, err
		}
		results[mem.ID] = &FTS5Result{
			Memory: mem,
			Rank:   bm25Rank,
			Score:  bm25ToScore(bm25Rank),
		}
	}
	return results, nil
}

// scoreByVector calcula similitud coseno in-memory contra todos los candidatos
// que cumplen el filtro y tienen embedding.
func (r *Repository) scoreByVector(ctx context.Context, queryEmbedding []float32, filter HybridSearchFilter, limit int) (map[string]*VectorResult, error) {
	where, args := buildFilterWhere(filter)

	sqlQuery := fmt.Sprintf(`
		SELECT m.id, m.created_at, m.updated_at, m.type, m.title, m.content, COALESCE(m.source, ''), m.status, COALESCE(m.tags, ''), COALESCE(m.topic_key, ''), m.pinned, e.embedding
		FROM memories m
		INNER JOIN memory_embeddings e ON m.id = e.memory_id
		%s
		ORDER BY m.pinned DESC, m.created_at DESC
		LIMIT ?
	`, where)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	results := make(map[string]*VectorResult)
	for rows.Next() {
		mem, embeddingBytes, err := scanMemoryWithEmbedding(rows)
		if err != nil {
			continue
		}
		score := 0.0
		if embeddingBytes != nil {
			if memEmbedding := embeddings.DeserializeEmbedding(embeddingBytes); memEmbedding != nil {
				score = float64(embeddings.CosineSimilarity(queryEmbedding, memEmbedding))
			}
		}
		results[mem.ID] = &VectorResult{Memory: mem, Score: score}
	}
	return results, nil
}

func vectorOnlyResults(vectorResults map[string]*VectorResult, k int) []*HybridSearchResult {
	if len(vectorResults) == 0 {
		return []*HybridSearchResult{}
	}

	maxScore, minScore := scoresRange(vectorResults)

	results := make([]*HybridSearchResult, 0, len(vectorResults))
	for _, vector := range vectorResults {
		normalized := normalizeScore(vector.Score, minScore, maxScore)
		results = append(results, &HybridSearchResult{
			Memory:        vector.Memory,
			VectorScore:   vector.Score,
			FTS5Score:     0.0,
			CombinedScore: normalized,
			MatchType:     "vector",
		})
	}

	sortResults(results)
	return truncateResults(results, k)
}

func mergeSearchResults(fts5Results map[string]*FTS5Result, vectorResults map[string]*VectorResult, k int) []*HybridSearchResult {
	_, maxFTS5, minFTS5 := scoresRangeFTS5(fts5Results)
	maxVector, minVector := 0.0, 0.0
	if len(vectorResults) > 0 {
		maxVector, minVector = scoresRange(vectorResults)
	}

	seen := make(map[string]bool, len(fts5Results))
	merged := make([]*HybridSearchResult, 0, len(fts5Results))

	for id, fts5 := range fts5Results {
		seen[id] = true
		normalizedFTS5 := normalizeScore(fts5.Score, minFTS5, maxFTS5)

		vectorScore := 0.0
		matchType := "fts5"
		combined := normalizedFTS5 * weightFTS5Only
		if vector, ok := vectorResults[id]; ok {
			vectorScore = vector.Score
			normalizedVector := normalizeScore(vector.Score, minVector, maxVector)
			matchType = "both"
			combined = normalizedFTS5*weightBoth + normalizedVector*weightBoth
		}

		merged = append(merged, &HybridSearchResult{
			Memory:        fts5.Memory,
			VectorScore:   vectorScore,
			FTS5Score:     normalizedFTS5,
			CombinedScore: combined,
			MatchType:     matchType,
		})
	}

	for id, vector := range vectorResults {
		if seen[id] {
			continue
		}
		seen[id] = true
		normalized := normalizeScore(vector.Score, minVector, maxVector)
		merged = append(merged, &HybridSearchResult{
			Memory:        vector.Memory,
			VectorScore:   vector.Score,
			FTS5Score:     0.0,
			CombinedScore: normalized * weightVectorOnly,
			MatchType:     "vector",
		})
	}

	sortResults(merged)
	return truncateResults(merged, k)
}

func buildFilterWhere(filter HybridSearchFilter) (string, []interface{}) {
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
	return where, args
}

func (r *Repository) list(ctx context.Context, filter MemoryFilter) ([]*Memory, error) {
	where, args := buildFilterWhere(HybridSearchFilter{Type: filter.Type, Status: filter.Status})

	limit := filter.Limit
	if limit <= 0 {
		limit = defaultLimit
	}

	sqlQuery := fmt.Sprintf(`
		SELECT m.id, m.created_at, m.updated_at, m.type, m.title, m.content, COALESCE(m.source, ''), m.status, COALESCE(m.tags, ''), COALESCE(m.topic_key, ''), m.pinned
		FROM memories m
		%s
		ORDER BY m.pinned DESC, m.created_at DESC
		LIMIT ?
	`, where)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	memories := make([]*Memory, 0)
	for rows.Next() {
		mem, err := scanMemory(rows)
		if err != nil {
			return nil, err
		}
		memories = append(memories, mem)
	}
	return memories, nil
}

func (r *Repository) Update(ctx context.Context, id string, update *MemoryUpdate) error {
	sets := make([]string, 0, 4)
	args := make([]interface{}, 0, 4)

	titleChanged := update.Title != nil
	contentChanged := update.Content != nil

	if titleChanged {
		sets = append(sets, "title = ?")
		args = append(args, *update.Title)
	}
	if contentChanged {
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

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, fmt.Sprintf("UPDATE memories SET %s WHERE id = ?", strings.Join(sets, ", ")), args...); err != nil {
		return err
	}

	if (titleChanged || contentChanged) && r.embeddingGenerator != nil {
		var title, content string
		if err := tx.QueryRowContext(ctx, `SELECT title, content FROM memories WHERE id = ?`, id).Scan(&title, &content); err != nil {
			return err
		}
		if embedding, err := r.embeddingGenerator.Generate(ctx, title+" "+content); err == nil {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO memory_embeddings (memory_id, embedding, created_at)
				VALUES (?, ?, ?)
				ON CONFLICT(memory_id) DO UPDATE SET embedding = excluded.embedding, created_at = excluded.created_at
			`, id, embeddings.SerializeEmbedding(embedding), time.Now().UTC().Format(time.RFC3339)); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *Repository) UpdateMemory(ctx context.Context, mem *Memory) error {
	source := sourcePtrVal(mem.Source)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
		UPDATE memories
		SET title = ?, content = ?, type = ?, source = ?, status = ?, tags = ?, topic_key = ?, updated_at = ?
		WHERE id = ?
	`, mem.Title, mem.Content, mem.Type, source, mem.Status, strings.Join(mem.Tags, "|"), mem.TopicKey, now, mem.ID)
	if err != nil {
		return err
	}

	if r.embeddingGenerator != nil {
		if embedding, err := r.embeddingGenerator.Generate(ctx, mem.Title+" "+mem.Content); err == nil {
			if saveErr := r.saveEmbedding(ctx, r.db, mem.ID, embedding); saveErr != nil {
				log.Printf("warning: failed to save embedding for %s: %v", mem.ID, saveErr)
			}
		} else {
			log.Printf("warning: failed to generate embedding for %s: %v", mem.ID, err)
		}
	}
	return nil
}

func (r *Repository) GetByTag(ctx context.Context, tag string, filter MemoryFilter) ([]*Memory, error) {
	where := "WHERE instr('|' || m.tags || '|', ?) > 0"
	args := []interface{}{"|" + tag + "|"}

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
		limit = defaultLimit
	}

	sqlQuery := fmt.Sprintf(`
		SELECT m.id, m.created_at, m.updated_at, m.type, m.title, m.content, COALESCE(m.source, ''), m.status, COALESCE(m.tags, ''), COALESCE(m.topic_key, ''), m.pinned
		FROM memories m
		%s
		ORDER BY m.pinned DESC, m.created_at DESC
		LIMIT ?
	`, where)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	memories := make([]*Memory, 0)
	for rows.Next() {
		mem, err := scanMemory(rows)
		if err != nil {
			return nil, err
		}
		memories = append(memories, mem)
	}
	return memories, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM memories WHERE id = ?", id)
	return err
}

// SimilarMemory representa una memoria existente y su similitud coseno con un texto.
type SimilarMemory struct {
	Memory *Memory
	Score  float64
}

// SimilarMemories devuelve las memorias activas cuyo coseno con el texto supera
// el umbral, ordenadas de mayor a menor similitud y limitadas a `limit`.
func (r *Repository) SimilarMemories(ctx context.Context, text string, threshold float64, limit int) ([]SimilarMemory, error) {
	if r.embeddingGenerator == nil {
		return nil, nil
	}

	embedding, err := r.embeddingGenerator.Generate(ctx, text)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT m.id, m.created_at, m.updated_at, m.type, m.title, m.content, COALESCE(m.source, ''), m.status, COALESCE(m.tags, ''), COALESCE(m.topic_key, ''), m.pinned, e.embedding
		FROM memory_embeddings e
		INNER JOIN memories m ON m.id = e.memory_id
		WHERE m.status = 'active'
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var similar []SimilarMemory
	for rows.Next() {
		mem, embeddingBytes, err := scanMemoryWithEmbedding(rows)
		if err != nil {
			continue
		}
		if embeddingBytes == nil {
			continue
		}
		memEmbedding := embeddings.DeserializeEmbedding(embeddingBytes)
		if memEmbedding == nil {
			continue
		}
		score := float64(embeddings.CosineSimilarity(embedding, memEmbedding))
		if score >= threshold {
			similar = append(similar, SimilarMemory{Memory: mem, Score: score})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(similar, func(i, j int) bool {
		return similar[i].Score > similar[j].Score
	})
	if limit > 0 && len(similar) > limit {
		similar = similar[:limit]
	}
	return similar, nil
}

// MaxSimilarity devuelve la similitud coseno máxima entre el texto y las
// memorias existentes. Útil para detectar duplicados al agregar una memoria.
func (r *Repository) MaxSimilarity(ctx context.Context, text string) (float64, error) {
	similar, err := r.SimilarMemories(ctx, text, 0, 1)
	if err != nil {
		return 0, err
	}
	if len(similar) == 0 {
		return 0, nil
	}
	return similar[0].Score, nil
}

// GetByTopicKey devuelve la memoria activa con la clave de tema dada, si existe.
func (r *Repository) GetByTopicKey(ctx context.Context, topicKey string) (*Memory, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, created_at, updated_at, type, title, content, COALESCE(source, ''), status, COALESCE(tags, ''), COALESCE(topic_key, ''), pinned
		FROM memories WHERE topic_key = ? AND status = 'active'
	`, topicKey)
	mem, err := scanMemory(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mem, nil
}

// Pin marca o desmarca una memoria como fijada.
func (r *Repository) Pin(ctx context.Context, id string, pinned bool) error {
	pinnedInt := 0
	if pinned {
		pinnedInt = 1
	}
	_, err := r.db.ExecContext(ctx, `UPDATE memories SET pinned = ?, updated_at = ? WHERE id = ?`, pinnedInt, time.Now().UTC().Format(time.RFC3339), id)
	return err
}

// DeleteAll borra todas las memorias y sesiones. Los embeddings, relaciones y
// entregas se eliminan en cascada vía foreign keys.
func (r *Repository) DeleteAll(ctx context.Context) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, "DELETE FROM memories"); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM sessions"); err != nil {
		return err
	}
	return tx.Commit()
}

// --- helpers de scoring ---

// bm25ToScore convierte el rank negativo de FTS5 en un score positivo donde
// mayor = mejor. FTS5 bm25() devuelve valores negativos (menor = mejor).
func bm25ToScore(rank float64) float64 {
	return -rank
}

func normalizeScore(score, min, max float64) float64 {
	if max == min {
		return 1.0
	}
	return (score - min) / (max - min)
}

func scoresRange(results map[string]*VectorResult) (float64, float64) {
	var max, min float64
	first := true
	for _, v := range results {
		if first || v.Score > max {
			max = v.Score
		}
		if first || v.Score < min {
			min = v.Score
		}
		first = false
	}
	return max, min
}

func scoresRangeFTS5(results map[string]*FTS5Result) (float64, float64, float64) {
	var max, min float64
	first := true
	for _, v := range results {
		if first || v.Score > max {
			max = v.Score
		}
		if first || v.Score < min {
			min = v.Score
		}
		first = false
	}
	return 0.0, max, min
}

func sortResults(results []*HybridSearchResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].CombinedScore > results[j].CombinedScore
	})
}

func truncateResults(results []*HybridSearchResult, k int) []*HybridSearchResult {
	if len(results) > k {
		results = results[:k]
	}
	return results
}

// --- helpers de escaneo ---

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanMemory(row rowScanner) (*Memory, error) {
	mem := &Memory{}
	var tagsStr, topicKeyStr, createdAtStr, updatedAtStr string
	var source sql.NullString
	var pinned int
	if err := row.Scan(&mem.ID, &createdAtStr, &updatedAtStr, &mem.Type, &mem.Title, &mem.Content, &source, &mem.Status, &tagsStr, &topicKeyStr, &pinned); err != nil {
		return nil, err
	}
	return buildMemory(mem, createdAtStr, updatedAtStr, source, tagsStr, topicKeyStr, pinned), nil
}

func scanMemoryWithRank(row rowScanner) (*Memory, float64, error) {
	mem := &Memory{}
	var tagsStr, topicKeyStr, createdAtStr, updatedAtStr string
	var source sql.NullString
	var pinned int
	var rank float64
	if err := row.Scan(&mem.ID, &createdAtStr, &updatedAtStr, &mem.Type, &mem.Title, &mem.Content, &source, &mem.Status, &tagsStr, &topicKeyStr, &pinned, &rank); err != nil {
		return nil, 0, err
	}
	return buildMemory(mem, createdAtStr, updatedAtStr, source, tagsStr, topicKeyStr, pinned), rank, nil
}

func scanMemoryWithEmbedding(row rowScanner) (*Memory, []byte, error) {
	mem := &Memory{}
	var tagsStr, topicKeyStr, createdAtStr, updatedAtStr string
	var source sql.NullString
	var pinned int
	var embedding []byte
	if err := row.Scan(&mem.ID, &createdAtStr, &updatedAtStr, &mem.Type, &mem.Title, &mem.Content, &source, &mem.Status, &tagsStr, &topicKeyStr, &pinned, &embedding); err != nil {
		return nil, nil, err
	}
	return buildMemory(mem, createdAtStr, updatedAtStr, source, tagsStr, topicKeyStr, pinned), embedding, nil
}

func buildMemory(mem *Memory, createdAtStr, updatedAtStr string, source sql.NullString, tagsStr, topicKeyStr string, pinned int) *Memory {
	mem.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	mem.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
	mem.Source = scanSource(source)
	if tagsStr != "" {
		mem.Tags = strings.Split(tagsStr, "|")
	}
	mem.TopicKey = topicKeyStr
	mem.Pinned = pinned != 0
	return mem
}

func sanitizeFTS5Query(query string) string {
	terms := strings.Fields(query)
	safe := make([]string, 0, len(terms))
	for _, term := range terms {
		term = strings.Trim(term, "\"'*():")
		term = strings.ReplaceAll(term, "\"", "")
		if term == "" {
			continue
		}
		safe = append(safe, "\""+term+"\"")
	}
	if len(safe) == 0 {
		return "\"\""
	}
	return strings.Join(safe, " ")
}

func scanSource(ns sql.NullString) *string {
	if ns.Valid && ns.String != "" {
		return &ns.String
	}
	return nil
}

func sourcePtrVal(s *string) sql.NullString {
	if s != nil {
		return sql.NullString{String: *s, Valid: true}
	}
	return sql.NullString{}
}