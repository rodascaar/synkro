# Synkro MCP Installation

One-command installation for opencode, claudecode, geminicode, or any MCP-compatible environment.

## Quick Install

### macOS / Linux (One command)
```bash
curl -fsSL https://raw.githubusercontent.com/nichogram/synkro/main/install.sh | sh
```

### Windows (PowerShell)
```powershell
iwr -useb https://raw.githubusercontent.com/nichogram/synkro/main/install.ps1 | iex
```

## Manual Install

1. **Download binary:**
```bash
# macOS ARM64 (Apple Silicon)
wget https://github.com/rodascaar/synkro/releases/latest/download/synkro-darwin-arm64 -O /usr/local/bin/synkro

# macOS Intel
wget https://github.com/rodascaar/synkro/releases/latest/download/synkro-darwin-amd64 -O /usr/local/bin/synkro

# Linux ARM64
wget https://github.com/rodascaar/synkro/releases/latest/download/synkro-linux-arm64 -O /usr/local/bin/synkro

# Linux AMD64
wget https://github.com/rodascaar/synkro/releases/latest/download/synkro-linux-amd64 -O /usr/local/bin/synkro

chmod +x /usr/local/bin/synkro
```

2. **Initialize database:**
```bash
synkro init
```

3. **Enable embeddings (optional):**
```bash
synkro init --with-models
```

## IDE Configuration

### Opencode
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

### Claude Desktop (macOS)
Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "synkro": {
      "command": "synkro",
      "args": ["mcp"]
    }
  }
}
```

### Continue.dev
Edit `.continuerc.json`:
```json
{
  "mcpServers": {
    "synkro": {
      "command": "synkro",
      "args": ["mcp"]
    }
  }
}
```

### GPT-4 / Other
Most IDEs support the same MCP server format. Use:
- Command: `synkro`
- Args: `["mcp"]`

## Verify Installation

```bash
# Check synkro works
synkro --help

# Test TUI
synkro tui

# Add test memory
synkro add --title "Test memory" --content "This is a test"

# List memories
synkro list
```

## Environment Variables (Optional)

```bash
# Custom database location
export SYNKRO_DB_PATH=~/custom/path/memory.db

# Enable debug mode
export SYNKRO_DEBUG=true
```

## Dependencies

Synkro is a **single binary** with zero dependencies:
- ✅ No Python required
- ✅ No npm/node required
- ✅ No external servers
- ✅ No API keys needed

## Building from Source

```bash
git clone https://github.com/rodascaar/synkro.git
cd synkro
go build -o synkro ./cmd/synkro/
sudo mv synkro /usr/local/bin/
```

## Troubleshooting

**"command not found"**: Ensure `/usr/local/bin` is in your PATH
```bash
export PATH=$PATH:/usr/local/bin
```

**Database locked**: Close other Synkro instances or delete `.db-wal` files
```bash
rm -f memory.db-wal memory.db-shm
```

**MCP not showing**: Restart your IDE after adding config
