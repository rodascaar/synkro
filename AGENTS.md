# Synkro - Agent Guide

## Quick Commands

```bash
# Build
go build -o synkro ./cmd/synkro/

# Initialize database (required before any operations)
./synkro init

# CLI operations
./synkro add --title "Test" --content "Content" --type note
./synkro list --limit 10
./synkro search "query"

# Model management (NEW)
./synkro model list                  # List available embedding models
./synkro model info all-MiniLM-L6-v2 # Get model details
./synkro model download all-MiniLM-L6-v2    # Download specific model
./synkro model delete all-MiniLM-L6-v2 # Delete downloaded model

# Run TUI
./synkro tui

# Run MCP Server
./synkro mcp

# Full test suite
make ci
```

## Project Structure

```
cmd/synkro/          # CLI entrypoint (main.go, commands.go)
internal/            # All internal packages
  ├── config/       # Environment configuration (SYNKRO_DB_PATH, etc.)
  ├── db/          # SQLite with FTS5 virtual tables + WAL mode + FK
  ├── memory/      # Memory models and repository (MATCH queries, BM25)
  ├── embeddings/  # TF-IDF + N-gram embeddings (384 dims) with persistent cache
  ├── graph/       # Memory relationship graph (BFS) with repository
  ├── mcp/         # MCP Server using Go SDK official
  ├── pruner/      # Context pruning logic
  ├── session/     # Session tracking (in-memory + SQLite persistent)
  └── tui/         # Bubble Tea TUI (3 panels + Add Memory form)
memory.db           # SQLite database (created by init)
```

## Key Technical Details

- **Go version**: 1.25 (upgraded for MCP SDK)
- **CLI framework**: Cobra
- **TUI framework**: Bubble Tea + Lipgloss
- **Database**: SQLite3 with FTS5 virtual tables, WAL mode, foreign keys
- **Search**: FTS5 full-text with BM25 scoring + hybrid search (vectorial + FTS5)
- **Embeddings**: 384-dim vectors using TF-IDF + N-grams with stopwords filtering, persistent cache in SQLite
- **ONNX Models**: Support for high-quality sentence-transformers models (all-MiniLM-L6-v2, paraphrase-multilingual-MiniLM-L12-v2, stsb-roberta-base-v2) with automatic download from Hugging Face
- **MCP Server**: Fully implemented using github.com/modelcontextprotocol/go-sdk
- **Testing**: Comprehensive test suite with >90% coverage
- **Linting**: golangci-lint configured with essential linters

## Architecture

### Entry Points
- `cmd/synkro/main.go` - minimal bootstrap
- `cmd/synkro/commands.go` - Cobra commands (init, add, list, search, tui, mcp)

### Core Components
- **MemoryRepository** (`internal/memory/repository.go`) - CRUD + FTS5 search + hybrid search
- **TFIDFEmbeddingGenerator** (`internal/embeddings/generator.go`) - generates 384-dim vectors with persistent cache
- **EmbeddingCache** (`internal/embeddings/cache.go`) - SQLite-backed cache with SHA256 hash keys
- **Graph** (`internal/graph/graph.go`) - BFS-based relation finding with repository
- **GraphRepository** (`internal/graph/repository.go`) - Persistencia de relaciones en SQLite
- **SessionTracker** (`internal/session/tracker.go`) - dual storage: in-memory + SQLite persistent
- **SessionRepository** (`internal/session/repository.go`) - SQLite repository for sessions
- **MCPServer** (`internal/mcp/server.go`) - Full MCP Server using Go SDK official

### Database Schema
- `memories` - main memory table
- `memories_fts` - FTS5 virtual table for full-text search
- `memory_embeddings` - vector storage (BLOB)
- `embedding_cache` - persistent embedding cache (SQLite)
- `memory_relations` - graph edges (6 types: related_to, part_of, extends, conflicts_with, example_of, depends_on)
- `sessions` - session tracking (persisted)
- `session_memories` - delivered memories per session (persisted)

### Memory Types
- `note` - general notes
- `decision` - architectural/technical decisions
- `task` - actionable tasks
- `context` - contextual information

## Environment Variables

