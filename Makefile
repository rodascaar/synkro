# SYNKRO - Makefile for building and packaging
# Professional build and packaging system

# Version information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0-dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -s -w"

# Go build flags
CGO_ENABLED := 1
GOFLAGS := -v

# Directories
DIST_DIR := dist
BUILD_DIR := build
MODELS_DIR := models
INSTALL_DIR := $(HOME)/.synkro

# Build targets
.PHONY: all build clean install test test-short bench lint package help

# Default target
all: build

# Build binary for current platform
build:
	@echo "🔨 Building Synkro $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/synkro ./cmd/synkro/
	@echo "✅ Build complete: $(BUILD_DIR)/synkro"

# Build for all platforms
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 build-windows-arm64

build-linux-amd64:
	@echo "🔨 Building for Linux AMD64..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/synkro-linux-amd64 ./cmd/synkro/
	@echo "✅ Built: $(DIST_DIR)/synkro-linux-amd64"

build-linux-arm64:
	@echo "🔨 Building for Linux ARM64..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/synkro-linux-arm64 ./cmd/synkro/
	@echo "✅ Built: $(DIST_DIR)/synkro-linux-arm64"

build-darwin-amd64:
	@echo "🔨 Building for macOS Intel..."
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/synkro-darwin-amd64 ./cmd/synkro/
	@echo "✅ Built: $(DIST_DIR)/synkro-darwin-amd64"

build-darwin-arm64:
	@echo "🔨 Building for macOS Apple Silicon..."
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/synkro-darwin-arm64 ./cmd/synkro/
	@echo "✅ Built: $(DIST_DIR)/synkro-darwin-arm64"

build-windows-amd64:
	@echo "🔨 Building for Windows AMD64..."
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/synkro.exe ./cmd/synkro/
	@echo "✅ Built: $(DIST_DIR)/synkro.exe"

build-windows-arm64:
	@echo "🔨 Building for Windows ARM64..."
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/synkro.exe ./cmd/synkro/
	@echo "✅ Built: $(DIST_DIR)/synkro.exe"

# Run tests
test:
	@echo "🧪 Running tests..."
	CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" go test -v -race -coverprofile=coverage.out ./...
	@echo "✅ Tests complete"

# Run tests with coverage
test-coverage: test
	@echo "📊 Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

# Run quick tests
test-short:
	@echo "⚡ Running quick tests..."
	go test ./... -short -timeout 30s
	@echo "✅ Quick tests complete"

# Run benchmarks
bench:
	@echo "📊 Running benchmarks..."
	go test ./... -bench=. -benchmem -timeout 60s
	@echo "✅ Benchmarks complete"

# Run linter
lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --config .golangci.yml; \
	else \
		echo "⚠ golangci-lint not installed. Install from: https://golangci-lint.run/usage/install"; \
	fi

# Format code
fmt:
	@echo "📝 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html
	rm -f memory.db memory.db-wal memory.db-shm
	@echo "✅ Clean complete"

