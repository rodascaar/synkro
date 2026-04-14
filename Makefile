.PHONY: build test lint ci clean help

all: build

build:
	@echo "Building synkro..."
	CGO_ENABLED=1 go build \
		-tags sqlite_fts5 \
		-ldflags="-s -w" \
		-trimpath \
		-o synkro \
		./cmd/synkro/

build-release:
	@echo "Building release binaries..."
	CGO_ENABLED=1 go build \
		-tags sqlite_fts5 \
		-ldflags="-s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(DATE)" \
		-trimpath \
		-o dist/synkro \
		./cmd/synkro/
	@ls -lh dist/synkro

test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-short:
	@echo "Running short tests..."
	go test -v -short ./...

bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

lint:
	@echo "Running linter..."
	golangci-lint run --config=.golangci.yml ./...

lint-fix:
	@echo "Running linter with auto-fix..."
	golangci-lint run --config=.golangci.yml --fix ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

ci: lint test

clean:
	@echo "Cleaning..."
	rm -f synkro
	rm -f dist/synkro
	rm -f coverage.out coverage.html
	rm -f memory.db memory.db-wal memory.db-shm

deps:
	@echo "Installing dependencies..."
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

init-db:
	@echo "Initializing database..."
	./synkro init

help:
	@echo "Available commands:"
	@echo "  make build        - Build synkro binary"
	@echo "  make build-release - Build release binary with version info"
	@echo "  make test         - Run all tests with coverage"
	@echo "  make test-short   - Run short tests"
	@echo "  make bench        - Run benchmarks"
	@echo "  make lint         - Run linter"
	@echo "  make lint-fix     - Run linter with auto-fix"
	@echo "  make fmt          - Format code"
	@echo "  make ci           - Run lint and test (CI)"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make deps         - Install development dependencies"
	@echo "  make init-db      - Initialize database"
	@echo "  make help         - Show this help message"