```bash
SYNKRO_DB_PATH=memory.db           # Database location (default: memory.db)
SYNKRO_DEBUG=true                   # Enable debug logging (default: false)
SYNKRO_MAX_TOKENS=4000             # Default context token limit (default: 4000)
SYNKRO_SESSION_BUFFER=20           # Ring buffer size (default: 20)
SYNKRO_CACHE_SIZE=1000             # Embedding cache size (default: 1000)
SYNKRO_SIMILARITY_THRESHOLD=0.5     # Minimum similarity for results (default: 0.5)
SYNKRO_EMBEDDING_DIM=384           # Embedding dimension (default: 384)
SYNKRO_MODEL_TYPE=tfidf             # Model type (default: tfidf) or onnx
SYNKRO_MODEL_DIR=models             # Model download directory (default: models)
SYNKRO_PREFERRED_MODEL=all-MiniLM-L6-v2  # Default ONNX model to use
```

## Verification & Testing

```bash
# Rebuild after changes
go build -o synkro ./cmd/synkro/

# Test basic CRUD
./synkro init                    # Required first
./synkro add --title "Test" --content "Test content" --type note
./synkro list
./synkro search "Test"

# Test TUI (requires terminal with AltScreen support, min 120x40)
./synkro tui
# Navigate: ↑/↓ or j/k
# Search: /
# Toggle graph: g
# Add memory: a
# Quit: Ctrl+C

# Test MCP Server
./synkro mcp  # Starts MCP server using Go SDK

# Run tests
make test         # Full test suite with coverage
make test-short   # Quick tests
make bench        # Benchmarks

# Run linter
make lint
```

## MCP Server

**FULLY IMPLEMENTED** using `github.com/modelcontextprotocol/go-sdk` v1.5.0

Available tools:
- `add_memory` - Add new memory
- `get_memory` - Get memory by ID
- `list_memories` - List memories with filters
- `search_memories` - FTS5 full-text search
- `update_memory` - Update existing memory
- `archive_memory` - Archive memory (mark as archived)
- `activate_context` - Activate context with pruning and deduplication

Configuration for clients:
```json
{
  "mcp_servers": {
    "synkro": {
      "command": "synkro",
      "args": ["mcp"]
    }
  }
}
```

## TUI Add Memory

**FULLY IMPLEMENTED** - Press `a` in TUI to open add memory form with 4 fields:
1. Type (note/decision/task/context)
2. Title (required)
3. Content (detailed information)
4. Tags (comma separated, optional)

Controls:
- Enter: Next field / Save
- Tab: Next field
- Shift+Tab: Previous field
- Esc: Cancel

## Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/spf13/cobra` - CLI
- `github.com/mattn/go-sqlite3` - SQLite driver (CGO)
- `github.com/google/uuid` - UUID generation
- `github.com/modelcontextprotocol/go-sdk` - MCP Go SDK official
- `github.com/stretchr/testify` - Testing framework

## When Working On This Codebase

- Adding CLI commands: Edit `cmd/synkro/commands.go`
- Modifying database schema: Update `internal/db/db.go`
- Changing embeddings: Edit `internal/embeddings/generator.go`
- TUI changes: Edit `internal/tui/model.go`, `internal/tui/tui.go`, `internal/tui/add_model.go`
- MCP Server: Edit `internal/mcp/server.go` (uses Go SDK)
- Adding tests: Create `*_test.go` files in respective packages
- Linting: Update `.golangci.yml`

## Important Gotchas

1. **Database init**: Must run `./synkro init` before any other operations
2. **Go version**: Requires Go 1.25+ for MCP SDK (automatically upgraded by go get)
3. **FTS5 implemented**: Full-text search uses FTS5 virtual tables with BM25 scoring
4. **Session tracking**: Dual storage (in-memory + SQLite) - persisted across restarts
5. **Embedding cache**: SQLite-backed cache with SHA256 hash keys - persisted across restarts
6. **Graph relations**: Fully persisted in SQLite with repository pattern
7. **TUI terminal**: Requires AltScreen support - set `TERM=xterm-256color` if issues
8. **MCP functional**: Full implementation using Go SDK official, not just a stub
9. **Environment variables**: Fully functional, loaded via `internal/config/config.go`
10. **Tests available**: Comprehensive test suite in `*_test.go` files
11. **Linting configured**: `.golangci.yml` with essential linters
12. **CI/CD configured**: GitHub Actions workflow in `.github/workflows/test.yml`

