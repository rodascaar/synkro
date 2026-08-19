# SYNKRO - Makefile for building and packaging

# Version information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0-dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -s -w"

# modernc.org/sqlite es puro Go y no requiere CGO, pero onnxruntime_go
# (embeddings ONNX) y go test -race sí requieren CGO.
CGO_ENABLED := 1

# Directories
DIST_DIR := dist
BUILD_DIR := build
MODELS_DIR := models
INSTALL_DIR := $(HOME)/.synkro

.PHONY: all build build-all build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 build-windows-arm64 clean test test-coverage test-short bench lint fmt package install release checksums dev dev-install help

all: build

build:
	@echo "Building Synkro $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $(BUILD_DIR)/synkro ./cmd/synkro/
	@echo "Build complete: $(BUILD_DIR)/synkro"

build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 build-windows-arm64

build-linux-amd64:
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $(DIST_DIR)/synkro-linux-amd64 ./cmd/synkro/

build-linux-arm64:
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $(DIST_DIR)/synkro-linux-arm64 ./cmd/synkro/

build-darwin-amd64:
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $(DIST_DIR)/synkro-darwin-amd64 ./cmd/synkro/

build-darwin-arm64:
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $(DIST_DIR)/synkro-darwin-arm64 ./cmd/synkro/

build-windows-amd64:
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $(DIST_DIR)/synkro-amd64.exe ./cmd/synkro/

build-windows-arm64:
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $(DIST_DIR)/synkro-arm64.exe ./cmd/synkro/

test:
	@echo "Running tests..."
	go test -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

test-coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-short:
	@echo "Running quick tests..."
	go test ./... -short -timeout 30s

bench:
	@echo "Running benchmarks..."
	go test ./... -bench=. -benchmem -timeout 60s

lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --config .golangci.yml; \
	else \
		echo "golangci-lint not installed. Install from: https://golangci-lint.run/usage/install"; \
	fi

fmt:
	go fmt ./...
	@echo "Code formatted"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html
	rm -f memory.db memory.db-wal memory.db-shm
	@echo "Clean complete"

package: build-all
	@echo "Creating packages..."
	@mkdir -p $(DIST_DIR)/packages

	@for os in linux-amd64 linux-arm64 darwin-amd64 darwin-arm64; do \
		mkdir -p $(DIST_DIR)/packages/synkro-$(VERSION)-$$os; \
		cp $(DIST_DIR)/synkro-$$os $(DIST_DIR)/packages/synkro-$(VERSION)-$$os/synkro; \
		cp install.sh $(DIST_DIR)/packages/synkro-$(VERSION)-$$os/; \
		chmod +x $(DIST_DIR)/packages/synkro-$(VERSION)-$$os/synkro; \
		tar czf $(DIST_DIR)/synkro-$(VERSION)-$$os.tar.gz -C $(DIST_DIR)/packages synkro-$(VERSION)-$$os; \
	done

	@for arch in amd64 arm64; do \
		mkdir -p $(DIST_DIR)/packages/synkro-$(VERSION)-windows-$$arch; \
		cp $(DIST_DIR)/synkro-$$arch.exe $(DIST_DIR)/packages/synkro-$(VERSION)-windows-$$arch/synkro.exe; \
		cp install.ps1 $(DIST_DIR)/packages/synkro-$(VERSION)-windows-$$arch/; \
		cd $(DIST_DIR)/packages && zip -qr ../synkro-$(VERSION)-windows-$$arch.zip synkro-$(VERSION)-windows-$$arch && cd ../..; \
	done

	@echo "Packages created in $(DIST_DIR)/packages/"

install: build
	@echo "Installing Synkro to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)/bin
	@cp $(BUILD_DIR)/synkro $(INSTALL_DIR)/bin/
	@chmod +x $(INSTALL_DIR)/bin/synkro
	@mkdir -p $(INSTALL_DIR)/data
	@mkdir -p $(INSTALL_DIR)/models
	@echo "Installed to $(INSTALL_DIR)"

release: clean package
	@echo "Creating release..."
	@echo "Version: $(VERSION)"
	@ls -lh $(DIST_DIR)/packages/

checksums: package
	@cd $(DIST_DIR)/packages && for file in synkro-$(VERSION)-*.tar.gz synkro-$(VERSION)-*.zip; do \
		sha256sum $$file > $$file.sha256; \
	done

dev:
	./$(BUILD_DIR)/synkro mcp

dev-install: install
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin; \
	fi

help:
	@echo "SYNKRO - Makefile for building and packaging"
	@echo ""
	@echo "Targets: build build-all test test-coverage test-short bench lint fmt clean"
	@echo "         package install release checksums dev dev-install help"
	@echo ""
	@echo "Examples:"
	@echo "  make build       # Build for current platform"
	@echo "  make package     # Create all distribution packages"
	@echo "  make install     # Install to ~/.synkro"
	@echo "  make test        # Run all tests"