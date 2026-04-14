# Documentación del Servidor MCP Synkro

## 🎯 Qué es MCP (Model Context Protocol)

MCP es un protocolo estándar que permite a los LLM (asistentes de IA) acceder a herramientas externas de manera estructurada. Synkro expone 6 herramientas MCP para gestionar memorias.

## 📋 Herramientas MCP Expuestas

### 1. add_memory
Agregar una nueva memoria a la base de datos.

**Entrada:**
```json
{
  "type": "note|decision|context|task|fact|rule|contact",
  "title": "Título corto (requerido)",
  "content": "Contenido detallado",
  "source": "user|agent|meeting|file",
  "tags": ["tag1", "tag2"]
}
```

**Salida:** ID de la memoria creada

---

### 2. get_memory
Obtener una memoria específica por ID.

**Entrada:**
```json
{
  "id": "mem-20260410135310-8414"
}
```

**Salida:** Detalle completo de la memoria

---

### 3. list_memories
Listar memorias con filtros opcionales.

**Entrada:**
```json
{
  "type": "decision",        // Opcional: filtrar por tipo
  "status": "active",        // Opcional: "active" o "archived"
  "source": "agent",        // Opcional: filtrar por fuente
  "limit": 50               // Opcional: límite de resultados (default: 50)
}
```

**Salida:** Lista de memorias

---

### 4. search_memories
Buscar memorias usando FTS5 (full-text search).

**Entrada:**
```json
{
  "query": "búsqueda",       // Requerido: query de búsqueda
  "type": "note",           // Opcional: filtrar por tipo
  "status": "active",       // Opcional: "active" o "archived"
  "limit": 20               // Opcional: límite de resultados (default: 20)
}
```

**Salida:** Lista de memorias que coinciden con la búsqueda

---

### 5. update_memory
Actualizar una memoria existente.

**Entrada:**
```json
{
  "id": "mem-20260410135310-8414",  // Requerido
  "title": "Nuevo título",               // Opcional
  "content": "Nuevo contenido",          // Opcional
  "status": "archived",                 // Opcional
  "tags": ["nuevo", "tag"]             // Opcional
}
```

**Salida:** Confirmación de actualización

---

### 6. archive_memory
Archivar una memoria (marcar como "archived").

**Entrada:**
```json
{
  "id": "mem-20260410135310-8414"
}
```

**Salida:** Confirmación de archivado

---

## 🔧 Cómo Funciona el Servidor MCP

### Configuración del Servidor

**Archivo:** `internal/mcp/types.go`

```go
server, mcpCmd = cobramcp.ServerAndCommand(&cobramcp.Config{
    Name:         "synkro",
    Version:      "1.0.0",
    Instructions: "Synkro memory management - use tools to add, get, list, search, update, and archive memories stored in SQLite",
})
```

**Campos Importantes:**
- `Name`: Nombre del servidor MCP (usado por el cliente)
- `Version`: Versión del servidor
- `Instructions`: **Instrucciones para el LLM** - esto es lo que dice al modelo qué puede hacer

### Handlers de Herramientas

**Archivo:** `internal/mcp/handlers.go`

Cada función handler:
1. Recibe el input validado
2. Interactúa con el repository de memorias
3. Escribe el resultado en `io.Writer`

**Ejemplo:**
```go
func AddMemory(input AddMemoryInput, w io.Writer) error {
    ctx := context.Background()
    mem := &memory.Memory{
        Type:    input.Type,
        Title:   input.Title,
        Content: input.Content,
        Source:  input.Source,
        Status:  "active",
        Tags:    input.Tags,
    }

    if err := globalRepo.Create(ctx, mem); err != nil {
        fmt.Fprintf(w, "Error creating memory: %v\n", err)
        return err
    }

    fmt.Fprintf(w, "Memory created: %s\nType: %s\nTitle: %s", mem.ID, mem.Type, mem.Title)
    return nil
}
```

---

## 📁 Archivo de Configuración del Cliente

