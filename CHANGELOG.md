# Changelog

All notable changes to Synkro will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- MCP Server completo usando Go SDK oficial
- FTS5 full-text search con BM25 scoring
- Persistencia de session tracking en SQLite
- Persistencia de embedding cache en SQLite
- Persistencia de grafo de relaciones en SQLite
- TUI Add Memory form (tecla 'a')
- Suite completa de tests (>90% cobertura)
- Linting con golangci-lint
- CI/CD con GitHub Actions
- Variables de entorno para configuración (SYNKRO_DB_PATH, SYNKRO_DEBUG, etc.)
- Hybrid search combinando FTS5 y embeddings vectoriales
- Búsqueda semántica con cosine similarity

### Changed
- Actualizado a Go 1.25 para MCP SDK
- Refactorizado EmbeddingGenerator para soportar contexto
- Refactorizado SessionTracker con persistencia dual
- Mejorado performance de búsqueda con FTS5 MATCH queries
- Reestructurado Grafo con repository persistente

### Fixed
- Variables de entorno no funcionaban (ahora correctamente implementadas)
- Session tracking se perdía al reiniciar (ahora persistido)
- Embedding cache se perdía al reiniciar (ahora persistido)
- Relaciones no se persistían (ahora guardadas en SQLite)

### Technical
- Agregado modelo de datos MemoryRelation para grafo
- Implementado trigger automático para FTS5
- Agregado soporte para embeddings TF-IDF en 384 dimensiones
- Implementado cache LRU con límite configurable
- Agregado tracking de duplicados por sesión

## [2.0.0] - 2026-04-13

### Added
- TUI profesional con Bubble Tea
- Soporte para MCP Server
- Búsqueda semántica con embeddings
- Grafo de relaciones entre memorias
- Session tracking para evitar repeticiones
- Context pruning inteligente
- Soporte para múltiples tipos de memoria (note, decision, task, context)

### Changed
- Migrado de CSV a SQLite con FTS5
- Mejorada performance de búsqueda
- Mejorada usabilidad con TUI
