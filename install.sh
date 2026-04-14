#!/bin/bash

set -e

# SYNKRO INSTALLER - One-Command Installation Script
# Copyright (c) 2024 Synkro Project

VERSION="1.0.0"
INSTALL_DIR="$HOME/.synkro"
BIN_DIR="$INSTALL_DIR/bin"
CONFIG_DIR="$INSTALL_DIR/config"
DATA_DIR="$INSTALL_DIR/data"
MODELS_DIR="$INSTALL_DIR/models"
BINARY_NAME="synkro"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

banner() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════╗"
    echo "║                                        ║"
    echo "║        SYNKRO INSTALLER v$VERSION       ║"
    echo "║      Motor de Contexto Inteligente      ║"
    echo "║                                        ║"
    echo "╚══════════════════════════════════════╝"
    echo -e "${NC}"
}

check_dependencies() {
    echo -e "${BLUE}📦 Checking dependencies...${NC}"
    
    # Check if Go is installed (for building from source)
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        echo -e "${GREEN}✓${NC} Go found: $GO_VERSION"
    else
        echo -e "${YELLOW}⚠${NC}  Go not found (will use pre-built binary)"
    fi
    
    # Check CGO availability
    if [ "$(uname)" = "Darwin" ]; then
        if command -v xcode-select &> /dev/null; then
            echo -e "${GREEN}✓${NC} Xcode command line tools available"
        else
            echo -e "${RED}✗${NC} Xcode command line tools required"
            echo "  Install with: xcode-select --install"
            exit 1
        fi
    fi
    
    # Check disk space
    REQUIRED_SPACE=500 # MB
    AVAILABLE_SPACE=$(df -m "$HOME" | awk 'NR==2 {print $4}')
    if [ "$AVAILABLE_SPACE" -lt "$REQUIRED_SPACE" ]; then
        echo -e "${RED}✗${NC} Insufficient disk space"
        echo "  Required: ${REQUIRED_SPACE}MB, Available: ${AVAILABLE_SPACE}MB"
        exit 1
    else
        echo -e "${GREEN}✓${NC} Sufficient disk space (${AVAILABLE_SPACE}MB available)"
    fi
}

detect_os() {
    OS=$(uname -s)
    ARCH=$(uname -m)
    
    case "$OS" in
        Linux)
            PLATFORM="linux"
            ;;
        Darwin)
            PLATFORM="darwin"
            ;;
        *)
            echo -e "${RED}✗${NC} Unsupported OS: $OS"
            exit 1
            ;;
    esac
    
    case "$ARCH" in
        x86_64)
            BINARY_ARCH="amd64"
            ;;
        aarch64|arm64)
            BINARY_ARCH="arm64"
            ;;
        *)
            echo -e "${RED}✗${NC} Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    PLATFORM_SUFFIX="${PLATFORM}-${BINARY_ARCH}"
    echo -e "${GREEN}✓${NC} Detected platform: $PLATFORM_SUFFIX"
}

create_directories() {
    echo -e "${BLUE}📁 Creating directories...${NC}"
    
    mkdir -p "$BIN_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$DATA_DIR"
    mkdir -p "$MODELS_DIR"
    
    echo -e "${GREEN}✓${NC} Directories created in $INSTALL_DIR"
}

download_binary() {
    echo -e "${BLUE}⬇️  Downloading Synkro binary...${NC}"
    
    TEMP_DIR=$(mktemp -d)
    BINARY_PATH="$TEMP_DIR/$BINARY_NAME"
    
    RELEASES_URL="https://github.com/rodascaar/synkro/releases/latest"
    
    # Try to download from releases
    if curl -fsSL "$RELEASES_URL" | grep -q "browser_download_url"; then
        DOWNLOAD_URL=$(curl -fsSL "$RELEASES_URL" | grep "browser_download_url" | grep "darwin" | grep "amd64" | head -1 | sed 's/.*"https:/https/' | sed 's/".*//')
        
        echo "Downloading from: $DOWNLOAD_URL"
        
        if ! curl -fsSL -o "$BINARY_PATH" "$DOWNLOAD_URL"; then
            echo -e "${RED}✗${NC} Failed to download binary"
            echo "  Falling back to building from source..."
            build_from_source
            return
        fi
    else
        echo -e "${YELLOW}⚠${NC}  No pre-built binary available"
        echo "  Building from source..."
        build_from_source
        return
    fi
    
    chmod +x "$BINARY_PATH"
    
    # Move binary to installation directory
    mv "$BINARY_PATH" "$BIN_DIR/$BINARY_NAME"
    rm -rf "$TEMP_DIR"
    
    echo -e "${GREEN}✓${NC} Binary installed to $BIN_DIR/$BINARY_NAME"
}

build_from_source() {
    if ! command -v go &> /dev/null; then
        echo -e "${RED}✗${NC} Go is required to build from source"
        echo "  Install from: https://go.dev/dl/"
        exit 1
    fi
    
    echo "Building Synkro from source..."
    
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # Clone repository
    if ! git clone https://github.com/rodascaar/synkro.git synkro; then
        echo -e "${RED}✗${NC} Failed to clone repository"
        exit 1
    fi
    
    cd synkro
    
    # Build binary
    if ! go build -o "$BIN_DIR/$BINARY_NAME" ./cmd/synkro/; then
        echo -e "${RED}✗${NC} Failed to build Synkro"
        exit 1
    fi
    
    chmod +x "$BIN_DIR/$BINARY_NAME"
    
    # Cleanup
    cd "$HOME"
    rm -rf "$TEMP_DIR"
    
    echo -e "${GREEN}✓${NC} Synkro built and installed"
}

