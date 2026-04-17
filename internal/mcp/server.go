package mcp

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rodascaar/synkro/internal/graph"
	"github.com/rodascaar/synkro/internal/memory"
	"github.com/rodascaar/synkro/internal/pruner"
	"github.com/rodascaar/synkro/internal/session"
)

type Server struct {
	repo           *memory.Repository
	graph          *graph.Graph
	sessionTracker *session.SessionTracker
	contextPruner  *pruner.ContextPruner
	server         *mcp.Server
	serverVersion  string
	embeddingType  string
}

func NewServer(repo *memory.Repository, g *graph.Graph, st *session.SessionTracker, cp *pruner.ContextPruner) *Server {
	return &Server{
		repo:           repo,
		graph:          g,
		sessionTracker: st,
		contextPruner:  cp,
		serverVersion:  "1.0",
	}
}

func (s *Server) SetVersion(v string) {
	s.serverVersion = v
}

func (s *Server) SetEmbeddingType(t string) {
	s.embeddingType = t
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

	log.SetOutput(os.Stderr)
	log.Printf("Synkro MCP Server v%s starting...\n", s.serverVersion)

	return server.Run(ctx, &mcp.StdioTransport{})
}

func (s *Server) handleAddMemory(ctx context.Context, req *mcp.CallToolRequest, args AddMemoryArgs) (*mcp.CallToolResult, any, error) {
	input := AddMemoryInput{
		Type:    args.Type,
		Title:   args.Title,
		Content: args.Content,
		Source:  args.Source,
		Tags:    args.Tags,
	}

	resultData, err := s.AddMemory(ctx, input)
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
	if err := s.GetMemory(ctx, GetMemoryInput{ID: args.ID}, buf); err != nil {
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
	if err := s.ListMemory(ctx, ListMemoryInput{
		Type:   args.Type,
		Status: args.Status,
		Limit:  args.Limit,
	}, buf); err != nil {
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
	if err := s.SearchMemory(ctx, SearchMemoryInput{
		Query:  args.Query,
		Type:   args.Type,
		Status: args.Status,
		Limit:  args.Limit,
	}, buf); err != nil {
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
	if err := s.UpdateMemory(ctx, UpdateMemoryInput{
		ID:      args.ID,
		Title:   args.Title,
		Content: args.Content,
		Status:  args.Status,
		Tags:    args.Tags,
	}, buf); err != nil {
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
	if err := s.ArchiveMemory(ctx, ArchiveMemoryInput{ID: args.ID}, buf); err != nil {
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
	if err := s.ActivateContext(ctx, ActivateContextInput{
		Query:     args.Query,
		SessionID: args.SessionID,
		MaxTokens: args.MaxTokens,
		Limit:     args.Limit,
	}, buf); err != nil {
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
	if err := s.AddRelation(ctx, AddRelationInput{
		SourceID: args.SourceID,
		TargetID: args.TargetID,
		Type:     args.Type,
		Strength: args.Strength,
	}, buf); err != nil {
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
	if err := s.GetRelations(ctx, GetRelationsInput{MemoryID: args.MemoryID}, buf); err != nil {
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
	if err := s.DeleteRelation(ctx, DeleteRelationInput{
		SourceID: args.SourceID,
		TargetID: args.TargetID,
	}, buf); err != nil {
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
	if err := s.FindPath(ctx, FindPathInput{
		FromID: args.FromID,
		ToID:   args.ToID,
	}, buf); err != nil {
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
	Type    string   `json:"type" jsonschema:"Memory type (note, decision, task, context)"`
	Title   string   `json:"title" jsonschema:"Memory title (required)"`
	Content string   `json:"content" jsonschema:"Memory content"`
	Source  string   `json:"source" jsonschema:"Source of the memory"`
	Tags    []string `json:"tags" jsonschema:"Tags for the memory"`
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
