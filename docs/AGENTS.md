# Synkro - Guía para Agentes de IA

🤖 **Guía completa para integrar Synkro con agentes de IA como Claude, GPT-4, Gemini, etc.**

## 📋 Resumen

Synkro proporciona un servidor MCP (Model Context Protocol) que permite a los agentes de IA:
- ✅ Acceder a memorias almacenadas
- ✅ Buscar en base de datos vectoriales
- ✅ Crear relaciones automáticas
- ✅ Activar contexto inteligente
- ✅ Mantener sesión tracking
- ✅ Aplicar pruning inteligente

## 🚀 Instalación del Servidor MCP

### Opción 1: One-line Installation

```bash
curl -fsSL https://raw.githubusercontent.com/rodascaar/synkro/main/install.sh | sh
```

### Opción 2: Manual Installation

```bash
git clone https://github.com/rodascaar/synkro.git
cd synkro
go build -o synkro ./cmd/synkro/
```

### Opción 3: Usar Binario Existente

```bash
# Si ya tienes el binario compilado
cd ~/projects/synkro
./synkro init
```

## 🔧 Configuración por IDE

### Claude Desktop (macOS)

Editar archivo: `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "synkro": {
      "command": "/usr/local/bin/synkro",
      "args": ["mcp"]
    }
  }
}
```

**O con ruta absoluta:**
```json
{
  "mcpServers": {
    "synkro": {
      "command": "/Users/home/projects/synkro/synkro",
      "args": ["mcp"]
    }
  }
}
```

### opencode

Agregar a configuración de opencode:

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

### Continue.dev

Editar archivo: `.continuerc.json`

```json
{
  "mcpServers": {
    "synkro": {
      "command": "synkro",
      "args": ["mcp"]
    }
  }
}
```

### Cursor / VS Code

Editar `settings.json`:

```json
{
  "mcp.servers": {
    "synkro": {
      "command": "synkro",
      "args": ["mcp"]
    }
  }
}
```

### Cline

Editar `settings.json`:

```json
{
  "mcp": {
    "synkro": {
      "command": "synkro",
      "args": ["mcp"]
    }
  }
}
```

## 🎯 Herramientas Disponibles

### 1. Add Memory

Agrega una nueva memoria al sistema.

**Parámetros:**
- `type` (string, opcional): Tipo de memoria (note, decision, task, context)
- `title` (string, requerido): Título de la memoria
- `content` (string, requerido): Contenido de la memoria
- `source` (string, opcional): Fuente de la memoria (default: user)
- `tags` (array of strings, opcional): Etiquetas

**Respuesta:**
```json
{
  "success": true,
  "memory_id": "mem-20260410120000-1234",
  "similarity_score": 0.0,
  "embedding_used": "tfidf"
}
```

### 2. Get Memory

Obtiene una memoria por ID.

**Parámetros:**
- `id` (string, requerido): ID de la memoria

**Respuesta:**
```json
{
  "memory": {
    "id": "mem-20260410120000-1234",
    "type": "decision",
    "title": "Usar Bubble Tea",
    "content": "Implementar TUI profesional con Bubble Tea y Lipgloss",
    "source": "user",
    "status": "active",
    "tags": ["ui", "tui"],
    "created_at": "2026-04-10T12:00:00Z",
    "updated_at": "2026-04-10T12:00:00Z"
  },
  "relations": []
}
```

### 3. List Memory

Lista memorias con filtros opcionales.

**Parámetros:**
- `type` (string, opcional): Filtrar por tipo
- `status` (string, opcional): Filtrar por estado (active, archived)
- `limit` (number, opcional): Límite de resultados (default: 50)

**Respuesta:**
```json
{
  "memories": [
    {
      "id": "mem-20260410120000-1234",
      "type": "decision",
      "title": "Usar Bubble Tea",
      "content": "Implementar TUI profesional con Bubble Tea y Lipgloss",
      "source": "user",
      "status": "active",
      "tags": ["ui", "tui"],
      "created_at": "2026-04-10T12:00:00Z",
      "updated_at": "2026-04-10T12:00:00Z"
    }
  ],
  "count": 1,
  "timestamp": "2026-04-10T12:00:00Z"
}
```

### 4. Search Memory

Busca memorias con FTS5 + embeddings híbridos.

**Parámetros:**
- `query` (string, requerido): Término de búsqueda
- `type` (string, opcional): Filtrar por tipo
- `status` (string, opcional): Filtrar por estado
- `limit` (number, opcional): Límite de resultados (default: 20)

