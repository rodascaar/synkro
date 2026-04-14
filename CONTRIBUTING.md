# Contributing to Synkro

Gracias por tu interés en contribuir a Synkro! Este documento te guiará a través del proceso de contribución.

## Desarrollo

### Prerrequisitos

- Go 1.25 o superior
- SQLite3 con soporte FTS5
- golangci-lint (instalado con `make deps`)

### Configuración del Entorno

```bash
# Clonar el repositorio
git clone https://github.com/rodascaar/synkro.git
cd synkro

# Instalar dependencias de desarrollo
make deps

# Inicializar base de datos para pruebas
./synkro init
```

### Ejecutar Tests

```bash
# Ejecutar todos los tests
make test

# Ejecutar tests rápidos
make test-short

# Ejecutar benchmarks
make bench
```

### Ejecutar Linter

```bash
# Ejecutar linter
make lint

# Ejecutar linter con auto-fix
make lint-fix
```

### Formatear Código

```bash
# Formatear código
make fmt
```

### CI/CD

El proyecto usa GitHub Actions para CI/CD. Los tests y linters se ejecutan automáticamente en cada push y PR.

## Estructura del Proyecto

```
cmd/synkro/          # CLI entrypoint
internal/
  ├── config/        # Configuración y variables de entorno
  ├── db/           # SQLite initialization
  ├── memory/       # Memory models y repository
  ├── embeddings/   # Embeddings (TF-IDF, cache)
  ├── graph/        # Grafo de relaciones
  ├── mcp/          # MCP Server con Go SDK
  ├── pruner/       # Context pruning
  ├── session/      # Session tracking
  └── tui/          # Bubble Tea TUI
```

## Convenciones de Código

### Estilo de Código

- Seguir `golangci-lint` configuración en `.golangci.yml`
- Usar nombres en MixedCaps para exportados
- Usar nombres en camelCase para privados
- Agregar godoc comments a funciones exportadas

### Commit Messages

- Usar presente: "Add feature" no "Added feature"
- Mantener commits atómicos y enfocados
- Referenciar issues cuando sea aplicable

## Testing

### Escribir Tests

- Los tests deben ir en archivos `*_test.go`
- Usar `github.com/stretchr/testify` para assertions
- Crear helper functions `setupTest*` para preparar el entorno
- Usar `require` para errores fatales y `assert` para checks regulares

### Cobertura

- Buscar mantener >90% de cobertura
- Tests críticos para funcionalidad principal
- Tests de integración para MCP Server
- Benchmarks para operaciones frecuentes

## Pull Requests

### Antes de Crear un PR

1. Ejecutar `make ci` para asegurar que tests y linters pasan
2. Actualizar documentación si es necesario
3. Agregar tests para nuevas funcionalidades
4. Agregar entradas al CHANGELOG.md

### Proceso de Review

1. Los PRs son revisados por maintainers
2. Los tests deben pasar en CI
3. El código debe pasar linters
4. Los comentarios de review deben ser respondidos

## Reportar Issues

Al reportar issues, incluye:

- Versión de Go: `go version`
- Sistema operativo
- Pasos para reproducir
- Comportamiento esperado vs actual
- Logs relevantes

## Licencia

Al contribuir, aceptas que tus contribuciones sean licenciadas bajo la MIT License.
