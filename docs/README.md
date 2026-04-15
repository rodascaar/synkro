# Synkro v2 🚀

> **Motor de Contexto Inteligente para LLMs** - Sistema de memoria con embeddings, grafo de relaciones y pruning inteligente

![Version](https://img.shields.io/badge/version-2.0-blue)
![Go](https://img.shields.io/badge/Go-1.24+-00ADD8E)
![License](https://img.shields.io/badge/license-MIT-green)

Synkro es una herramienta de gestión de memoria portable que funciona como una **extensión de memoria RAM para LLMs**, con:

- 🔍 **Embeddings vectoriales** para búsqueda semántica (TF-IDF + N-grams, 384 dims)
- 🕸️ **Grafo de relaciones** con 6 tipos de conexiones
- 🧹 **Context pruning** inteligente para eliminar ruido
- 📌 **Grounding injection** para evitar alucinaciones
- 🔄 **Session tracking** con Ring Buffer (20 memorias) para evitar repeticiones
- ⚡ **SQLite con FTS5** para búsquedas full-text rápidas
- 🤖 **MCP integrado** para Claude Desktop, VS Code, Cursor, opencode
- 🖥️ **TUI profesional** con Bubble Tea - 3 paneles interactivos
- 🔎 **Búsqueda en tiempo real** - filtra mientras escribes
- 🎨 **Interfaz moderna** - AltScreen, colores, empty states elegantes

## 🚀 Quick Start

### One-line Installation (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/rodascaar/synkro/main/install.sh | sh
```

### Manual Installation

```bash
git clone https://github.com/rodascaar/synkro.git
cd synkro
go build -o synkro ./cmd/synkro
./synkro init
```

### Enable Embeddings (Opcional)

```bash
./synkro init --with-models
```

## ✨ Características v2

### TUI Profesional (Bubble Tea + Lipgloss)

```bash
./synkro tui
```

**Layout de 3 paneles:**
- **Panel Izquierdo (Sidebar)**: Filtros por tipo (All, Decisions, Tasks, Notes, Archive)
- **Panel Central (Content)**: Lista de memorias con navegación y búsqueda en tiempo real
- **Panel Derecho (Detail)**: Detalles completos + relaciones explicadas

**Características:**
- ⚡ **Búsqueda en tiempo real** - Filtra mientras escribes (`/`)
- 🎯 **Navegación intuitiva** - Arriba/abajo (`↑/↓` o `j/k`)
- 📋 **Detalles completos** - Muestra todos los campos de memoria
- 🔗 **Relaciones claras** - Explica qué significa cada relación
- 🎨 **Interfaz elegante** - AltScreen, colores, empty states
- 🚀 **Sin lag** - Actualizaciones instantáneas

### Motor de Búsqueda Inteligente

```go
// Búsqueda semántica con TF-IDF + N-grams
embedMgr := embeddings.NewEmbeddingManager(embeddings.Config{
    ModelType: embeddings.ModelTypeTFIDF,
})
results, _ := repo.HybridSearch(ctx, query, k, filter)
```

### Grafo de Relaciones

```go
// Auto-linking y traversal BFS
graph.AddRelation(ctx, &memory.MemoryRelation{
    SourceID: id1,
    TargetID: id2,
    Type: "extends",
    Strength: 1.0,
})
path, _ := graph.FindPath(ctx, fromID, toID)
```

**Tipos de relaciones:**
- **Related to**: Memorias conectadas
- **Part of**: Parte de un grupo más grande
- **Extends**: Extiende otra memoria
- **Conflicts with**: En conflicto con otra
- **Depends on**: Depende de otra
- **Example of**: Ejemplo de otra memoria

### Context Pruning Inteligente

```go
// Filtra por similitud, densidad y stop-words
pruner := pruner.NewContextPruner()
results := pruner.Prune(results, query)
```

### Session Tracking

```go
// Ring buffer (20 memorias) para evitar repeticiones
session := session.NewSessionTracker()
session.MarkAsDelivered(sessionID, memoryID)
recentDeliveries := session.GetRecentDeliveries(sessionID, 20)
```

## 📦 Arquitectura

```
synkro/
├── cmd/synkro/
│   ├── main.go
│   ├── commands.go (CLI commands with Cobra)
│   ├── update.go (Auto-update with SHA256)
│   └── health.go (Health check)
├── internal/
│   ├── db/
│   │   ├── db.go (Database wrapper + schema)
│   │   ├── migrations.go (Migration system)
│   │   ├── vector.go (sqlite-vec operations)
│   │   └── extensions/ (sqlite-vec por plataforma)
│   ├── errors/
│   │   └── errors.go (Synkro error types)
│   ├── embeddings/
│   │   ├── generator.go (TF-IDF + N-grams)
│   │   └── manager.go (Embedding manager con MiniLM support)
│   ├── memory/
│   │   ├── model.go (Memory, Relation, Embedding models)
│   │   └── repository.go (CRUD + HybridSearch)
│   ├── graph/
│   │   └── graph.go (Relations, BFS path finding)
│   ├── pruner/
│   │   └── pruner.go (Similarity filtering, grounding)
│   ├── session/
│   │   └── tracker.go (Ring buffer, duplicate detection)
│   ├── mcp/
│   │   └── handlers.go (MCP Server methods, no globals)
│   └── tui/
│       ├── model.go (Bubble Tea model - TUI profesional)
│       └── tui.go (Simple TUI - backup)
├── go.mod
└── README.md
```

## 🔧 Uso

### Inicializar Base de Datos

```bash
./synkro init
./synkro init --with-models  # Con embeddings
```

### Agregar Memoria

```bash
./synkro add --title "Título" --content "Contenido" --type note --tags "tag1|tag2"
```

### Listar Memorias

```bash
./synkro list --type note --limit 20
```

### Buscar Memorias

```bash
./synkro search "término de búsqueda"
```

### TUI Interactiva

```bash
./synkro tui
```

**Atajos de teclado:**
- `↑/↓` o `j/k` - Navegar arriba/abajo
- `/` o `s` - Búsqueda en tiempo real
- `g` - Ver grafo de relaciones
- `Enter` - Ver detalles/alternar grafo
- `Esc` - Volver atrás/salir del grafo
- `Ctrl+C` - Salir

### MCP Server

```bash
./synkro mcp
```

**Configuración IDE:**
```json
{
  "mcp_servers": {
    "synkro": {
      "command": "synkro",
      "args": ["mcp"]
    }
  }
}
```

## 🎯 Características Avanzadas

### Embeddings Vectoriales

- **TF-IDF + N-grams**: Genera embeddings de 384 dimensiones
- **Cache inteligente**: Almacenamiento en memoria para búsquedas repetidas
- **Stop words**: Filtra palabras comunes en inglés y español
- **N-grams**: Captura contexto con bigrams y trigrams
- **MiniLM support**: Preparado para all-minilm-l6-v2 (int8)

### Grafo de Relaciones

- **6 tipos de relaciones**: extends, depends_on, conflicts_with, example_of, part_of, related_to
- **Auto-linking**: Conexiones automáticas basadas en similitud
- **BFS path finding**: Encuentra caminos entre memorias
- **Explicaciones claras**: Cada tipo de relación está explicado en la TUI

### Context Pruning

- **Filtrado por similitud**: Umbral configurable (default: 0.5)
- **Filtrado por contenido**: Elimina contenido repetitivo
- **Grounding injection**: Prefijos que indican fuente de información
- **Token counting**: Limita longitud del contexto

### Session Tracking

- **Ring buffer (20)**: Mantiene últimas 20 memorias entregadas
- **Deduplicación de queries**: Evita respuestas repetidas
- **Tracking por sesión**: Soporta múltiples sesiones concurrentes

### TUI Profesional (Bubble Tea)

- **AltScreen**: Pantalla completa (sin scroll secuencial)
- **Búsqueda en tiempo real**: Filtra mientras escribes
- **3 paneles**: Sidebar (filtros), Content (memorias), Detail (seleccionado)
- **Empty states elegantes**: Mensajes cuando no hay datos
- **Barra de ayuda fija**: Atajos de teclado siempre visibles
- **Vi-style navigation**: Compatible con j/k y flechas

## 🚀 Instalación MCP

### opencode

```json
{
  "mcp_servers": {
    "synkro": {
      "command": "synkro",
      "args": ["mcp"]
    }
  }
}
```

### Claude Desktop (macOS)

Editar `~/Library/Application Support/Claude/claude_desktop_config.json`

### Continue.dev

Editar `.continuerc.json`

### Cursor / VS Code

Editar settings JSON con configuración MCP

## 📊 Rendimiento

- **Embeddings**: ~50-100ms por texto (cacheado: ~1-5ms)
- **Hibrid Search**: ~10-50ms para 20 resultados
- **Grafo traversal**: ~5-20ms para paths de profundidad 3
- **Context pruning**: ~5-10ms
- **Session tracking**: ~1-2ms
- **TUI**: Actualizaciones instantáneas (<5ms)

## 🔨 Development

### Build

```bash
go build -o synkro ./cmd/synkro/
```

**Tamaño del binario:** 8.7MB

### Test

```bash
./synkro init
./synkro add --title "Test" --content "Test content" --type note
./synkro list
./synkro search "test"
./synkro tui
```

## 📝 Documentación

- [README.md](./README.md) - Documentación principal
- [INSTALL.md](./INSTALL.md) - Guía de instalación MCP
- [EMBEDDINGS.md](./EMBEDDINGS.md) - Modelos de embeddings
- [TUI.md](./TUI.md) - Guía de la TUI profesional
- [AGENTS.md](./AGENTS.md) - Guía para agentes de IA

## 🤝 Contribuciones

1. Fork el proyecto
2. Crea tu feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push al branch (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📄 Licencia

MIT License - ver archivo [LICENSE](./LICENSE) para detalles

## 🙏 Agradecimientos

- [sqlite-vec](https://github.com/asg017/sqlite-vec) - Extensiones vectoriales para SQLite
- [go-sqlite3](https://github.com/mattn/go-sqlite3) - Driver de SQLite para Go
- [cobra](https://github.com/spf13/cobra) - Framework CLI para Go
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Framework de TUI para Go
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling para terminales en Go

---

**Synkro v2** - Motor de Contexto Inteligente para LLMs

**Build:** 8.7MB
**Estado:** 100% Completado ✅
**TUI:** Profesional con Bubble Tea ✅
**MCP:** Integrado y listo para usar ✅
