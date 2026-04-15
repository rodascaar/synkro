# Synkro

> Intelligent Context Engine for LLMs

![Version](https://img.shields.io/badge/version-2.0-blue)
![Go](https://img.shields.io/badge/Go-1.24+-00ADD8E)
![License](https://img.shields.io/badge/license-MIT-green)

Memory management system with embeddings, relationship graph, and intelligent context pruning.

## Quick Start

### One-Command Installation (Recommended)

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/rodascaar/synkro/main/install.sh | bash

# Windows
irm https://raw.githubusercontent.com/rodascaar/synkro/main/install.ps1 | iex
```

The installer handles platform detection (macOS/Linux/Windows, Intel/ARM), dependency checking, binary download, shell integration, and database initialization.

### Manual Build

```bash
git clone https://github.com/rodascaar/synkro.git
cd synkro
make build
./synkro init
```

### Quick Commands

```bash
./synkro init           # Initialize database
./synkro add            # Add a memory
./synkro list           # List memories
./synkro search <query> # Search memories
./synkro delete <id>    # Delete a memory
./synkro model list     # List embedding models
./synkro tui            # Launch TUI
./synkro mcp            # Start MCP server
```

## Documentation

| Document | Description |
|----------|-------------|
| [INDEX.md](docs/INDEX.md) | Complete documentation index |
| [QUICKSTART.md](docs/QUICKSTART.md) | 5-minute quick start guide |
| [AGENTS.md](docs/AGENTS.md) | AI agent integration guide |
| [INSTALL.md](docs/INSTALL.md) | MCP setup for all IDEs |
| [EMBEDDINGS.md](docs/EMBEDDINGS.md) | Available embedding models |
| [TUI.md](docs/TUI.md) | Complete TUI guide |
| [CHANGELOG.md](CHANGELOG.md) | Changelog |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contributing guide |

## Features

- **FTS5 Full-Text Search** with BM25 scoring
- **Semantic Embeddings** — TF-IDF + N-grams (384 dims) with persistent cache
- **Relationship Graph** — 6 relation types with BFS pathfinding
- **Professional TUI** — 3 interactive panels + Add Memory form
- **MCP Server** — Built with official Go SDK, compatible with all major IDEs
- **SQLite + WAL** — Fast full-text searches
- **Session Tracking** — Persistent deduplication across sessions
- **Context Pruning** — Intelligent result filtering
- **sqlite-vec KNN** — Real vector search (Linux/macOS)
- **CI/CD** — GitHub Actions with lint, tests, and vulnerability scanning

## Platform Notes

| Platform | KNN Vectorial | Notes |
|----------|---------------|-------|
| Linux | Yes (sqlite-vec) | Requires `libsqlite3-dev` |
| macOS | Yes (sqlite-vec) | Uses Xcode CLT SQLite |
| Windows | No | Falls back to in-memory cosine similarity |

On Windows, vector search uses in-memory cosine similarity as a fallback. It is functional but slower with large datasets.

ONNX models (optional) require [ONNX Runtime](https://onnxruntime.ai/) to be installed manually. The TF-IDF generator works without additional dependencies.

## TUI

```bash
./synkro tui
```

**Shortcuts:**
- `Up/Down` or `j/k` — Navigate
- `/` — Search
- `Tab` — Cycle filters (All/Decisions/Tasks/Notes/Archive)
- `g` — Toggle graph view
- `a` — Add memory
- `Ctrl+C` — Quit

Requires a terminal with AltScreen support (min 120x40).

## MCP Integration

Add to your AI IDE configuration:

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

Compatible with: Claude Desktop, Cursor, VS Code, Continue.dev, Cline, Windsurf.

See [INSTALL.md](docs/INSTALL.md) for detailed setup instructions.

## Performance

- **Embeddings**: ~50-100ms (cached: ~1-5ms)
- **Search**: ~10-50ms for 20 results
- **TUI**: <5ms updates
- **Binary**: ~8.7MB

## Environment Variables

```bash
SYNKRO_DB_PATH=memory.db             # Database location (default: memory.db)
SYNKRO_DEBUG=true                     # Enable debug logging
SYNKRO_MAX_TOKENS=4000                 # Default context token limit
SYNKRO_SESSION_BUFFER=20              # Ring buffer size
SYNKRO_CACHE_SIZE=1000                 # Embedding cache size
SYNKRO_SIMILARITY_THRESHOLD=0.5       # Minimum similarity for results
SYNKRO_EMBEDDING_DIM=384              # Embedding dimension
SYNKRO_MODEL_TYPE=tfidf               # Model type: tfidf or onnx
SYNKRO_MODEL_DIR=models                # Model download directory
SYNKRO_PREFERRED_MODEL=all-MiniLM-L6-v2  # Default ONNX model
SYNKRO_AUTO_UPDATE=true               # Enable auto-update check
```

## License

MIT License - see [LICENSE](LICENSE)
