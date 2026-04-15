# Changelog

All notable changes to Synkro will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.1.0] - 2026-04-15

### Security

- Added `govulncheck` vulnerability scanning to CI pipeline
- Pinned `goreleaser-action` to v5.4.0 (supply chain risk mitigation)
- Added explicit `permissions` blocks to all GitHub Actions workflows

### Bug Fixes

- `saveEmbedding` now logs warnings instead of silently discarding errors
- MCP `AddMemory` handler validates Type enum (note/decision/task/context)
- `addCmd` and `mcpCmd` now use `cfg.PreferredModel` and `cfg.ModelDir` from config
- `initCmd` no longer blocks in non-interactive environments (`--no-tutorial` flag + terminal detection)
- Replaced 11 silent `json.MarshalIndent` error discards with proper `writeJSON` helper

### Tests

- MCP graph tools: AddRelation, GetRelations, DeleteRelation, FindPath (12 new tests)
- MCP activate_context: deduplication, low similarity paths
- Hybrid search: multi-result coverage
- Memory repository: GetByTag with type filter and limit
- CLI commands: init, add, version, health (subprocess-based)
- MCP Server: Run() start/stop integration test
- TUI tutorial: 9 tests (navigation, progress, all steps)
- TUI graph_view: 10 tests (CRUD, layout, render, GetNodeAt)
- TUI model: renderDetail, renderContent with tags
- Update: parseSemver edge cases, fileSHA256, findChecksumForAsset with mock HTTP
- Embeddings: ModelManager (GetModel, DeleteModel, ValidateModel, GetPreferredModel)

### Documentation

- README translated to English
- CONTRIBUTING.md rewritten in English with correct references
- AGENTS.md coverage claim corrected (~70%+)
- Added PR template and issue templates (bug report, feature request)

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
- FTS5 query sanitization tests
- LRU cache tests (eviction, persistence, ordering)
- sqlite-vec graceful degradation tests
- Session deterministic delivery ordering tests
- End-to-end CRUD + search + tags flow
- Semver comparison tests

### Documentation
- Fixed Go version badge (1.22+ -> 1.24+)
- Added `delete` command to all docs
- Fixed dead links and incorrect command references

### CI/CD
- Migrated `.golangci.yml` to v2 schema
- Build constraints for sqlite-vec (`//go:build !windows`)
- Graceful degradation on Windows (cosine similarity fallback)

## [1.0.0] - Initial Release

- MCP Server with 11 tools using Go SDK
- FTS5 full-text search with BM25 scoring
- TF-IDF embeddings (384 dims) with persistent cache
- Memory graph with 6 relation types
- Bubble Tea TUI with 3 panels
- ONNX model support
- Session tracking (in-memory + SQLite persistent)
- Context pruning and deduplication
- Auto-update mechanism
