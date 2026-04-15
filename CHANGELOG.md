# Changelog

All notable changes to Synkro will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2025-04-15

### Architecture (20 improvements across 4 phases)

#### Foundation
- Config JSON with auto-migration from KEY=VALUE format and env var overrides
- sqlite-vec KNN search with `vec0` virtual tables (Linux/macOS)
- Embedding regeneration on memory update when title/content changes
- Tags normalization to `memory_tags` junction table
- Versioned migration system with 3 initial migrations

#### Brain & Stability
- Bidirectional BFS graph pathfinding
- `internal/errors` package integrated across MCP handlers
- Health check with `go/version` stdlib comparison
- FTS5 query sanitization (special characters escaping)
- Real LRU cache with `container/list` and proper eviction

#### Delivery
- MCP handlers moved to Server methods (no global state)
- TUI search using HybridSearch with substring fallback
- SHA256 checksum verification on auto-update

#### Refinement
- Session tracker deterministic ordering (sorted by `DeliveredAt`)
- Clean semver parsing (`parseSemver` + `compareVersions`)
- ONNX mean pooling with attention mask
- CLI `delete` command with memory verification
- Configurable model download timeout

### Testing
- `sanitize_test.go` — 10 test cases for FTS5 query sanitization
- `cache_test.go` — 9 test cases for LRU cache (eviction, persistence, ordering)
- `vector_test.go` — sqlite-vec graceful degradation tests
- `session_test.go` — deterministic delivery ordering tests
- `e2e_test.go` — end-to-end CRUD + search + tags flow
- `update_test.go` — semver comparison tests

### Documentation
- Fixed Go version badge (1.22+ → 1.24+)
- Added `delete` command to all docs
- Fixed dead links and incorrect command references
- Updated architecture docs with new files

### CI/CD
- Migrated `.golangci.yml` to v2 schema
- Build constraints for sqlite-vec (`//go:build !windows`)
- Graceful degradation on Windows (cosine similarity fallback)
- Added macOS dependencies step

### Cleanup
- Removed `.bak` files, test binaries, stale scripts

## [1.0.0] - Initial Release

- MCP Server with 11 tools using Go SDK official
- FTS5 full-text search with BM25 scoring
- TF-IDF embeddings (384 dims) with persistent cache
- Memory graph with 6 relation types
- Bubble Tea TUI with 3 panels
- ONNX model support
- Session tracking (in-memory + SQLite persistent)
- Context pruning and deduplication
- Auto-update mechanism
