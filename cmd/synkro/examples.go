package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rodascaar/synkro/internal/db"
	"github.com/rodascaar/synkro/internal/embeddings"
	"github.com/rodascaar/synkro/internal/memory"
	"github.com/spf13/cobra"
)

var examplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Load example memories to get started",
	Long:  "Load pre-configured example memories to demonstrate Synkro's capabilities",
	Run: func(cmd *cobra.Command, args []string) {
		examples := []struct {
			memType string
			title   string
			content string
			source  string
			tags    []string
		}{
			{
				memType: "decision",
				title:   "Use Go 1.25+ for Synkro",
				content: "Synkro requires Go 1.25 or later to support MCP SDK. We chose this version for better compatibility across different platforms and CI/CD environments.",
				source:  "architecture",
				tags:    []string{"go", "version", "architecture"},
			},
			{
				memType: "decision",
				title:   "Implement MCP Server with Official SDK",
				content: "Synkro uses github.com/modelcontextprotocol/go-sdk for MCP integration. This ensures full compatibility with Claude Desktop, opencode, Cursor, VS Code, and other MCP-compatible clients.",
				source:  "architecture",
				tags:    []string{"mcp", "sdk", "integration"},
			},
			{
				memType: "decision",
				title:   "Use SQLite with FTS5 for Search",
				content: "Full-text search uses SQLite FTS5 virtual tables with BM25 scoring. This provides fast, efficient text search with relevance ranking. Requires compilation with sqlite_fts5 tag.",
				source:  "architecture",
				tags:    []string{"sqlite", "fts5", "search"},
			},
			{
				memType: "note",
				title:   "Bubble Tea for Professional TUI",
				content: "Synkro's TUI is built with Bubble Tea and Lipgloss for a modern, professional terminal experience. Features include AltScreen mode, 3-panel layout, real-time search, and Vi-style navigation.",
				source:  "architecture",
				tags:    []string{"tui", "bubbletea", "ui"},
			},
			{
				memType: "note",
				title:   "TF-IDF + N-grams for Embeddings",
				content: "Semantic search uses TF-IDF (Term Frequency-Inverse Document Frequency) combined with N-grams to generate 384-dimensional vectors. This captures both individual words and phrases for better semantic understanding.",
				source:  "architecture",
				tags:    []string{"embeddings", "tfidf", "search"},
			},
			{
				memType: "note",
				title:   "Graph Relationships Between Memories",
				content: "Memories can be connected with 6 relationship types: related_to, part_of, extends, conflicts_with, example_of, depends_on. The graph uses BFS traversal for finding connected memories.",
				source:  "architecture",
				tags:    []string{"graph", "relationships", "connections"},
			},
			{
				memType: "note",
				title:   "Session Tracking for Context",
				content: "Session tracking uses a ring buffer (default: 20 memories) to prevent repeated information in the same session. Includes dual storage: in-memory for speed and SQLite for persistence across restarts.",
				source:  "architecture",
				tags:    []string{"session", "tracking", "context"},
			},
			{
				memType: "note",
				title:   "Context Pruning for Quality",
				content: "Context pruning filters results by similarity threshold (default: 0.5), removes repetitive content, and eliminates stop-words. This ensures high-quality, relevant context for LLMs.",
				source:  "architecture",
				tags:    []string{"pruning", "context", "quality"},
			},
			{
				memType: "task",
				title:   "Add MCP Configuration to IDE",
				content: "To use Synkro with Claude Desktop, opencode, Cursor, or VS Code:\n\n1. Find MCP configuration file for your IDE\n2. Add synkro server configuration\n3. Restart IDE\n\nConfiguration format:\n{ \"mcp_servers\": { \"synkro\": { \"command\": \"synkro\", \"args\": [\"mcp\"] } } }",
				source:  "setup",
				tags:    []string{"mcp", "configuration", "ide"},
			},
			{
				memType: "note",
				title:   "Common Commands Quick Reference",
				content: "Essential Synkro commands:\n\n• synkro init - Initialize database (first time only)\n• synkro add --title \"Title\" --content \"Content\" --type note\n• synkro list --limit 10 - List memories\n• synkro search \"query\" - Search memories\n• synkro tui - Launch TUI\n• synkro mcp - Start MCP server\n• synkro update - Check for updates\n• synkro version - Show version info",
				source:  "reference",
				tags:    []string{"cli", "commands", "reference"},
			},
		}

		d, err := db.New(cfg.DatabasePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = d.Close() }()

		repo := memory.NewRepository(d.DB())

		embedMgr, err := embeddings.NewEmbeddingManager(embeddings.Config{
			ModelType: embeddings.ModelTypeTFIDF,
			DB:        d.DB(),
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating embedding manager: %v\n", err)
			os.Exit(1)
		}
		repo.SetEmbeddingGenerator(embedMgr)

		fmt.Println("📦 Loading example memories...")
		for i, ex := range examples {
			mem := &memory.Memory{
				Type:    ex.memType,
				Title:   ex.title,
				Content: ex.content,
				Source:  ex.source,
				Status:  "active",
				Tags:    ex.tags,
			}
			if err := repo.Create(context.Background(), mem); err != nil {
				fmt.Printf("❌ Failed to add example %d: %v\n", i+1, err)
			} else {
				fmt.Printf("✅ Added example: %s\n", ex.title)
			}
		}
		fmt.Printf("\n✅ Loaded %d example memories!\n", len(examples))
		fmt.Println("\nNext steps:")
		fmt.Println("  synkro list      - View all memories")
		fmt.Println("  synkro tui       - Explore with TUI")
		fmt.Println("  synkro mcp       - Start MCP server")
	},
}