**Respuesta:**
```json
{
  "query": "TUI",
  "memories": [
    {
      "id": "mem-20260410120000-1234",
      "type": "decision",
      "title": "Usar Bubble Tea",
      "content": "Implementar TUI profesional con Bubble Tea y Lipgloss",
      "source": "user",
      "status": "active",
      "tags": ["ui", "tui"],
      "created_at": "2026-04-10T12:00:00Z",
      "updated_at": "2026-04-10T12:00:00Z"
    }
  ],
  "count": 1,
  "timestamp": "2026-04-10T12:00:00Z"
}
```

### 5. Update Memory

Actualiza una memoria existente.

**Parámetros:**
- `id` (string, requerido): ID de la memoria
- `title` (string, opcional): Nuevo título
- `content` (string, opcional): Nuevo contenido
- `status` (string, opcional): Nuevo estado
- `tags` (array of strings, opcional): Nuevas etiquetas

**Respuesta:**
```json
{
  "success": true,
  "memory_id": "mem-20260410120000-1234",
  "updated_at": "2026-04-10T12:00:00Z"
}
```

### 6. Archive Memory

Archiva una memoria (cambia estado a "archived").

**Parámetros:**
- `id` (string, requerido): ID de la memoria

**Respuesta:**
```json
{
  "success": true,
  "memory_id": "mem-20260410120000-1234",
  "archived_at": "2026-04-10T12:00:00Z"
}
```

### 7. Activate Context

Activa contexto inteligente con pruning, session tracking y deduplicación.

**Parámetros:**
- `query` (string, requerido): Query de contexto
- `session_id` (string, opcional): ID de sesión (default: default)
- `max_tokens` (number, opcional): Límite de tokens (default: 4000)
- `limit` (number, opcional): Límite de resultados (default: 10)

**Respuesta:**
```json
{
  "query": "¿Cómo implementar TUI?",
  "session_id": "default",
  "duplicate_detected": false,
  "max_tokens": 4000,
  "total_tokens": 1500,
  "primary_results": [
    {
      "memory": {
        "id": "mem-20260410120000-1234",
        "type": "decision",
        "title": "Usar Bubble Tea",
        "content": "Implementar TUI profesional con Bubble Tea y Lipgloss",
        "source": "user",
        "status": "active",
        "tags": ["ui", "tui"],
        "created_at": "2026-04-10T12:00:00Z",
        "updated_at": "2026-04-10T12:00:00Z"
      },
      "similarity": 0.85,
      "confidence": "high",
      "is_reminder": false,
      "source": "hybrid"
    }
  ],
  "low_priority_results": [
    {
      "memory": {
        "id": "mem-20260410120001-5678",
        "type": "task",
        "title": "Pruebas TUI",
        "content": "Verificar navegación, búsqueda y detalles",
        "source": "user",
        "status": "active",
        "tags": [],
        "created_at": "2026-04-10T12:00:01Z",
        "updated_at": "2026-04-10T12:00:01Z"
      },
      "similarity": 0.65,
      "confidence": "medium",
      "is_reminder": true,
      "source": "hybrid"
    }
  ]
}
```

## 🧠 Funciones Inteligentes

### 1. Búsqueda Híbrida

Combina FTS5 (full-text search) con embeddings vectoriales:

- **FTS5**: Búsqueda exacta en título y contenido
- **Embeddings**: Búsqueda semántica con similitud
- **TF-IDF + N-grams**: Genera embeddings de 384 dimensiones
- **Cache inteligente**: Almacena embeddings para búsquedas repetidas

### 2. Session Tracking

Evita repeticiones con Ring Buffer:

- **Ring buffer (20)**: Mantiene últimas 20 memorias entregadas
- **Deduplicación de queries**: Detecta queries repetitivos
- **Tracking por sesión**: Soporta múltiples sesiones concurrentes
- **Priority marking**: Marca resultados como high/low priority

### 3. Context Pruning

Filtra resultados inteligentemente:

- **Por similitud**: Umbral configurable (default: 0.5)
- **Por contenido**: Elimina contenido repetitivo
- **Por densidad**: Prefiere memorias más densas
- **Grounding**: Prefijos que indican fuente de información

### 4. Grafo de Relaciones

Conecta memorias automáticamente:

- **6 tipos de relaciones**: related_to, part_of, extends, conflicts_with, example_of, depends_on
- **Auto-linking**: Conexiones basadas en similitud
- **BFS path finding**: Encuentra caminos entre memorias
- **Explicaciones claras**: Cada tipo está explicado

## 💡 Casos de Uso

### Caso 1: Asistente de Proyecto

```
Usuario: "¿Qué decidimos sobre la arquitectura del proyecto?"

Agente (con Synkro):
1. Activa contexto: synkro_activate_context(query="arquitectura proyecto", max_tokens=2000)
2. Obtiene memorias relevantes sobre arquitectura
3. Usa session tracking para evitar repeticiones
4. Aplica pruning para filtrar ruido
5. Responde con contexto de decisiones anteriores
```

