#!/bin/bash
# Synkro One-Click Installer v1.0
# Supports: macOS (Intel/Apple Silicon), Linux (AMD64/ARM64)
set -e

VERSION="latest"
REPO="rodascaar/synkro"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="synkro"

echo "🚀 Installing Synkro v1..."

# Detectar plataforma
detect_platform() {
    OS=$(uname -s)
    ARCH=$(uname -m)

    case $OS in
        Darwin)
            if [ "$ARCH" = "arm64" ]; then
                echo "darwin-arm64"
            else
                echo "darwin-amd64"
            fi
            ;;
        Linux)
            case $ARCH in
                aarch64|arm64)
                    echo "linux-arm64"
                    ;;
                *)
                    echo "linux-amd64"
                    ;;
            esac
            ;;
        *)
            echo "unsupported"
            return 1
            ;;
    esac
}

PLATFORM=$(detect_platform)
if [ $? -ne 0 ]; then
    echo "❌ Unsupported platform: $(uname -s)/$(uname -m)"
    echo "Supported platforms:"
    echo "  - macOS (Intel/Apple Silicon)"
    echo "  - Linux (AMD64/ARM64)"
    exit 1
fi

echo "📦 Detected platform: $PLATFORM"

GITHUB_RELEASE_URL="https://github.com/$REPO/releases/latest/download/synkro-${PLATFORM}"

# Fallback a compilación desde fuente
compile_from_source() {
    echo "🔨 Fallback: Compiling from source..."

    if ! command -v go &> /dev/null; then
        echo "❌ Go not installed. Please install Go from https://go.dev/dl/"
        exit 1
    fi

    echo "📦 Cloning repository..."
    git clone --depth 1 https://github.com/$REPO.git /tmp/synkro-source

    echo "🔨 Building..."
    cd /tmp/synkro-source
    CGO_ENABLED=1 go build -tags sqlite_fts5 -ldflags="-s -w" -o /tmp/synkro ./cmd/synkro/

    echo "✅ Build complete!"
    rm -rf /tmp/synkro-source
}

# Intentar descargar binario
echo "📥 Attempting to download pre-built binary..."
if curl -fsSL -f "$GITHUB_RELEASE_URL" -o /tmp/synkro 2>/dev/null; then
    echo "✅ Binary downloaded successfully"
else
    echo "⚠️  Pre-built binary not available"
    echo "🔨 Compiling from source instead..."
    compile_from_source
fi

# Hacer ejecutable
chmod +x /tmp/synkro

# Instalar
echo "📦 Installing to $INSTALL_DIR..."
if [ ! -w "$INSTALL_DIR" ]; then
    echo "⚠️  Need sudo to install to $INSTALL_DIR"
    sudo mkdir -p "$INSTALL_DIR"
    sudo mv /tmp/synkro "$INSTALL_DIR/synkro"
else
    mkdir -p "$INSTALL_DIR"
    mv /tmp/synkro "$INSTALL_DIR/synkro"
fi

# Verificar
if command -v synkro &> /dev/null; then
    INSTALLED_VERSION=$(synkro version 2>&1 | head -1 || echo "unknown")
    echo ""
    echo "✅ Synkro installed successfully!"
    echo "   Version: $INSTALLED_VERSION"
    echo ""
    echo "🎯 Quick Start:"
    echo "   synkro init              # Initialize database"
    echo "   synkro add --help        # Add your first memory"
    echo "   synkro tui               # Launch TUI"
    echo "   synkro mcp               # Start MCP server"
    echo ""
    echo "📚 Documentation:"
    echo "   https://github.com/$REPO"
else
    echo "❌ Installation failed"
    echo "Please ensure $INSTALL_DIR is in your PATH"
    exit 1
fi

echo ""
echo "🎉 Installation complete!"