setup_environment() {
    echo -e "${BLUE}⚙️  Setting up environment...${NC}"
    
    # Create shell configuration
    SHELL_CONFIG="$CONFIG_DIR/synkro.sh"
    cat > "$SHELL_CONFIG" <<EOF
#!/bin/bash
# Synkro environment configuration
export SYNKRO_HOME="$INSTALL_DIR"
export SYNKRO_CONFIG="$CONFIG_DIR"
export SYNKRO_DATA="$DATA_DIR"
export SYNKRO_MODELS="$MODELS_DIR"
export PATH="$BIN_DIR:\$PATH"
EOF
    
    chmod +x "$SHELL_CONFIG"
    
    echo -e "${GREEN}✓${NC} Environment configured"
}

install_shell_integration() {
    echo -e "${BLUE}🐚 Installing shell integration...${NC}"
    
    SHELL_NAME=$(basename "$SHELL")
    
    case "$SHELL_NAME" in
        bash|zsh)
            SHELL_CONFIG_FILE="$HOME/.$SHELL_NAME"
            
            if ! grep -q 'Synkro' "$SHELL_CONFIG_FILE"; then
                echo "" >> "$SHELL_CONFIG_FILE"
                echo "# Synkro integration" >> "$SHELL_CONFIG_FILE"
                echo "source \"$CONFIG_DIR/synkro.sh\" # Added by Synkro installer" >> "$SHELL_CONFIG_FILE"
                
                echo -e "${GREEN}✓${NC} Shell integration added to $SHELL_CONFIG_FILE"
                
                echo -e "${YELLOW}⚠${NC}  Please run: source $SHELL_CONFIG_FILE"
            else
                echo -e "${YELLOW}⚠${NC}  Shell integration already exists"
            fi
            ;;
        *)
            echo -e "${YELLOW}⚠${NC}  Shell integration not supported for $SHELL_NAME"
            ;;
    esac
}

initialize_database() {
    echo -e "${BLUE}💾 Initializing database...${NC}"
    
    # Check if database already exists
    if [ -f "$DATA_DIR/memory.db" ]; then
        echo -e "${YELLOW}⚠${NC}  Database already exists at $DATA_DIR/memory.db"
        read -p "Do you want to reinitialize? (y/N): " -n 1 -r
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${GREEN}✓${NC} Keeping existing database"
            return
        fi
    fi
    
    # Initialize database
    if ! "$BIN_DIR/$BINARY_NAME" init; then
        echo -e "${RED}✗${NC} Failed to initialize database"
        exit 1
    fi
    
    echo -e "${GREEN}✓${NC} Database initialized"
}

download_default_model() {
    echo -e "${BLUE}🤖 Setting up embedding model...${NC}"
    
    # Ask if user wants to download default model
    echo ""
    echo "Synkro can use semantic embeddings for better search accuracy."
    echo "Do you want to download the default embedding model (all-MiniLM-L6-v2)?"
    echo ""
    read -p "Download default model? (Y/n): " -n 1 -r
    
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        echo -e "${YELLOW}⚠${NC}  Skipping model download"
        echo "  You can download models later with: synkro model download <model-name>"
        return
    fi
    
    # Download default model
    echo "Downloading default model..."
    if ! "$BIN_DIR/$BINARY_NAME" model download all-MiniLM-L6-v2; then
        echo -e "${YELLOW}⚠${NC}  Model download failed, but Synkro is ready to use"
        echo "  You can download models later with: synkro model list"
    else
        echo -e "${GREEN}✓${NC} Default model installed"
    fi
}

print_success() {
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║${NC}       SYNKRO INSTALLED SUCCESSFULLY!      ${GREEN}║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}🚀 Getting Started:${NC}"
    echo ""
    echo "1. Reload your shell or run:"
    echo -e "   ${YELLOW}source $CONFIG_DIR/synkro.sh${NC}"
    echo ""
    echo "2. Initialize database (if not already done):"
    echo -e "   ${YELLOW}synkro init${NC}"
    echo ""
    echo "3. Add your first memory:"
    echo -e "   ${YELLOW}synkro add --title \"My first memory\" --content \"Hello Synkro!\" --type note${NC}"
    echo ""
    echo "4. Start TUI:"
    echo -e "   ${YELLOW}synkro tui${NC}"
    echo ""
    echo -e "${BLUE}📚 Documentation:${NC}"
    echo "  Visit: https://github.com/rodascaar/synkro/wiki"
    echo ""
    echo -e "${BLUE}💡 Tips:${NC}"
    echo "  - Use 'synkro --help' for all commands"
    echo "  - Check available models with 'synkro model list'"
    echo "  - Run MCP server with 'synkro mcp'"
    echo ""
}

# Main installation flow
main() {
    banner
    check_dependencies
    detect_os
    create_directories
    download_binary
    setup_environment
    install_shell_integration
    initialize_database
    download_default_model
    print_success
}

main
