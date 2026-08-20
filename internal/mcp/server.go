package mcp

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rodascaar/synkro/internal/graph"
	"github.com/rodascaar/synkro/internal/memory"
	"github.com/rodascaar/synkro/internal/pruner"
	"github.com/rodascaar/synkro/internal/session"
)

type Server struct {
	repo              *memory.Repository
	graph             *graph.Graph
	sessionTracker    *session.SessionTracker
	contextPruner     *pruner.ContextPruner
	server            *mcp.Server
	serverVersion     string
	embeddingType     string
	conflictThreshold float64
}

const defaultConflictThreshold = 0.7

// ServerOption configures a Server at construction time.
type ServerOption func(*Server)

// WithVersion sets the server version reported in the MCP handshake.
func WithVersion(v string) ServerOption {
	return func(s *Server) { s.serverVersion = v }
}

// WithEmbeddingType sets the embedding model type (tfidf or onnx).
func WithEmbeddingType(t string) ServerOption {
	return func(s *Server) { s.embeddingType = t }
}

// WithConflictThreshold sets the similarity threshold for detecting potential
// conflicts between memories (0.0 to 1.0). Values outside (0, 1] are ignored.
func WithConflictThreshold(t float64) ServerOption {
	return func(s *Server) {
		if t > 0 && t <= 1 {
			s.conflictThreshold = t
		}
	}
}

