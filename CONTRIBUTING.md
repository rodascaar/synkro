# Contributing to Synkro

Thanks for your interest in contributing to Synkro! This guide covers the contribution process.

## Development

### Prerequisites

- Go 1.24+
- SQLite3 with FTS5 support
- CGO enabled (required for SQLite)

### Setup

```bash
git clone https://github.com/rodascaar/synkro.git
cd synkro
go mod download
./synkro init
```

For ONNX embeddings (optional):
```bash
brew install onnxruntime    # macOS
./synkro init --with-models  # Auto-download model + enable ONNX
```

### Running Tests

```bash
make test          # Run all tests
make test-short    # Quick tests
make test-coverage # Tests with coverage report
make bench         # Benchmarks
```

### Linting

```bash
make lint       # Run golangci-lint
make fmt        # Format code
```

### CI/CD

GitHub Actions runs tests, linting, and vulnerability scanning (`govulncheck`) on every push and PR.

## Project Structure

```
cmd/synkro/          # CLI entrypoint
internal/
  ├── config/        # Configuration and environment variables
  ├── db/           # SQLite initialization
  ├── memory/       # Memory models and repository
  ├── embeddings/   # Embeddings (TF-IDF, ONNX, cache)
  ├── graph/        # Relationship graph
  ├── mcp/          # MCP Server (Go SDK)
  ├── pruner/       # Context pruning
  ├── session/      # Session tracking
  └── tui/          # Bubble Tea TUI
```

## Code Conventions

- Follow `golangci-lint` configuration in `.golangci.yml`
- Use `MixedCaps` for exported names, `camelCase` for unexported
- Add `godoc` comments to exported functions
- No comments unless necessary

## Commit Messages

- Use imperative present tense: "Add feature" not "Added feature"
- Keep commits atomic and focused
- Reference issues when applicable

## Testing

- Tests go in `*_test.go` files
- Use `github.com/stretchr/testify` for assertions
- Create `setupTest*` helpers for environment setup
- Use `require` for fatal assertions, `assert` for regular checks
- Use `t.TempDir()` for temporary files

## Pull Requests

1. Ensure tests pass: `make test`
2. Ensure lint passes: `make lint`
3. Update documentation if needed
4. Add tests for new functionality
5. Update CHANGELOG.md if applicable

## Reporting Issues

When reporting issues, include:

- Go version: `go version`
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
