# SYNKRO - Quick Installation Guide

**Professional One-Command Installation** - Install Synkro with a single command!

## 🚀 Quick Start (Recommended)

### macOS / Linux

```bash
# One-command installation
curl -fsSL https://raw.githubusercontent.com/rodascaar/synkro/main/install.sh | bash
```

### Windows

```powershell
# One-command installation
irm https://raw.githubusercontent.com/rodascaar/synkro/main/install.ps1 | iex
```

## What Gets Installed

✅ **Synkro Binary** - Ready-to-use CLI tool
✅ **Directory Structure** - Organized data, models, and config
✅ **Shell Integration** - Auto-configured for your shell
✅ **Database** - SQLite initialized and ready
✅ **Optional Embeddings** - Semantic search models (optional download)

## Installation Process

The professional installer handles everything:

1. 📦 **Dependency Check** - Verifies Go, disk space, and system requirements
2. 🖥️ **Platform Detection** - Automatically detects OS and architecture
3. 📁 **Directory Setup** - Creates organized directory structure
4. ⬇️ **Binary Installation** - Downloads or builds the binary
5. ⚙️ **Environment Setup** - Configures shell integration
6. 💾 **Database Init** - Initializes SQLite database
7. 🤖 **Optional Models** - Downloads embedding models (optional)

## Manual Installation (Optional)

If automatic installation doesn't work:

### Prerequisites

- **Go** 1.24+ (for building from source)
- **CGO** enabled (for SQLite support)
- **500MB** free disk space

### Build from Source

```bash
# Clone repository
git clone https://github.com/rodascaar/synkro.git
cd synkro

# Build binary
make build
# or: go build -o synkro ./cmd/synkro/

# Install locally
make install
```

## Getting Started

After installation, Synkro is ready to use:

```bash
# 1. Reload your shell (automatic installer users)
source ~/.synkro/synkro.sh

# 2. Initialize database
synkro init

# 3. Add your first memory
synkro add --title "My project" --content "Project details..." --type note

# 4. Launch TUI
synkro tui

# 5. Start MCP server
synkro mcp
```

## Configuration

Environment variables (optional, set in `~/.synkro/config/synkro.sh`):

```bash
# Database location
export SYNKRO_DB_PATH="$HOME/.synkro/data/memory.db"

# Model type (tfidf or onnx)
export SYNKRO_MODEL_TYPE="onnx"

# Preferred model
export SYNKRO_PREFERRED_MODEL="all-MiniLM-L6-v2"

# Model directory
export SYNKRO_MODEL_DIR="$HOME/.synkro/models"

# Enable debug logging
export SYNKRO_DEBUG="true"

# Token limit for context
export SYNKRO_MAX_TOKENS=4000

# Similarity threshold
export SYNKRO_SIMILARITY_THRESHOLD=0.5
```

## Embedding Models

Synkro supports semantic embeddings for better search accuracy:

### Available Models

```bash
# List all available models
synkro model list

# Get detailed information
synkro model info all-MiniLM-L6-v2

# Download a specific model
synkro model download all-MiniLM-L6-v2

# Delete a downloaded model
synkro model delete all-MiniLM-L6-v2
```

### Model Options

| Model | Dimensions | Size | Language | Speed | Accuracy |
|-------|-----------|------|----------|-------|----------|
| all-MiniLM-L6-v2 | 384 | 22.7M | English | Fastest | Good |
| paraphrase-multilingual-MiniLM-L12-v2 | 384 | 118M | 50+ | Medium | Good |
| stsb-roberta-base-v2 | 768 | 110M | English | Slow | Best |

## Development Setup

### Building from Source

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Create distribution packages
make package

# Run tests
make test

# Run linter
make lint
```

### Make Targets

```bash
make all              # Build for current platform (default)
make build            # Build binary
make build-all        # Build for all platforms
make clean            # Remove build artifacts
make test             # Run tests
make test-coverage    # Run tests with coverage
make lint             # Run linter
make fmt              # Format code
make package          # Create distribution packages
make install          # Install to ~/.synkro
make checksums        # Generate SHA256 checksums
make release          # Create GitHub release
```

## Troubleshooting

### Database Issues

```bash
# Reinitialize database
rm -f ~/.synkro/data/memory.db
synkro init
```

### Model Issues

```bash
# Check downloaded models
synkro model list

# Redownload model
synkro model delete <model-name>
synkro model download <model-name>
```

### Path Issues

```bash
# Check if Synkro is in PATH
which synkro

# Manually add to PATH
export PATH="$HOME/.synkro/bin:$PATH"

# Add to shell config (for zsh)
echo 'export PATH="$HOME/.synkro/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### Shell Integration Issues

```bash
# For bash
# Edit ~/.bashrc and add:
# source ~/.synkro/synkro.sh

# For zsh
# Edit ~/.zshrc and add:
# source ~/.synkro/synkro.sh

# For fish
# Edit ~/.config/fish/config.fish and add:
# source ~/.synkro/synkro/config/synkro.fish
```

## Uninstallation

```bash
# Complete removal
rm -rf ~/.synkro

# Remove shell integration
# Remove the line that sources ~/.synkro/synkro.sh from your shell config
```

## Support and Documentation

- **📚 Wiki**: https://github.com/rodascaar/synkro/wiki
- **🐛 Issues**: https://github.com/rodascaar/synkro/issues
- **💬 Discussions**: https://github.com/rodascaar/synkro/discussions
- **📧 Source**: https://github.com/rodascaar/synkro

## System Requirements

- **macOS**: 10.15+ (Intel or Apple Silicon)
- **Linux**: Any modern distribution (x86_64 or ARM64)
- **Windows**: 10+ (x86_64 or ARM64)
- **Memory**: 500MB minimum
- **Disk Space**: 500MB free space

## License

Apache License 2.0

---

**Synkro** - Motor de Contexto Inteligente para LLMs

*One command. Ready to go.*
