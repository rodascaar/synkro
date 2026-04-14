# 📚 Índice de Documentación - Synkro v2

👋 **Bienvenido a la documentación de Synkro v2**

## 🚀 Comenzar Rápido

Si es la primera vez que usas Synkro, empieza aquí:

1. **[QUICKSTART.md](QUICKSTART.md)** - Guía rápida de 5 minutos
   - Instalación one-shot
   - Primeros pasos
   - TUI básica
   - Ejemplos simples

## 📋 Documentación Principal

### Documentos Esenciales

| Documento | Para quién | Contenido |
|-----------|-----------|-----------|
| **[README.md](README.md)** | Todos | Documentación completa del proyecto, arquitectura y características |
| **[INSTALL.md](INSTALL.md)** | Usuarios | Guía de instalación MCP para opencode, Claude Desktop, Cursor, VS Code, Continue.dev |

### Guías Especializadas

| Documento | Para quién | Contenido |
|-----------|-----------|-----------|
| **[AGENTS.md](AGENTS.md)** | Desarrolladores de IA | Guía completa para integrar Synkro con agentes (Claude, GPT-4, Gemini, etc.) |
| **[EMBEDDINGS.md](EMBEDDINGS.md)** | Usuarios avanzados | Modelos de embeddings disponibles (TF-IDF, MiniLM) y cómo usarlos |
| **[TUI.md](TUI.md)** | Usuarios | Guía completa de la TUI profesional con Bubble Tea |

## 🎯 Por Tópico

### Instalación y Configuración

- **[QUICKSTART.md](QUICKSTART.md)** - Instalación rápida y primeros pasos
- **[INSTALL.md](INSTALL.md)** - Configuración MCP para todos los IDEs
- **[AGENTS.md](AGENTS.md)** - Integración con agentes de IA

### Uso de la TUI

- **[QUICKSTART.md](QUICKSTART.md)** - TUI básica
- **[TUI.md](TUI.md)** - TUI profesional completa
  - Navegación (↑/↓, j/k)
  - Búsqueda en tiempo real (/)
  - Paneles de detalles
  - Relaciones explicadas
  - Atajos de teclado

### Búsqueda y Memorias

- **[README.md](README.md)** - Búsqueda híbrida, embeddings, FTS5
- **[EMBEDDINGS.md](EMBEDDINGS.md)** - Modelos de embeddings y cuándo usarlos
- **[AGENTS.md](AGENTS.md)** - Herramientas MCP para búsqueda y gestión

### Integración con IDEs

- **[INSTALL.md](INSTALL.md)** - Configuración para:
  - opencode
  - Claude Desktop (macOS)
  - Continue.dev
  - Cursor
  - VS Code
  - Cline

## 🔗 Referencias Rápidas

### Atajos de Teclado TUI

| Tecla | Acción |
|-------|--------|
| `↑/↓` o `j/k` | Navegar arriba/abajo |
| `/` o `s` | Buscar |
| `g` | Ver grafo de relaciones |
| `Enter` | Ver detalles |
| `Esc` | Volver atrás |
| `Ctrl+C` | Salir |

### Comandos CLI

| Comando | Descripción |
|---------|-------------|
| `synkro init` | Inicializar base de datos |
| `synkro add` | Agregar memoria |
| `synkro list` | Listar memorias |
| `synkro search` | Buscar memorias |
| `synkro tui` | Lanzar TUI profesional |
| `synkro mcp` | Iniciar servidor MCP |

### Herramientas MCP

| Herramienta | Descripción |
|-------------|-------------|
| `add_memory` | Agregar nueva memoria |
| `get_memory` | Obtener memoria por ID |
| `list_memory` | Listar memorias con filtros |
| `search_memory` | Buscar con FTS5 + embeddings |
| `update_memory` | Actualizar memoria existente |
| `archive_memory` | Archivar memoria |
| `activate_context` | Activar contexto inteligente |

## 🧪 Troubleshooting

### Problemas Comunes

1. **TUI no inicia**
   - **Solución**: `export TERM=xterm-256color && ./synkro tui`

2. **Búsqueda no encuentra**
   - **Solución**: Verificar ortografía, usar términos más cortos

3. **MCP no funciona**
   - **Solución**: Verificar ruta de synkro, reiniciar IDE
   - **Ver más en**: [INSTALL.md](INSTALL.md)

4. **Detalles no muestran**
   - **Solución**: Seleccionar memoria primero con ↑/↓

### Para más ayuda

- Ver [QUICKSTART.md](QUICKSTART.md) - Sección Troubleshooting
- Ver [AGENTS.md](AGENTS.md) - Sección Debugging
- Ver [INSTALL.md](INSTALL.md) - Problemas comunes de MCP

## 📊 Resumen de Características

- ✅ **Búsqueda en tiempo real** - Filtra mientras escribes
- ✅ **3 paneles interactivos** - Sidebar, Content, Detail
- ✅ **Relaciones explicadas** - 6 tipos con significados claros
- ✅ **MCP integrado** - Compatible con todos los IDEs
- ✅ **Embeddings** - TF-IDF + MiniLM support
- ✅ **Grafo** - Traversal BFS, auto-linking
- ✅ **Session tracking** - Ring buffer, deduplicación
- ✅ **Context pruning** - Filtrado inteligente
- ✅ **TUI profesional** - Bubble Tea, AltScreen, colores

## 🚀 Roadmap

- [ ] Embeddings reales con MiniLM (int8)
- [ ] Filtrado por tipo en sidebar
- [ ] Agregar memorias desde TUI
- [ ] Editar memorias desde TUI
- [ ] Exportar/importar memorias
- [ ] Soporte multi-idioma

## 📝 Contribuciones

Para contribuir a Synkro:

1. Fork el proyecto
2. Crea tu feature branch
3. Commit tus cambios
4. Push al branch
5. Abre un Pull Request

Ver más en: https://github.com/rodascaar/synkro

---

**¡Synkro v2 está listo para usar!** 🚀

📞 **¿Necesitas ayuda?**
- Revisa las guías en este índice
- Abre un issue en GitHub
- Revisa los ejemplos en QUICKSTART.md