### Caso 2: Tutor de Código

```
Usuario: "¿Cómo implementar la TUI?"

Agente (con Synkro):
1. Busca: synkro_search(query="TUI Bubble Tea")
2. Encuentra memorias sobre implementación de TUI
3. Relaciona con tareas y decisiones relacionadas
4. Proporciona guía con contexto histórico
```

### Caso 3: Documentador Automático

```
Usuario: [completando una tarea]

Agente (con Synkro):
1. Agrega memoria: synkro_add_memory(type="task", title="Implementar TUI", content="Usar Bubble Tea y Lipgloss para crear TUI profesional de 3 paneles")
2. Sistema genera embeddings automáticamente
3. Sistema detecta similitudes con memorias existentes
4. Sistema crea relaciones automáticas
```

## ⚙️ Configuración Avanzada

### Variables de Entorno

```bash
# Ruta de base de datos (default: memory.db)
export SYNKRO_DB_PATH=~/custom/path/memory.db

# Modo debug
export SYNKRO_DEBUG=true

# Límite de tokens (default: 4000)
export SYNKRO_MAX_TOKENS=4000

# Tamaño del ring buffer (default: 20)
export SYNKRO_SESSION_BUFFER=20
```

### Ajustes de Rendimiento

```bash
# Memoria caché para embeddings (default: 1000)
export SYNKRO_CACHE_SIZE=1000

# Umbral de similitud (default: 0.5)
export SYNKRO_SIMILARITY_THRESHOLD=0.5

# Tamaño de embeddings (default: 384)
export SYNKRO_EMBEDDING_DIM=384
```

## 🔍 Debugging

### Ver Logs del Servidor MCP

```bash
# Iniciar con modo debug
SYNKRO_DEBUG=true ./synkro mcp
```

### Verificar Conexión MCP

```bash
# Verificar que synkro esté en PATH
which synkro

# Verificar versión
./synkro --version

# Test conexión
./synkro mcp
```

### Probar Herramientas Individualmente

```bash
# Probar búsqueda
./synkro search "TUI"

# Probar listado
./synkro list --limit 5

# Probar agregar
./synkro add --title "Test MCP" --content "Probando conexión MCP" --type note

# Probar eliminar
./synkro delete <id>
```

## 🚨 Problemas Comunes

### Problema 1: Agente no ve Synkro

**Síntomas:**
- Agente no lista herramientas de Synkro
- Errores de conexión MCP

**Solución:**
1. Verificar que synkro esté en PATH:
   ```bash
   which synkro
   ```

2. Verificar configuración MCP:
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

3. Reiniciar IDE completamente

### Problema 2: Búsqueda no encuentra resultados

**Síntomas:**
- `activate_context` devuelve resultados vacíos
- Búsqueda no encuentra memorias conocidas

**Solución:**
1. Verificar que base de datos tiene memorias:
   ```bash
   ./synkro list
   ```

2. Probar búsqueda manual:
   ```bash
   ./synkro search "término"
   ```

3. Verificar embeddings:
   ```bash
   # Agregar memoria de prueba
   ./synkro add --title "Test" --content "Test" --type note
   ```

### Problema 3: Session tracking no funciona

**Síntomas:**
- Agente repite mismas memorias
- No detecta queries duplicados

**Solución:**
1. Verificar session_id:
   ```json
   {
     "session_id": "mi-sesion-unique"
   }
   ```

2. Usar mismo session_id en todo el contexto
3. Verificar que no se reinicie la sesión

## 📚 Referencias

- [README.md](./README.md) - Documentación completa
- [INSTALL.md](./INSTALL.md) - Guía de instalación
- [EMBEDDINGS.md](./EMBEDDINGS.md) - Modelos de embeddings
- [TUI.md](./TUI.md) - Guía de la TUI profesional

## 🎓 Mejores Prácticas

1. **Session IDs únicos**: Usa session_id diferentes para cada conversación
2. **Max tokens razonables**: Ajusta según el contexto necesario
3. **Pruning inteligente**: No desactivar el pruning para mejor precisión
4. **Relaciones manuales**: Agrega relaciones importantes manualmente
5. **Tags útiles**: Usa tags para agrupar memorias relacionadas
6. **Tipos apropiados**: Usa decision, task, note según el caso de uso
7. **Fuentes claras**: Especifica la fuente (user, agent, meeting, file)

---

**Synkro está listo para potenciar tus agentes de IA!** 🤖✨

Para soporte y contribuciones:
https://github.com/rodascaar/synkro