### OpenCode
**Ubicación:** `~/.config/opencode/opencode.json`

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "synkro": {
      "type": "local",
      "command": [
        "/Users/home/Downloads/nichogram/synkro/synkro",
        "mcp",
        "--db",
        "/Users/home/Downloads/nichogram/synkro/memory.db"
      ],
      "enabled": true,
      "environment": {
        "DB_PATH": "/Users/home/Downloads/nichogram/synkro/memory.db"
      }
    }
  }
}
```

### Claude Desktop
**Ubicación:** `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "synkro": {
      "command": "/Users/home/Downloads/nichogram/synkro/synkro",
      "args": ["mcp", "--db", "/Users/home/Downloads/nichogram/synkro/memory.db"],
      "description": "Synkro memory management tool with SQLite and MCP support"
    }
  }
}
```

---

## 🚀 Flujo de Comunicación

### 1. Cliente Inicia el Servidor
```
OpenCode → ejecuta: /path/to/synkro mcp --db /path/to/memory.db
```

### 2. Servidor Responde con Protocolo MCP
```
Servidor → Envía: JSON-RPC con lista de herramientas disponibles
```

### 3. Cliente Pregunta al LLM
```
OpenCode → LLM: "Tienes acceso a estas herramientas de Synkro: [...]"
```

### 4. LLM Solicita Ejecutar una Herramienta
```
LLM → OpenCode: "Quiero agregar una memoria"
OpenCode → Servidor: JSON-RPC: call tool "add_memory" con {...}
```

### 5. Servidor Ejecuta y Responde
```
Servidor → Repository: Crear memoria en SQLite
Repository → Servidor: Memoria creada: mem-xxx
Servidor → OpenCode: "Memory created: mem-xxx\nType: decision\nTitle: ..."
```

### 6. LLM Procesa la Respuesta
```
OpenCode → LLM: Memoria creada exitosamente
LLM → Usuario: "He guardado tu memoria: Usar SQLite para Synkro"
```

---

## 🎯 Lo que el LLM Sobre Synkro

### Instrucciones del Servidor
```text
Synkro memory management - use tools to add, get, list, search, update, and archive memories stored in SQLite
```

### Qué Sabe el LLM:

1. **Nombre:** Synkro
2. **Versión:** 1.0.0
3. **Propósito:** Gestión de memorias en SQLite
4. **Herramientas disponibles:**
   - `add_memory`: Crear memorias
   - `get_memory`: Obtener memoria por ID
   - `list_memories`: Listar memorias con filtros
   - `search_memories`: Buscar con FTS5
   - `update_memory`: Actualizar memorias
   - `archive_memory`: Archivar memorias

### Ejemplo de Prompt Generado por OpenCode:

```
System: Tienes acceso a las siguientes herramientas MCP:

Tool: synkro_add_memory
Description: Agregar una nueva memoria a la base de datos
Input Schema:
{
  "type": {"type": "string", "description": "Memory type (note, decision, fact, task, rule, contact)"},
  "title": {"type": "string", "description": "Memory title (short summary)"},
  "content": {"type": "string", "description": "Memory content (detailed information)"},
  "source": {"type": "string", "description": "Memory source (user, agent, meeting, file)"},
  "tags": {"type": "array", "description": "Memory tags (lowercase, separated by |)"}
}

Tool: synkro_search_memories
Description: Buscar memorias usando FTS5 full-text search
Input Schema:
{
  "query": {"type": "string", "description": "Search query (FTS5 full-text search)"},
  "type": {"type": "string", "description": "Filter by memory type"},
  "status": {"type": "string", "description": "Filter by status (active, archived)"},
  "limit": {"type": "integer", "description": "Limit results (default: 20)"}
}

[... resto de herramientas ...]

Instructions: Synkro memory management - use tools to add, get, list, search, update, and archive memories stored in SQLite

User: Guárdame esta información: Decidí usar SQLite en lugar de CSV para Synkro porque FTS5 permite búsqueda full-text más rápida.

Assistant: Usaré la herramienta synkro_add_memory para guardar esto.
→ call synkro_add_memory with: {"type": "decision", "title": "Usar SQLite para Synkro", "content": "Decidí usar SQLite en lugar de CSV para Synkro porque FTS5 permite búsqueda full-text más rápida.", "source": "user", "tags": ["database", "decision"]}

Resultado: Memory created: mem-20260410135310-8414
Type: decision
Title: Usar SQLite para Synkro

Assistant: ✓ He guardado tu memoria: "Usar SQLite para Synkro" (tipo: decision, ID: mem-20260410135310-8414)
```

---

## 🔍 Problemas Conocidos

### 1. Servidor MCP No Inicia
**Síntoma:** `read error: EOF` al iniciar el servidor
**Causa:** Normal cuando no hay cliente conectado
**Solución:** El servidor debe ser iniciado por el cliente (OpenCode, Claude Desktop), no manualmente

### 2. Herramientas No Aparecen en el Cliente
**Causas posibles:**
- Configuración incorrecta en `opencode.json`
- Ruta incorrecta del binario
- Permisos de ejecución faltantes
- Base de datos no accesible

**Verificación:**
```bash
# Verificar configuración
cat ~/.config/opencode/opencode.json | jq '.mcp.synkro'

# Verificar binario
ls -la /Users/home/Downloads/nichogram/synkro/synkro

# Verificar base de datos
ls -la /Users/home/Downloads/nichogram/synkro/memory.db
```

### 3. LLM No Usa las Herramientas
**Causas posibles:**
- `Instructions` muy cortas o confusas
- Schemas JSON mal definidos
- Handler devuelve errores

**Solución:**
- Mejorar las instrucciones del servidor
- Agregar ejemplos en las descripciones de herramientas
- Probar herramientas manualmente con el comando CLI

---

## 📚 Referencias

- **MCP Protocol:** https://modelcontextprotocol.io/
- **cobra-mcp:** https://github.com/eat-pray-ai/cobra-mcp
- **MCP Go SDK:** https://github.com/modelcontextprotocol/go-sdk