func NewServer(repo *memory.Repository, g *graph.Graph, st *session.SessionTracker, cp *pruner.ContextPruner, opts ...ServerOption) *Server {
	s := &Server{
		repo:              repo,
		graph:             g,
		sessionTracker:    st,
		contextPruner:     cp,
		serverVersion:     "1.0",
		conflictThreshold: defaultConflictThreshold,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) Run(ctx context.Context) error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "synkro",
		Version: s.serverVersion,
	}, nil)

	s.server = server

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_memory",
		Description: "Add a new memory to Synkro",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Memory type (note, decision, task, context)",
					"enum":        []string{"note", "decision", "task", "context"},
				},
				"title": map[string]interface{}{
					"type":        "string",
					"description": "Memory title (required)",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Memory content",
				},
				"source": map[string]interface{}{
					"type":        "string",
					"description": "Source of the memory",
				},
				"tags": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "Tags for the memory",
				},
			},
			"required": []string{"title"},
		},
	}, s.handleAddMemory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_memory",
		Description: "Get a memory by ID",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "Memory ID (required)",
				},
			},
			"required": []string{"id"},
		},
	}, s.handleGetMemory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_memories",
		Description: "List memories with optional filters",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Filter by memory type",
					"enum":        []string{"note", "decision", "task", "context"},
				},
				"status": map[string]interface{}{
					"type":        "string",
					"description": "Filter by status (active, archived)",
					"enum":        []string{"active", "archived"},
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Maximum number of results",
				},
			},
		},
	}, s.handleListMemory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_memories",
		Description: "Search memories with FTS5 full-text search",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query (required)",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Filter by memory type",
					"enum":        []string{"note", "decision", "task", "context"},
				},
				"status": map[string]interface{}{
					"type":        "string",
					"description": "Filter by status (active, archived)",
					"enum":        []string{"active", "archived"},
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Maximum number of results",
				},
			},
			"required": []string{"query"},
		},
	}, s.handleSearchMemory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_memory",
		Description: "Update an existing memory",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "Memory ID (required)",
				},
				"title": map[string]interface{}{
					"type":        "string",
					"description": "New title",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "New content",
				},
				"status": map[string]interface{}{
					"type":        "string",
					"description": "New status (active, archived)",
					"enum":        []string{"active", "archived"},
				},
				"tags": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "New tags",
				},
			},
			"required": []string{"id"},
		},
	}, s.handleUpdateMemory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "archive_memory",
		Description: "Archive a memory (mark as archived)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "Memory ID (required)",
				},
			},
			"required": []string{"id"},
		},
	}, s.handleArchiveMemory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "activate_context",
		Description: "Activate context with pruning, deduplication, and session tracking",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Context query (required)",
				},
				"session_id": map[string]interface{}{
					"type":        "string",
					"description": "Session ID for tracking",
				},
				"max_tokens": map[string]interface{}{
					"type":        "number",
					"description": "Maximum tokens to return",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Maximum number of results",
				},
			},
			"required": []string{"query"},
		},
	}, s.handleActivateContext)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_relation",
		Description: "Add a relation between two memories (types: extends, depends_on, conflicts_with, example_of, part_of, related_to)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"source_id": map[string]interface{}{
					"type":        "string",
					"description": "Source memory ID (required)",
				},
				"target_id": map[string]interface{}{
					"type":        "string",
					"description": "Target memory ID (required)",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Relation type",
					"enum":        []string{"extends", "depends_on", "conflicts_with", "example_of", "part_of", "related_to"},
				},
				"strength": map[string]interface{}{
					"type":        "number",
					"description": "Relation strength 0.0-1.0 (default: 0.5)",
				},
			},
			"required": []string{"source_id", "target_id", "type"},
		},
	}, s.handleAddRelation)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_relations",
		Description: "Get all relations for a memory",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"memory_id": map[string]interface{}{
					"type":        "string",
					"description": "Memory ID (required)",
				},
			},
			"required": []string{"memory_id"},
		},
	}, s.handleGetRelations)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_relation",
		Description: "Delete a relation between two memories",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"source_id": map[string]interface{}{
					"type":        "string",
					"description": "Source memory ID (required)",
				},
				"target_id": map[string]interface{}{
					"type":        "string",
					"description": "Target memory ID (required)",
				},
			},
			"required": []string{"source_id", "target_id"},
		},
	}, s.handleDeleteRelation)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "find_path",
		Description: "Find a path between two memories using BFS",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"from_id": map[string]interface{}{
					"type":        "string",
					"description": "Source memory ID (required)",
				},
				"to_id": map[string]interface{}{
					"type":        "string",
					"description": "Target memory ID (required)",
				},
			},
			"required": []string{"from_id", "to_id"},
		},
	}, s.handleFindPath)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "pin_memory",
		Description: "Pin a memory so it appears first in listings",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "Memory ID (required)",
				},
			},
			"required": []string{"id"},
		},
	}, s.handlePinMemory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "unpin_memory",
		Description: "Unpin a previously pinned memory",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "Memory ID (required)",
				},
			},
			"required": []string{"id"},
		},
	}, s.handleUnpinMemory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "save_prompt",
		Description: "Save a prompt or query as a context memory",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "The prompt or query text (required)",
				},
				"session_id": map[string]interface{}{
					"type":        "string",
					"description": "Session ID for tracking",
				},
			},
			"required": []string{"prompt"},
		},
	}, s.handleSavePrompt)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "detect_conflicts",
		Description: "Pre-check text against existing memories for potential semantic conflicts",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text": map[string]interface{}{
					"type":        "string",
					"description": "Text to check for conflicts (required)",
				},
				"threshold": map[string]interface{}{
					"type":        "number",
					"description": "Similarity threshold (default: 0.7)",
				},
			},
			"required": []string{"text"},
		},
	}, s.handleDetectConflicts)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "judge_conflict",
		Description: "Resolve a potential conflict between two memories by creating the relation (conflicts_with, supersedes, related_to, part_of, not_conflict)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"memory_id": map[string]interface{}{
					"type":        "string",
					"description": "Memory ID (required)",
				},
				"candidate_id": map[string]interface{}{
					"type":        "string",
					"description": "Candidate memory ID (required)",
				},
				"verdict": map[string]interface{}{
					"type":        "string",
					"description": "Verdict: conflicts_with, supersedes, related_to, part_of, not_conflict (required)",
				},
				"confidence": map[string]interface{}{
					"type":        "number",
					"description": "Confidence 0.0-1.0",
				},
				"reasoning": map[string]interface{}{
					"type":        "string",
					"description": "Reasoning for the verdict",
				},
			},
			"required": []string{"memory_id", "candidate_id", "verdict"},
		},
	}, s.handleJudgeConflict)

	log.SetOutput(os.Stderr)
	log.Printf("Synkro MCP Server v%s starting...\n", s.serverVersion)

	return server.Run(ctx, &mcp.StdioTransport{})
}

func (s *Server) handleAddMemory(ctx context.Context, req *mcp.CallToolRequest, args AddMemoryArgs) (*mcp.CallToolResult, any, error) {
	resultData, err := s.AddMemory(ctx, args)
	if err != nil {
		return errorResult(err), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultData)},
		},
	}, nil, nil
}

