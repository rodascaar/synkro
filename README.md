# Synkro v2 🚀

> **Motor de Contexto Inteligente para LLMs**

![Version](https://img.shields.io/badge/version-2.0-blue)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8E)
![License](https://img.shields.io/badge/license-MIT-green)

Sistema de gestión de memoria con embeddings, grafo de relaciones y pruning inteligente.

## 🚀 Quick Start

### One-line Installation

```bash
curl -fsSL https://raw.githubusercontent.com/rodascaar/synkro/main/install.sh | sh
```

### Quick Commands

```bash
./synkro init           # Inicializar
./synkro add            # Agregar memoria
./synkro list           # Listar
./synkro search         # Buscar
./synkro tui            # TUI profesional
./synkro mcp            # MCP server
```

## 📚 Documentación

👋 **Ver toda la documentación en** [docs/](docs/)

| Documento | Descripción |
|-----------|-------------|
| **[INDEX.md](docs/INDEX.md)** | 📚 Índice completo de documentación |
| **[README.md](docs/README.md)** | 📖 Documentación completa del proyecto |
| **[QUICKSTART.md](docs/QUICKSTART.md)** | ⚡ Guía rápida de 5 minutos |
| **[AGENTS.md](docs/AGENTS.md)** | 🤖 Guía para integrar con agentes de IA |
| **[INSTALL.md](docs/INSTALL.md)** | 🔧 Instalación MCP para todos los IDEs |
| **[EMBEDDINGS.md](docs/EMBEDDINGS.md)** | 🔍 Modelos de embeddings disponibles |
| **[TUI.md](docs/TUI.md)** | 🖥️ Guía completa de la TUI profesional |
| **[CHANGELOG.md](CHANGELOG.md)** | 📋 Historial de cambios |
| **[CONTRIBUTING.md](CONTRIBUTING.md)** | 🤝 Guía para contribuir |

## ✨ Características

- 🔍 **FTS5 Full-Text Search** - Búsqueda con BM25 scoring
- 🧠 **Embeddings semánticos** - TF-IDF + N-grams (384 dims) con cache persistente
- 🕸️ **Grafo de relaciones persistente** - 6 tipos con explicaciones claras
- 🖥️ **TUI profesional** - 3 paneles interactivos + Add Memory form
- 🤖 **MCP integrado (SDK oficial)** - Compatible con todos los IDEs
- 💾 **SQLite + WAL** - Búsquedas full-text rápidas
- 📌 **Session tracking persistente** - Evita repeticiones
- 🧹 **Context pruning** - Filtrado inteligente
- ✅ **Testing completo** - Suite de tests con >90% cobertura
- 🔧 **Variables de entorno** - Configuración flexible
- 🚀 **CI/CD** - GitHub Actions con lint y tests

## 🎯 TUI Professional

```bash
./synkro tui
```

**Atajos:**
- `↑/↓` o `j/k` - Navegar
- `/` - Buscar
- `g` - Ver grafo
- `Ctrl+C` - Salir

## 🔧 MCP Integration

### Configuración Universal

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

**IDEs compatibles:**
- opencode
- Claude Desktop
- Cursor
- VS Code
- Continue.dev
- Cline

Ver más en: [INSTALL.md](docs/INSTALL.md)

## 📦 Installation

### One-line (macOS/Linux)
```bash
curl -fsSL https://raw.githubusercontent.com/rodascaar/synkro/main/install.sh | sh
```

### Manual
```bash
git clone https://github.com/rodascaar/synkro.git
cd synkro
go build -o synkro ./cmd/synkro/
./synkro init
```

Ver más en: [QUICKSTART.md](docs/QUICKSTART.md)

## 📊 Performance

- **Embeddings**: ~50-100ms (cacheado: ~1-5ms)
- **Search**: ~10-50ms para 20 resultados
- **TUI**: <5ms actualizaciones
- **Binary**: 8.7MB

## 📄 License

MIT License - ver [LICENSE](LICENSE)

---

**Synkro v2** - Motor de Contexto Inteligente para LLMs

📚 **Documentación completa:** [docs/](docs/)
🚀 **Estado:** 100% Completado ✅
