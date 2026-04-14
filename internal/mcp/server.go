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
}

func NewServer(repo *memory.Repository, g *graph.Graph, st *session.SessionTracker, cp *pruner.ContextPruner) *Server {
	return &Server{
		repo:           repo,
		graph:          g,
		sessionTracker: st,
		contextPruner:  cp,
	}
}

func (s *Server) Run(ctx context.Context) error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "synkro",
		Version: "2.0.0",
	}, nil)

	s.server = server

	SetGlobalRepo(s.repo)
	SetGraph(s.graph)
	SetSessionTracker(s.sessionTracker)
	SetContextPruner(s.contextPruner)

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

	log.SetOutput(os.Stderr)
	log.Println("Synkro MCP Server v2.0.0 starting...")

	return server.Run(ctx, &mcp.StdioTransport{})
}

type AddMemoryArgs struct {
	Type    string   `json:"type" jsonschema:"Memory type (note, decision, task, context)"`
	Title   string   `json:"title" jsonschema:"Memory title (required)"`
	Content string   `json:"content" jsonschema:"Memory content"`
	Source  string   `json:"source" jsonschema:"Source of the memory"`
	Tags    []string `json:"tags" jsonschema:"Tags for the memory"`
}

func (s *Server) handleAddMemory(ctx context.Context, req *mcp.CallToolRequest, args AddMemoryArgs) (*mcp.CallToolResult, any, error) {
	input := AddMemoryInput{
		Type:    args.Type,
		Title:   args.Title,
		Content: args.Content,
		Source:  args.Source,
		Tags:    args.Tags,
	}

	if input.Title == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: title is required"},
			},
			IsError: true,
		}, nil, nil
	}

	resultData, err := AddMemory(input)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: " + err.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultData)},
		},
	}, nil, nil
}

type GetMemoryArgs struct {
	ID string `json:"id" jsonschema:"Memory ID (required)"`
}

func (s *Server) handleGetMemory(ctx context.Context, req *mcp.CallToolRequest, args GetMemoryArgs) (*mcp.CallToolResult, any, error) {
	input := GetMemoryInput{ID: args.ID}
	buf := &bufferWriter{}
	if err := GetMemory(input, buf); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: " + err.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

type ListMemoryArgs struct {
	Type   string `json:"type" jsonschema:"Filter by memory type"`
	Status string `json:"status" jsonschema:"Filter by status (active, archived)"`
	Limit  int    `json:"limit" jsonschema:"Maximum number of results"`
}

func (s *Server) handleListMemory(ctx context.Context, req *mcp.CallToolRequest, args ListMemoryArgs) (*mcp.CallToolResult, any, error) {
	input := ListMemoryInput{
		Type:   args.Type,
		Status: args.Status,
		Limit:  args.Limit,
	}
	buf := &bufferWriter{}
	if err := ListMemory(input, buf); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: " + err.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

type SearchMemoryArgs struct {
	Query  string `json:"query" jsonschema:"Search query (required)"`
	Type   string `json:"type" jsonschema:"Filter by memory type"`
	Status string `json:"status" jsonschema:"Filter by status (active, archived)"`
	Limit  int    `json:"limit" jsonschema:"Maximum number of results"`
}

func (s *Server) handleSearchMemory(ctx context.Context, req *mcp.CallToolRequest, args SearchMemoryArgs) (*mcp.CallToolResult, any, error) {
	input := SearchMemoryInput{
		Query:  args.Query,
		Type:   args.Type,
		Status: args.Status,
		Limit:  args.Limit,
	}

	if input.Query == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: query is required"},
			},
			IsError: true,
		}, nil, nil
	}

	buf := &bufferWriter{}
	if err := SearchMemory(input, buf); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: " + err.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

type UpdateMemoryArgs struct {
	ID      string   `json:"id" jsonschema:"Memory ID (required)"`
	Title   string   `json:"title" jsonschema:"New title"`
	Content string   `json:"content" jsonschema:"New content"`
	Status  string   `json:"status" jsonschema:"New status (active, archived)"`
	Tags    []string `json:"tags" jsonschema:"New tags"`
}

func (s *Server) handleUpdateMemory(ctx context.Context, req *mcp.CallToolRequest, args UpdateMemoryArgs) (*mcp.CallToolResult, any, error) {
	input := UpdateMemoryInput{
		ID:      args.ID,
		Title:   args.Title,
		Content: args.Content,
		Status:  args.Status,
		Tags:    args.Tags,
	}

	if input.ID == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: id is required"},
			},
			IsError: true,
		}, nil, nil
	}

	buf := &bufferWriter{}
	if err := UpdateMemory(input, buf); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: " + err.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

type ArchiveMemoryArgs struct {
	ID string `json:"id" jsonschema:"Memory ID (required)"`
}

func (s *Server) handleArchiveMemory(ctx context.Context, req *mcp.CallToolRequest, args ArchiveMemoryArgs) (*mcp.CallToolResult, any, error) {
	input := ArchiveMemoryInput{ID: args.ID}

	if input.ID == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: id is required"},
			},
			IsError: true,
		}, nil, nil
	}

	buf := &bufferWriter{}
	if err := ArchiveMemory(input, buf); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: " + err.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

type ActivateContextArgs struct {
	Query     string `json:"query" jsonschema:"Context query (required)"`
	SessionID string `json:"session_id" jsonschema:"Session ID for tracking"`
	MaxTokens int    `json:"max_tokens" jsonschema:"Maximum tokens to return"`
	Limit     int    `json:"limit" jsonschema:"Maximum number of results"`
}

func (s *Server) handleActivateContext(ctx context.Context, req *mcp.CallToolRequest, args ActivateContextArgs) (*mcp.CallToolResult, any, error) {
	input := ActivateContextInput{
		Query:     args.Query,
		SessionID: args.SessionID,
		MaxTokens: args.MaxTokens,
		Limit:     args.Limit,
	}

	if input.Query == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: query is required"},
			},
			IsError: true,
		}, nil, nil
	}

	buf := &bufferWriter{}
	if err := ActivateContext(input, buf); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: " + err.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: buf.String()},
		},
	}, nil, nil
}

type bufferWriter struct {
	buf []byte
}

func (b *bufferWriter) Write(p []byte) (n int, err error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *bufferWriter) String() string {
	return string(b.buf)
}