func (s *Server) handleGetMemory(ctx context.Context, req *mcp.CallToolRequest, args GetMemoryArgs) (*mcp.CallToolResult, any, error) {
	buf := &BufferWriter{}
	if err := s.GetMemory(ctx, args, buf); err != nil {
		return errorResult(err), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func (s *Server) handleListMemory(ctx context.Context, req *mcp.CallToolRequest, args ListMemoryArgs) (*mcp.CallToolResult, any, error) {
	buf := &BufferWriter{}
	if err := s.ListMemory(ctx, args, buf); err != nil {
		return errorResult(err), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func (s *Server) handleSearchMemory(ctx context.Context, req *mcp.CallToolRequest, args SearchMemoryArgs) (*mcp.CallToolResult, any, error) {
	buf := &BufferWriter{}
	if err := s.SearchMemory(ctx, args, buf); err != nil {
		return errorResult(err), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func (s *Server) handleUpdateMemory(ctx context.Context, req *mcp.CallToolRequest, args UpdateMemoryArgs) (*mcp.CallToolResult, any, error) {
	buf := &BufferWriter{}
	if err := s.UpdateMemory(ctx, args, buf); err != nil {
		return errorResult(err), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func (s *Server) handleArchiveMemory(ctx context.Context, req *mcp.CallToolRequest, args ArchiveMemoryArgs) (*mcp.CallToolResult, any, error) {
	buf := &BufferWriter{}
	if err := s.ArchiveMemory(ctx, args, buf); err != nil {
		return errorResult(err), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func (s *Server) handleActivateContext(ctx context.Context, req *mcp.CallToolRequest, args ActivateContextArgs) (*mcp.CallToolResult, any, error) {
	buf := &BufferWriter{}
	if err := s.ActivateContext(ctx, args, buf); err != nil {
		return errorResult(err), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func (s *Server) handleAddRelation(ctx context.Context, req *mcp.CallToolRequest, args AddRelationArgs) (*mcp.CallToolResult, any, error) {
	buf := &BufferWriter{}
	if err := s.AddRelation(ctx, args, buf); err != nil {
		return errorResult(err), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func (s *Server) handleGetRelations(ctx context.Context, req *mcp.CallToolRequest, args GetRelationsArgs) (*mcp.CallToolResult, any, error) {
	buf := &BufferWriter{}
	if err := s.GetRelations(ctx, args, buf); err != nil {
		return errorResult(err), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func (s *Server) handleDeleteRelation(ctx context.Context, req *mcp.CallToolRequest, args DeleteRelationArgs) (*mcp.CallToolResult, any, error) {
	buf := &BufferWriter{}
	if err := s.DeleteRelation(ctx, args, buf); err != nil {
		return errorResult(err), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func (s *Server) handleFindPath(ctx context.Context, req *mcp.CallToolRequest, args FindPathArgs) (*mcp.CallToolResult, any, error) {
	buf := &BufferWriter{}
	if err := s.FindPath(ctx, args, buf); err != nil {
		return errorResult(err), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func (s *Server) handlePinMemory(ctx context.Context, req *mcp.CallToolRequest, args PinMemoryArgs) (*mcp.CallToolResult, any, error) {
	if args.ID == "" {
		return errorResult(fmt.Errorf("id is required")), nil, nil
	}
	buf := &BufferWriter{}
	if err := s.Pin(ctx, args, buf); err != nil {
		return errorResult(err), nil, nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func (s *Server) handleUnpinMemory(ctx context.Context, req *mcp.CallToolRequest, args PinMemoryArgs) (*mcp.CallToolResult, any, error) {
	if args.ID == "" {
		return errorResult(fmt.Errorf("id is required")), nil, nil
	}
	buf := &BufferWriter{}
	if err := s.Unpin(ctx, args, buf); err != nil {
		return errorResult(err), nil, nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func (s *Server) handleSavePrompt(ctx context.Context, req *mcp.CallToolRequest, args SavePromptArgs) (*mcp.CallToolResult, any, error) {
	if args.Prompt == "" {
		return errorResult(fmt.Errorf("prompt is required")), nil, nil
	}
	buf := &BufferWriter{}
	if err := s.SavePrompt(ctx, args, buf); err != nil {
		return errorResult(err), nil, nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func (s *Server) handleDetectConflicts(ctx context.Context, req *mcp.CallToolRequest, args DetectConflictsArgs) (*mcp.CallToolResult, any, error) {
	if args.Text == "" {
		return errorResult(fmt.Errorf("text is required")), nil, nil
	}
	buf := &BufferWriter{}
	if err := s.DetectConflicts(ctx, args, buf); err != nil {
		return errorResult(err), nil, nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func (s *Server) handleJudgeConflict(ctx context.Context, req *mcp.CallToolRequest, args JudgeConflictArgs) (*mcp.CallToolResult, any, error) {
	if args.MemoryID == "" || args.CandidateID == "" || args.Verdict == "" {
		return errorResult(fmt.Errorf("memory_id, candidate_id and verdict are required")), nil, nil
	}
	buf := &BufferWriter{}
	if err := s.JudgeConflict(ctx, args, buf); err != nil {
		return errorResult(err), nil, nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

func errorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: err.Error()},
		},
		IsError: true,
	}
}

type AddMemoryArgs struct {
	Type     string   `json:"type" jsonschema:"Memory type (note, decision, task, context)"`
	Title    string   `json:"title" jsonschema:"Memory title (required)"`
	Content  string   `json:"content" jsonschema:"Memory content"`
	Source   string   `json:"source" jsonschema:"Source of the memory"`
	Tags     []string `json:"tags" jsonschema:"Tags for the memory"`
	TopicKey string   `json:"topic_key" jsonschema:"Optional topic key; updates existing memory with same key instead of creating a duplicate"`
}

type GetMemoryArgs struct {
	ID string `json:"id" jsonschema:"Memory ID (required)"`
}

type ListMemoryArgs struct {
	Type   string `json:"type" jsonschema:"Filter by memory type"`
	Status string `json:"status" jsonschema:"Filter by status (active, archived)"`
	Limit  int    `json:"limit" jsonschema:"Maximum number of results"`
}

type SearchMemoryArgs struct {
	Query  string `json:"query" jsonschema:"Search query (required)"`
	Type   string `json:"type" jsonschema:"Filter by memory type"`
	Status string `json:"status" jsonschema:"Filter by status (active, archived)"`
	Limit  int    `json:"limit" jsonschema:"Maximum number of results"`
}

type UpdateMemoryArgs struct {
	ID      string   `json:"id" jsonschema:"Memory ID (required)"`
	Title   string   `json:"title" jsonschema:"New title"`
	Content string   `json:"content" jsonschema:"New content"`
	Status  string   `json:"status" jsonschema:"New status (active, archived)"`
	Tags    []string `json:"tags" jsonschema:"New tags"`
}

type ArchiveMemoryArgs struct {
	ID string `json:"id" jsonschema:"Memory ID (required)"`
}

type ActivateContextArgs struct {
	Query     string `json:"query" jsonschema:"Context query (required)"`
	SessionID string `json:"session_id" jsonschema:"Session ID for tracking"`
	MaxTokens int    `json:"max_tokens" jsonschema:"Maximum tokens to return"`
	Limit     int    `json:"limit" jsonschema:"Maximum number of results"`
}

type AddRelationArgs struct {
	SourceID string  `json:"source_id" jsonschema:"Source memory ID (required)"`
	TargetID string  `json:"target_id" jsonschema:"Target memory ID (required)"`
	Type     string  `json:"type" jsonschema:"Relation type"`
	Strength float64 `json:"strength" jsonschema:"Relation strength 0.0-1.0"`
}

type GetRelationsArgs struct {
	MemoryID string `json:"memory_id" jsonschema:"Memory ID (required)"`
}

type DeleteRelationArgs struct {
	SourceID string `json:"source_id" jsonschema:"Source memory ID (required)"`
	TargetID string `json:"target_id" jsonschema:"Target memory ID (required)"`
}

type FindPathArgs struct {
	FromID string `json:"from_id" jsonschema:"Source memory ID (required)"`
	ToID   string `json:"to_id" jsonschema:"Target memory ID (required)"`
}

type PinMemoryArgs struct {
	ID string `json:"id" jsonschema:"Memory ID (required)"`
}

type SavePromptArgs struct {
	Prompt    string `json:"prompt" jsonschema:"The prompt or query text (required)"`
	SessionID string `json:"session_id" jsonschema:"Session ID for tracking"`
}

type DetectConflictsArgs struct {
	Text      string  `json:"text" jsonschema:"Text to check for conflicts (required)"`
	Threshold float64 `json:"threshold" jsonschema:"Similarity threshold (default: 0.7)"`
}

type JudgeConflictArgs struct {
	MemoryID    string  `json:"memory_id" jsonschema:"Memory ID (required)"`
	CandidateID string  `json:"candidate_id" jsonschema:"Candidate memory ID (required)"`
	Verdict     string  `json:"verdict" jsonschema:"Verdict: conflicts_with, supersedes, related_to, part_of, not_conflict (required)"`
	Confidence  float64 `json:"confidence" jsonschema:"Confidence 0.0-1.0"`
	Reasoning   string  `json:"reasoning" jsonschema:"Reasoning for the verdict"`
}

type BufferWriter struct {
	buf []byte
}

func (b *BufferWriter) Write(p []byte) (n int, err error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *BufferWriter) String() string {
	return string(b.buf)
}

func (b *BufferWriter) Reset() {
	b.buf = nil
}
