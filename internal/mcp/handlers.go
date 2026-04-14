package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/rodascaar/synkro/internal/graph"
	"github.com/rodascaar/synkro/internal/memory"
	"github.com/rodascaar/synkro/internal/pruner"
	"github.com/rodascaar/synkro/internal/session"
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

var (
	globalRepo           *memory.Repository
	globalGraph          *graph.Graph
	globalSessionTracker *session.SessionTracker
	globalContextPruner  *pruner.ContextPruner
)

func SetGlobalRepo(repo *memory.Repository) {
	globalRepo = repo
}

func SetGraph(g *graph.Graph) {
	globalGraph = g
}

func SetSessionTracker(tracker *session.SessionTracker) {
	globalSessionTracker = tracker
}

func SetContextPruner(pruner *pruner.ContextPruner) {
	globalContextPruner = pruner
}

func AddMemory(input AddMemoryInput) ([]byte, error) {
	var buf bytes.Buffer
	if err := AddMemoryWithWriter(input, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func AddMemoryWithWriter(input AddMemoryInput, w io.Writer) error {
	ctx := context.Background()
	mem := &memory.Memory{
		Type:    input.Type,
		Title:   input.Title,
		Content: input.Content,
		Source:  input.Source,
		Status:  "active",
		Tags:    input.Tags,
	}

	if err := globalRepo.Create(ctx, mem); err != nil {
		fmt.Fprintf(w, "Error creating memory: %v\n", err)
		return err
	}

	response := map[string]interface{}{
		"success":          true,
		"memory_id":        mem.ID,
		"similarity_score": 0.0,
		"embedding_used":   "tfidf",
	}

	jsonResponse, _ := json.MarshalIndent(response, "", "  ")
	fmt.Fprintf(w, "%s\n", jsonResponse)

	return nil
}

func GetMemory(input GetMemoryInput, w io.Writer) error {
	ctx := context.Background()
	mem, err := globalRepo.Get(ctx, input.ID)
	if err != nil {
		fmt.Fprintf(w, "Error getting memory: %v\n", err)
		return err
	}
	if mem == nil {
		fmt.Fprintf(w, "Memory not found: %s\n", input.ID)
		return nil
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
		"relations": []interface{}{},
	}

	jsonResponse, _ := json.MarshalIndent(response, "", "  ")
	fmt.Fprintf(w, "%s\n", jsonResponse)

	return nil
}

func ListMemory(input ListMemoryInput, w io.Writer) error {
	ctx := context.Background()
	filter := memory.MemoryFilter{
		Type:   input.Type,
		Status: input.Status,
		Limit:  input.Limit,
	}

	memories, err := globalRepo.Search(ctx, "", filter)
	if err != nil {
		fmt.Fprintf(w, "Error listing memories: %v\n", err)
		return err
	}

	response := map[string]interface{}{
		"memories":  memories,
		"count":     len(memories),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonResponse, _ := json.MarshalIndent(response, "", "  ")
	fmt.Fprintf(w, "%s\n", jsonResponse)

	return nil
}

func SearchMemory(input SearchMemoryInput, w io.Writer) error {
	ctx := context.Background()
	filter := memory.MemoryFilter{
		Type:   input.Type,
		Status: input.Status,
		Limit:  input.Limit,
	}

	memories, err := globalRepo.Search(ctx, input.Query, filter)
	if err != nil {
		fmt.Fprintf(w, "Error searching memories: %v\n", err)
		return err
	}

	response := map[string]interface{}{
		"query":     input.Query,
		"memories":  memories,
		"count":     len(memories),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonResponse, _ := json.MarshalIndent(response, "", "  ")
	fmt.Fprintf(w, "%s\n", jsonResponse)

	return nil
}

func UpdateMemory(input UpdateMemoryInput, w io.Writer) error {
	ctx := context.Background()
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

	if err := globalRepo.Update(ctx, input.ID, update); err != nil {
		fmt.Fprintf(w, "Error updating memory: %v\n", err)
		return err
	}

	response := map[string]interface{}{
		"success":    true,
		"memory_id":  input.ID,
		"updated_at": time.Now().Format(time.RFC3339),
	}

	jsonResponse, _ := json.MarshalIndent(response, "", "  ")
	fmt.Fprintf(w, "%s\n", jsonResponse)

	return nil
}

func ArchiveMemory(input ArchiveMemoryInput, w io.Writer) error {
	ctx := context.Background()
	update := &memory.MemoryUpdate{
		Status: func() *string { s := "archived"; return &s }(),
	}

	if err := globalRepo.Update(ctx, input.ID, update); err != nil {
		fmt.Fprintf(w, "Error archiving memory: %v\n", err)
		return err
	}

	response := map[string]interface{}{
		"success":     true,
		"memory_id":   input.ID,
		"archived_at": time.Now().Format(time.RFC3339),
	}

	jsonResponse, _ := json.MarshalIndent(response, "", "  ")
	fmt.Fprintf(w, "%s\n", jsonResponse)

	return nil
}

func ActivateContext(input ActivateContextInput, w io.Writer) error {
	ctx := context.Background()

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
	if globalSessionTracker != nil {
		if globalSessionTracker.IsDuplicateQuery(input.SessionID, input.Query) {
			duplicateDetected = true
		}
	}

	filter := memory.HybridSearchFilter{
		Limit:  input.Limit,
		Status: "active",
	}

	results, err := globalRepo.HybridSearch(ctx, input.Query, input.Limit, filter)
	if err != nil {
		fmt.Fprintf(w, "Error searching memories: %v\n", err)
		return err
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

		jsonResponse, _ := json.MarshalIndent(response, "", "  ")
		fmt.Fprintf(w, "%s\n", jsonResponse)
		return nil
	}

	maxSimilarity := results[0].VectorScore
	if maxSimilarity < 0.3 {
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

		jsonResponse, _ := json.MarshalIndent(response, "", "  ")
		fmt.Fprintf(w, "%s\n", jsonResponse)
		return nil
	}

	var prioritized []*memory.HybridSearchResult
	var lowPriority []*memory.HybridSearchResult

	if globalSessionTracker != nil {
		recentDeliveries := globalSessionTracker.GetRecentDeliveries(ctx, input.SessionID, 20)

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

	if globalContextPruner != nil {
		prioritized = globalContextPruner.Prune(prioritized, input.Query)
		lowPriority = globalContextPruner.Prune(lowPriority, input.Query)
	}

	response := ActivateContextResponse{
		Query:              input.Query,
		SessionID:          input.SessionID,
		DuplicateDetected:  duplicateDetected,
		MaxTokens:          input.MaxTokens,
		TotalTokens:        0,
		PrimaryResults:     convertToContextItems(prioritized, false, input.Query),
		LowPriorityResults: convertToContextItems(lowPriority, true, input.Query),
	}

	if globalSessionTracker != nil {
		for _, result := range prioritized {
			globalSessionTracker.MarkAsDelivered(ctx, input.SessionID, result.Memory.ID)
		}
		globalSessionTracker.UpdateLastQuery(ctx, input.SessionID, input.Query)
	}

	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Fprintf(w, "Error marshaling response: %v\n", err)
		return err
	}
	fmt.Fprintf(w, "%s\n", jsonResponse)

	return nil
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

func convertToContextItems(results []*memory.HybridSearchResult, isReminder bool, query string) []*ContextResultItem {
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
