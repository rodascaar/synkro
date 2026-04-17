package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	synkroerrors "github.com/rodascaar/synkro/internal/errors"
	"github.com/rodascaar/synkro/internal/memory"
)

type AddMemoryInput struct {
	Type    string   `json:"type"`
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Source  string   `json:"source"`
	Tags    []string `json:"tags"`
}

type GetMemoryInput struct {
	ID string `json:"id"`
}

type ListMemoryInput struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Limit  int    `json:"limit"`
}

type SearchMemoryInput struct {
	Query  string `json:"query"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Limit  int    `json:"limit"`
}

type UpdateMemoryInput struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Status  string   `json:"status"`
	Tags    []string `json:"tags"`
}

type ArchiveMemoryInput struct {
	ID string `json:"id"`
}

type ActivateContextInput struct {
	Query     string `json:"query"`
	SessionID string `json:"session_id"`
	MaxTokens int    `json:"max_tokens"`
	Limit     int    `json:"limit"`
}

type AddRelationInput struct {
	SourceID string  `json:"source_id"`
	TargetID string  `json:"target_id"`
	Type     string  `json:"type"`
	Strength float64 `json:"strength"`
}

type GetRelationsInput struct {
	MemoryID string `json:"memory_id"`
}

type DeleteRelationInput struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
}

type FindPathInput struct {
	FromID string `json:"from_id"`
	ToID   string `json:"to_id"`
}

func (s *Server) AddMemory(ctx context.Context, input AddMemoryInput) ([]byte, error) {
	var buf bytes.Buffer
	if err := s.AddMemoryWithWriter(ctx, input, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *Server) AddMemoryWithWriter(ctx context.Context, input AddMemoryInput, w io.Writer) error {
	if input.Title == "" {
		return synkroerrors.Wrap(
			fmt.Errorf("title is required"),
			synkroerrors.ErrInvalidInput.Code,
			synkroerrors.ErrInvalidInput.Message,
			synkroerrors.ErrInvalidInput.Help,
		)
	}

	validTypes := map[string]bool{
		"note": true, "decision": true, "task": true, "context": true,
	}
	if input.Type != "" && !validTypes[input.Type] {
		return synkroerrors.Wrap(
			fmt.Errorf("invalid type %q, must be one of: note, decision, task, context", input.Type),
			synkroerrors.ErrInvalidInput.Code,
			synkroerrors.ErrInvalidInput.Message,
			synkroerrors.ErrInvalidInput.Help,
		)
	}

	mem := &memory.Memory{
		Type:    input.Type,
		Title:   input.Title,
		Content: input.Content,
		Source:  input.Source,
		Status:  "active",
		Tags:    input.Tags,
	}

	if err := s.repo.Create(ctx, mem); err != nil {
		return synkroerrors.Wrap(err, synkroerrors.ErrEmbeddingFailed.Code, "Error creating memory", synkroerrors.ErrEmbeddingFailed.Help)
	}

	response := map[string]interface{}{
		"success":          true,
		"memory_id":        mem.ID,
		"similarity_score": 0.0,
		"embedding_used":   s.embeddingType,
	}

	return writeJSON(w, response)
}

func (s *Server) GetMemory(ctx context.Context, input GetMemoryInput, w io.Writer) error {
	if input.ID == "" {
		return synkroerrors.Wrap(
			fmt.Errorf("id is required"),
			synkroerrors.ErrInvalidInput.Code,
			synkroerrors.ErrInvalidInput.Message,
			synkroerrors.ErrInvalidInput.Help,
		)
	}

	mem, err := s.repo.Get(ctx, input.ID)
	if err != nil {
		return synkroerrors.Wrap(err, "DB_ERROR", "Error getting memory", "Check the memory ID and try again")
	}
	if mem == nil {
		return synkroerrors.ErrMemoryNotFound
	}

	var relations []interface{}
	if s.graph != nil {
		rels, err := s.graph.GetRelations(ctx, mem.ID)
		if err == nil {
			for _, rel := range rels {
				relations = append(relations, map[string]interface{}{
					"source_id":  rel.SourceID,
					"target_id":  rel.TargetID,
					"type":       rel.Type,
					"strength":   rel.Strength,
					"created_at": rel.CreatedAt.Format(time.RFC3339),
				})
			}
		}
	}

	response := map[string]interface{}{
		"memory": map[string]interface{}{
			"id":         mem.ID,
			"type":       mem.Type,
			"title":      mem.Title,
			"content":    mem.Content,
			"source":     mem.Source,
			"status":     mem.Status,
			"tags":       mem.Tags,
			"created_at": mem.CreatedAt.Format(time.RFC3339),
			"updated_at": mem.UpdatedAt.Format(time.RFC3339),
		},
		"relations": relations,
	}

	return writeJSON(w, response)
}

func (s *Server) ListMemory(ctx context.Context, input ListMemoryInput, w io.Writer) error {
	filter := memory.MemoryFilter{
		Type:   input.Type,
		Status: input.Status,
		Limit:  input.Limit,
	}

	memories, err := s.repo.Search(ctx, "", filter)
	if err != nil {
		return synkroerrors.Wrap(err, "DB_ERROR", "Error listing memories", "Check your database and filters")
	}

	response := map[string]interface{}{
		"memories":  memories,
		"count":     len(memories),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return writeJSON(w, response)
}

func (s *Server) SearchMemory(ctx context.Context, input SearchMemoryInput, w io.Writer) error {
	if input.Query == "" {
		return synkroerrors.Wrap(
			fmt.Errorf("query is required"),
			synkroerrors.ErrInvalidInput.Code,
			synkroerrors.ErrInvalidInput.Message,
			synkroerrors.ErrInvalidInput.Help,
		)
	}

	filter := memory.MemoryFilter{
		Type:   input.Type,
		Status: input.Status,
		Limit:  input.Limit,
	}

	memories, err := s.repo.Search(ctx, input.Query, filter)
	if err != nil {
		return synkroerrors.Wrap(err, synkroerrors.ErrFTS5Query.Code, "Error searching memories", synkroerrors.ErrFTS5Query.Help)
	}

	response := map[string]interface{}{
		"query":     input.Query,
		"memories":  memories,
		"count":     len(memories),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return writeJSON(w, response)
}

func (s *Server) UpdateMemory(ctx context.Context, input UpdateMemoryInput, w io.Writer) error {
	if input.ID == "" {
		return synkroerrors.Wrap(
			fmt.Errorf("id is required"),
			synkroerrors.ErrInvalidInput.Code,
			synkroerrors.ErrInvalidInput.Message,
			synkroerrors.ErrInvalidInput.Help,
		)
	}

	update := &memory.MemoryUpdate{}

	if input.Title != "" {
		update.Title = &input.Title
	}
	if input.Content != "" {
		update.Content = &input.Content
	}
	if input.Status != "" {
		update.Status = &input.Status
	}
	if input.Tags != nil {
		update.Tags = input.Tags
	}

	if err := s.repo.Update(ctx, input.ID, update); err != nil {
		return synkroerrors.Wrap(err, "DB_ERROR", "Error updating memory", "Check the memory ID and try again")
	}

	response := map[string]interface{}{
		"success":    true,
		"memory_id":  input.ID,
		"updated_at": time.Now().Format(time.RFC3339),
	}

	return writeJSON(w, response)
}

func (s *Server) ArchiveMemory(ctx context.Context, input ArchiveMemoryInput, w io.Writer) error {
	if input.ID == "" {
		return synkroerrors.Wrap(
			fmt.Errorf("id is required"),
			synkroerrors.ErrInvalidInput.Code,
			synkroerrors.ErrInvalidInput.Message,
			synkroerrors.ErrInvalidInput.Help,
		)
	}

	update := &memory.MemoryUpdate{
		Status: func() *string { s := "archived"; return &s }(),
	}

	if err := s.repo.Update(ctx, input.ID, update); err != nil {
		return synkroerrors.Wrap(err, "DB_ERROR", "Error archiving memory", "Check the memory ID and try again")
	}

	response := map[string]interface{}{
		"success":     true,
		"memory_id":   input.ID,
		"archived_at": time.Now().Format(time.RFC3339),
	}

	return writeJSON(w, response)
}

func (s *Server) ActivateContext(ctx context.Context, input ActivateContextInput, w io.Writer) error {
	if input.Query == "" {
		return synkroerrors.Wrap(
			fmt.Errorf("query is required"),
			synkroerrors.ErrInvalidInput.Code,
			synkroerrors.ErrInvalidInput.Message,
			synkroerrors.ErrInvalidInput.Help,
		)
	}

	if input.MaxTokens <= 0 {
		input.MaxTokens = 4000
	}
	if input.Limit <= 0 {
		input.Limit = 10
	}
	if input.SessionID == "" {
		input.SessionID = "default"
	}

	duplicateDetected := false
	if s.sessionTracker != nil {
		if s.sessionTracker.IsDuplicateQuery(input.SessionID, input.Query) {
			duplicateDetected = true
		}
	}

	filter := memory.HybridSearchFilter{
		Limit:  input.Limit,
		Status: "active",
	}

	results, err := s.repo.HybridSearch(ctx, input.Query, input.Limit, filter)
	if err != nil {
		return synkroerrors.Wrap(err, synkroerrors.ErrVecSearch.Code, "Error searching memories", synkroerrors.ErrVecSearch.Help)
	}

	if len(results) == 0 {
		response := map[string]interface{}{
			"query":                input.Query,
			"session_id":           input.SessionID,
			"duplicate_detected":   duplicateDetected,
			"max_tokens":           input.MaxTokens,
			"total_tokens":         0,
			"primary_results":      []interface{}{},
			"low_priority_results": []interface{}{},
			"warning":              "No memories found matching query",
		}

		return writeJSON(w, response)
	}

	maxSimilarity := results[0].VectorScore
	if maxSimilarity < 0.1 {
		response := map[string]interface{}{
			"query":                input.Query,
			"session_id":           input.SessionID,
			"duplicate_detected":   duplicateDetected,
			"max_tokens":           input.MaxTokens,
			"total_tokens":         0,
			"primary_results":      []interface{}{},
			"low_priority_results": []interface{}{},
			"warning":              "Low similarity - results may not be relevant",
		}

		return writeJSON(w, response)
	}

	var prioritized []*memory.HybridSearchResult
	var lowPriority []*memory.HybridSearchResult

	if s.sessionTracker != nil && duplicateDetected {
		recentDeliveries := s.sessionTracker.GetRecentDeliveries(ctx, input.SessionID, 20)

		for _, result := range results {
			isRecent := false
			for _, recentID := range recentDeliveries {
				if result.Memory.ID == recentID {
					isRecent = true
					break
				}
			}

			if isRecent {
				lowPriority = append(lowPriority, result)
			} else {
				prioritized = append(prioritized, result)
			}
		}
	} else {
		prioritized = results
	}

	if s.contextPruner != nil {
		prioritized = s.contextPruner.Prune(prioritized, input.Query)
		lowPriority = s.contextPruner.Prune(lowPriority, input.Query)
	}

	response := ActivateContextResponse{
		Query:              input.Query,
		SessionID:          input.SessionID,
		DuplicateDetected:  duplicateDetected,
		MaxTokens:          input.MaxTokens,
		TotalTokens:        0,
		PrimaryResults:     convertToContextItems(prioritized, false),
		LowPriorityResults: convertToContextItems(lowPriority, true),
	}

	if s.sessionTracker != nil {
		for _, result := range prioritized {
			s.sessionTracker.MarkAsDelivered(ctx, input.SessionID, result.Memory.ID)
		}
		s.sessionTracker.UpdateLastQuery(ctx, input.SessionID, input.Query)
	}

	if err := writeJSON(w, response); err != nil {
		return synkroerrors.Wrap(err, "MARSHAL_ERROR", "Error marshaling response", "Please report this issue")
	}
	return nil
}

func (s *Server) AddRelation(ctx context.Context, input AddRelationInput, w io.Writer) error {
	if input.SourceID == "" || input.TargetID == "" {
		return synkroerrors.Wrap(
			fmt.Errorf("source_id and target_id are required"),
			synkroerrors.ErrInvalidInput.Code,
			synkroerrors.ErrInvalidInput.Message,
			synkroerrors.ErrInvalidInput.Help,
		)
	}

	if s.graph == nil {
		return synkroerrors.Wrap(
			fmt.Errorf("graph not available"),
			"GRAPH_NOT_AVAILABLE",
			"Graph not available",
			"Ensure the graph module is initialized",
		)
	}

	validTypes := map[string]bool{
		"extends": true, "depends_on": true, "conflicts_with": true,
		"example_of": true, "part_of": true, "related_to": true,
	}
	if !validTypes[input.Type] {
		return synkroerrors.ErrInvalidRelationType
	}

	if input.Strength <= 0 {
		input.Strength = 0.5
	}
	if input.Strength > 1 {
		input.Strength = 1.0
	}

	relation := &memory.MemoryRelation{
		SourceID:  input.SourceID,
		TargetID:  input.TargetID,
		Type:      input.Type,
		Strength:  input.Strength,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.graph.AddRelation(ctx, relation); err != nil {
		return synkroerrors.Wrap(err, "GRAPH_ERROR", "Error adding relation", "Check source_id and target_id exist")
	}

	response := map[string]interface{}{
		"success":   true,
		"source_id": input.SourceID,
		"target_id": input.TargetID,
		"type":      input.Type,
		"strength":  input.Strength,
	}

	return writeJSON(w, response)
}

func (s *Server) GetRelations(ctx context.Context, input GetRelationsInput, w io.Writer) error {
	if input.MemoryID == "" {
		return synkroerrors.Wrap(
			fmt.Errorf("memory_id is required"),
			synkroerrors.ErrInvalidInput.Code,
			synkroerrors.ErrInvalidInput.Message,
			synkroerrors.ErrInvalidInput.Help,
		)
	}

	if s.graph == nil {
		return synkroerrors.Wrap(
			fmt.Errorf("graph not available"),
			"GRAPH_NOT_AVAILABLE",
			"Graph not available",
			"Ensure the graph module is initialized",
		)
	}

	relations, err := s.graph.GetRelations(ctx, input.MemoryID)
	if err != nil {
		return synkroerrors.Wrap(err, "GRAPH_ERROR", "Error getting relations", "Check the memory_id and try again")
	}

	response := map[string]interface{}{
		"memory_id": input.MemoryID,
		"relations": relations,
		"count":     len(relations),
	}

	return writeJSON(w, response)
}

func (s *Server) DeleteRelation(ctx context.Context, input DeleteRelationInput, w io.Writer) error {
	if input.SourceID == "" || input.TargetID == "" {
		return synkroerrors.Wrap(
			fmt.Errorf("source_id and target_id are required"),
			synkroerrors.ErrInvalidInput.Code,
			synkroerrors.ErrInvalidInput.Message,
			synkroerrors.ErrInvalidInput.Help,
		)
	}

	if s.graph == nil {
		return synkroerrors.Wrap(
			fmt.Errorf("graph not available"),
			"GRAPH_NOT_AVAILABLE",
			"Graph not available",
			"Ensure the graph module is initialized",
		)
	}

	if err := s.graph.DeleteRelation(ctx, input.SourceID, input.TargetID); err != nil {
		return synkroerrors.Wrap(err, synkroerrors.ErrRelationNotFound.Code, "Error deleting relation", synkroerrors.ErrRelationNotFound.Help)
	}

	response := map[string]interface{}{
		"success":   true,
		"source_id": input.SourceID,
		"target_id": input.TargetID,
	}

	return writeJSON(w, response)
}

func (s *Server) FindPath(ctx context.Context, input FindPathInput, w io.Writer) error {
	if input.FromID == "" || input.ToID == "" {
		return synkroerrors.Wrap(
			fmt.Errorf("from_id and to_id are required"),
			synkroerrors.ErrInvalidInput.Code,
			synkroerrors.ErrInvalidInput.Message,
			synkroerrors.ErrInvalidInput.Help,
		)
	}

	if s.graph == nil {
		return synkroerrors.Wrap(
			fmt.Errorf("graph not available"),
			"GRAPH_NOT_AVAILABLE",
			"Graph not available",
			"Ensure the graph module is initialized",
		)
	}

	path, err := s.graph.FindPath(ctx, input.FromID, input.ToID)
	if err != nil {
		response := map[string]interface{}{
			"found":   false,
			"error":   err.Error(),
			"from_id": input.FromID,
			"to_id":   input.ToID,
		}
		return writeJSON(w, response)
	}

	response := map[string]interface{}{
		"found":   true,
		"from_id": input.FromID,
		"to_id":   input.ToID,
		"path":    path,
		"length":  len(path),
	}

	return writeJSON(w, response)
}

type ActivateContextResponse struct {
	PrimaryResults     []*ContextResultItem `json:"primary_results"`
	LowPriorityResults []*ContextResultItem `json:"low_priority_results,omitempty"`
	Query              string               `json:"query"`
	SessionID          string               `json:"session_id"`
	DuplicateDetected  bool                 `json:"duplicate_detected"`
	MaxTokens          int                  `json:"max_tokens"`
	TotalTokens        int                  `json:"total_tokens"`
	Warning            string               `json:"warning,omitempty"`
}

type ContextResultItem struct {
	Memory     *MemoryResult `json:"memory"`
	Similarity float64       `json:"similarity"`
	Confidence string        `json:"confidence"`
	IsReminder bool          `json:"is_reminder"`
	Source     string        `json:"source"`
}

type MemoryResult struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Source    string   `json:"source"`
	Status    string   `json:"status"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

func convertToContextItems(results []*memory.HybridSearchResult, isReminder bool) []*ContextResultItem {
	items := make([]*ContextResultItem, 0, len(results))

	for _, result := range results {
		confidence := getConfidenceLevel(result.VectorScore)

		source := "fts5"
		if result.VectorScore > 0 {
			source = "hybrid"
		}

		item := &ContextResultItem{
			Memory: &MemoryResult{
				ID:        result.Memory.ID,
				Type:      result.Memory.Type,
				Title:     result.Memory.Title,
				Content:   result.Memory.Content,
				Source:    result.Memory.Source,
				Status:    result.Memory.Status,
				Tags:      result.Memory.Tags,
				CreatedAt: result.Memory.CreatedAt.Format(time.RFC3339),
				UpdatedAt: result.Memory.UpdatedAt.Format(time.RFC3339),
			},
			Similarity: result.VectorScore,
			Confidence: confidence,
			IsReminder: isReminder,
			Source:     source,
		}
		items = append(items, item)
	}

	return items
}

func getConfidenceLevel(similarity float64) string {
	if similarity >= 0.8 {
		return "high"
	} else if similarity >= 0.5 {
		return "medium"
	} else {
		return "low"
	}
}

func writeJSON(w io.Writer, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
