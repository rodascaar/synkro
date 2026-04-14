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

	err := r.db.QueryRowContext(ctx, `
		SELECT id, created_at, updated_at, type, title, content, source, status, tags
		FROM memories WHERE id = ?
	`, id).Scan(&mem.ID, &mem.CreatedAt, &mem.UpdatedAt, &mem.Type, &mem.Title, &mem.Content, &mem.Source, &mem.Status, &tagsStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if tagsStr != "" {
		mem.Tags = strings.Split(tagsStr, "|")
	}

	return mem, nil
}

func (r *Repository) Search(ctx context.Context, query string, filter MemoryFilter) ([]*Memory, error) {
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

	var sqlQuery string
	if query != "" {
		sqlQuery = fmt.Sprintf(`
			SELECT m.id, m.created_at, m.updated_at, m.type, m.title, m.content, m.source, m.status, m.tags, rank
			FROM memories m
			INNER JOIN memories_fts f ON m.id = f.id
			%s AND memories_fts MATCH ?
			ORDER BY rank
			LIMIT ?
		`, where)
		args = append(args, query)
		if filter.Limit > 0 {
			args = append(args, filter.Limit)
		} else {
			args = append(args, 50)
		}
	} else {
		sqlQuery = fmt.Sprintf("SELECT m.id, m.created_at, m.updated_at, m.type, m.title, m.content, m.source, m.status, m.tags FROM memories m %s ORDER BY m.created_at DESC", where)
		if filter.Limit > 0 {
			sqlQuery += " LIMIT ?"
			args = append(args, filter.Limit)
		}
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
		var bm25Rank float64 = 0

		if query != "" {
			if err := rows.Scan(&mem.ID, &createdAtStr, &updatedAtStr, &mem.Type, &mem.Title, &mem.Content, &mem.Source, &mem.Status, &tagsStr, &bm25Rank); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(&mem.ID, &createdAtStr, &updatedAtStr, &mem.Type, &mem.Title, &mem.Content, &mem.Source, &mem.Status, &tagsStr); err != nil {
				return nil, err
			}
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

func (r *Repository) HybridSearch(ctx context.Context, query string, k int, filter HybridSearchFilter) ([]*HybridSearchResult, error) {
	results := make([]*HybridSearchResult, 0, k)

	memories, err := r.Search(ctx, query, MemoryFilter{
		Type:   filter.Type,
		Status: filter.Status,
		Limit:  k,
	})
	if err != nil {
		return nil, err
	}

	for _, mem := range memories {
		vectorScore := 0.0
		if r.embeddingGenerator != nil {
			queryEmbedding, err := r.embeddingGenerator.Generate(ctx, query)
			if err == nil {
				memEmbedding, err := r.getEmbedding(ctx, mem.ID)
				if err == nil && memEmbedding != nil {
					vectorScore = float64(embeddings.CosineSimilarity(queryEmbedding, memEmbedding))
				}
			}
		}

		results = append(results, &HybridSearchResult{
			Memory:        mem,
			VectorScore:   vectorScore,
			FTS5Score:     0.5,
			CombinedScore: vectorScore*0.5 + 0.25,
			MatchType:     "fts5",
		})
	}

	if len(results) > 1 {
		sort.Slice(results, func(i, j int) bool {
			return results[i].CombinedScore > results[j].CombinedScore
		})
	}

	return results, nil
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