# Package for distribution
package: build-all
	@echo "📦 Creating packages..."
	@mkdir -p $(DIST_DIR)/packages
	
	# Create Linux packages
	@echo "Creating Linux AMD64 package..."
	@mkdir -p $(DIST_DIR)/packages/synkro-$(VERSION)-linux-amd64
	@cp $(DIST_DIR)/synkro-linux-amd64 $(DIST_DIR)/packages/synkro-$(VERSION)-linux-amd64/synkro
	@cp install.sh $(DIST_DIR)/packages/synkro-$(VERSION)-linux-amd64/
	@chmod +x $(DIST_DIR)/packages/synkro-$(VERSION)-linux-amd64/synkro
	@cd $(DIST_DIR)/packages && tar czf ../synkro-$(VERSION)-linux-amd64.tar.gz synkro-$(VERSION)-linux-amd64/ && cd ../../..
	
	@echo "Creating Linux ARM64 package..."
	@mkdir -p $(DIST_DIR)/packages/synkro-$(VERSION)-linux-arm64
	@cp $(DIST_DIR)/synkro-linux-arm64 $(DIST_DIR)/packages/synkro-$(VERSION)-linux-arm64/synkro
	@cp install.sh $(DIST_DIR)/packages/synkro-$(VERSION)-linux-arm64/
	@chmod +x $(DIST_DIR)/packages/synkro-$(VERSION)-linux-arm64/synkro
	@cd $(DIST_DIR)/packages && tar czf ../synkro-$(VERSION)-linux-arm64.tar.gz synkro-$(VERSION)-linux-arm64/ && cd ../../..
	
	# Create macOS packages
	@echo "Creating macOS Intel package..."
	@mkdir -p $(DIST_DIR)/packages/synkro-$(VERSION)-darwin-amd64
	@cp $(DIST_DIR)/synkro-darwin-amd64 $(DIST_DIR)/packages/synkro-$(VERSION)-darwin-amd64/synkro
	@cp install.sh $(DIST_DIR)/packages/synkro-$(VERSION)-darwin-amd64/
	@chmod +x $(DIST_DIR)/packages/synkro-$(VERSION)-darwin-amd64/synkro
	@cd $(DIST_DIR)/packages && tar czf ../synkro-$(VERSION)-darwin-amd64.tar.gz synkro-$(VERSION)-darwin-amd64/ && cd ../../..
	
	@echo "Creating macOS Apple Silicon package..."
	@mkdir -p $(DIST_DIR)/packages/synkro-$(VERSION)-darwin-arm64
	@cp $(DIST_DIR)/synkro-darwin-arm64 $(DIST_DIR)/packages/synkro-$(VERSION)-darwin-arm64/synkro/
	@cp install.sh $(DIST_DIR)/packages/synkro-$(VERSION)-darwin-arm64/
	@chmod +x $(DIST_DIR)/packages/synkro-$(VERSION)-darwin-arm64/synkro
	@cd $(DIST_DIR)/packages && tar czf ../synkro-$(VERSION)-darwin-arm64.tar.gz synkro-$(VERSION)-darwin-arm64/ && cd ../../..
	
	# Create Windows packages
	@echo "Creating Windows AMD64 package..."
	@mkdir -p $(DIST_DIR)/packages/synkro-$(VERSION)-windows-amd64
	@cp $(DIST_DIR)/synkro.exe $(DIST_DIR)/packages/synkro-$(VERSION)-windows-amd64/
	@cp install.ps1 $(DIST_DIR)/packages/synkro-$(VERSION)-windows-amd64/
	@cd $(DIST_DIR)/packages && powershell -Command "Compress-Archive -Path synkro-$(VERSION)-windows-amd64 -DestinationPath ../synkro-$(VERSION)-windows-amd64.zip" && cd ../../..
	
	@echo "Creating Windows ARM64 package..."
	@mkdir -p $(DIST_DIR)/packages/synkro-$(VERSION)-windows-arm64
	@cp $(DIST_DIR)/synkro.exe $(DIST_DIR)/packages/synkro-$(VERSION)-windows-arm64/synkro
	@cp install.ps1 $(DIST_DIR)/packages/synkro-$(VERSION)-windows-arm64/
	@cd $(DIST_DIR)/packages && powershell -Command "Compress-Archive -Path synkro-$(VERSION)-windows-arm64 -DestinationPath ../synkro-$(VERSION)-windows-arm64.zip" && cd ../../..
	
	@echo "✅ Packages created in $(DIST_DIR)/packages/"

# Install to local system
install: build
	@echo "📦 Installing Synkro to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)/bin
	@cp $(BUILD_DIR)/synkro $(INSTALL_DIR)/bin/
	@chmod +x $(INSTALL_DIR)/bin/synkro
	@mkdir -p $(INSTALL_DIR)/data
	@mkdir -p $(INSTALL_DIR)/models
	@echo "✅ Installed to $(INSTALL_DIR)"
	@echo ""
	@echo "🚀 Next steps:"
	@echo "1. Add $(INSTALL_DIR)/bin to your PATH"
	@echo "2. Run: synkro init"
	@echo "3. Run: synkro tui"

# Create release (for GitHub Actions)
release: clean package
	@echo "🚀 Creating release..."
	@echo "Version: $(VERSION)"
	@echo "Packages ready for upload"
	@ls -lh $(DIST_DIR)/packages/

# Generate checksums
checksums: package
	@echo "🔐 Generating checksums..."
	@cd $(DIST_DIR)/packages
	@for file in synkro-$(VERSION)-*.tar.gz synkro-$(VERSION)-*.zip; do \
		sha256sum $$file > $$file.sha256; \
	done
	@echo "✅ Checksums generated"

# Run development server
dev:
	@echo "🚀 Starting development server..."
	./$(BUILD_DIR)/synkro mcp

# Development helpers
dev-install: install
	@echo "🔧 Installing development dependencies..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin; \
	fi
	@echo "✅ Development dependencies installed"

# Help target
help:
	@echo "SYNKRO - Makefile for building and packaging"
	@echo ""
	@echo "Available targets:"
	@echo "  all             - Build binary for current platform (default)"
	@echo "  build           - Build binary for current platform"
	@echo "  build-all       - Build binaries for all platforms"
	@echo "  clean           - Remove build artifacts"
	@echo "  test            - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"
	@echo "  package         - Create distribution packages"
	@echo "  release         - Create GitHub release"
	@echo "  checksums       - Generate SHA256 checksums"
	@echo "  install         - Install to local system"
	@echo "  dev             - Start development server"
	@echo "  dev-install     - Install development dependencies"
	@echo "  help            - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build                    # Build for current platform"
	@echo "  make package                  # Create all distribution packages"
	@echo "  make install                  # Install to ~/.synkro"
	@echo "  make test                    # Run all tests"
