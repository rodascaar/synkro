# Synkro

> Intelligent Context Engine for LLMs

![Version](https://img.shields.io/badge/version-2.3-blue)
![Go](https://img.shields.io/badge/Go-1.26+-00ADD8E)
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
- **Hybrid Search** — FTS5 BM25 + in-memory cosine similarity reranking
- **Relationship Graph** — 7 relation types (incl. `supersedes`) with BFS pathfinding
- **Professional TUI** — 3 interactive panels + Add Memory form
- **MCP Server** — Built with official Go SDK, compatible with all major IDEs
- **Conflict Detection** — Pre-check + verdict resolution (`detect_conflicts`, `judge_conflict`)
- **SQLite + WAL** — Pure-Go driver (`modernc.org/sqlite`), no CGO required
- **Session Tracking** — Persistent deduplication across sessions
- **Context Pruning** — Intelligent result filtering
- **CI/CD** — GitHub Actions with lint, tests, and vulnerability scanning

## Platform Notes

Synkro uses a pure-Go SQLite driver (`modernc.org/sqlite`) with FTS5 built in, so it
builds and runs identically on Linux, macOS, and Windows with no native dependencies.
Vector search uses in-memory cosine similarity over all embeddings, so no external
vector extension is required.

### Embedding Models (Optional)

Synkro includes a TF-IDF generator that works out of the box. For higher quality semantic search, install ONNX Runtime and download a sentence-transformers model:

```bash
# 1. Install ONNX Runtime
brew install onnxruntime    # macOS
apt install libonnxruntime-dev  # Linux

# 2. Initialize with model auto-download (~90 MB)
./synkro init --with-models

# Or download manually
./synkro model download all-MiniLM-L6-v2
```

`init --with-models` automatically downloads the model, detects ONNX Runtime availability, and configures Synkro to use it. If ONNX Runtime is not installed, it prints installation instructions and falls back to TF-IDF.

Set `SYNKRO_MODEL_TYPE=onnx` in environment variables to use ONNX embeddings. Without it, Synkro defaults to TF-IDF.

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
- **Binary**: ~19MB (pure-Go, no native dependencies)

## Environment Variables

```bash
SYNKRO_DB_PATH=memory.db                 # Database location (default: memory.db)
SYNKRO_MODEL_TYPE=tfidf                  # Model type: tfidf or onnx
SYNKRO_MODEL_DIR=models                  # Model download directory
SYNKRO_PREFERRED_MODEL=all-MiniLM-L6-v2  # Default ONNX model
SYNKRO_CONFLICT_THRESHOLD=0.7            # Similarity threshold for conflict detection
```

## License

MIT License - see [LICENSE](LICENSE)
