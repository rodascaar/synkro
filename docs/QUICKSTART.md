# Synkro v2 - Quick Start Guide

🚀 **Guía rápida de inicio para Synkro v2**

## 📦 Instalación

### One-line Installation (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/rodascaar/synkro/main/install.sh | sh
```

### Manual Installation

```bash
# Clonar repositorio
git clone https://github.com/rodascaar/synkro.git
cd synkro

# Compilar
go build -o synkro ./cmd/synkro/

# Inicializar base de datos
./synkro init

# Opcional: Habilitar embeddings
./synkro init --with-models
```

## 🎯 Primeros Pasos

### 1. Inicializar Base de Datos

```bash
./synkro init
```

Esto crea `memory.db` con el esquema completo.

### 2. Agregar Memorias

```bash
# Agregar una nota
./synkro add --title "Proyecto Synkro" --content "Implementando motor de contexto inteligente para LLMs" --type note

# Agregar una decisión
./synkro add --title "Usar Bubble Tea" --content "Implementar TUI profesional con Bubble Tea y Lipgloss" --type decision

# Agregar una tarea
./synkro add --title "Pruebas TUI" --content "Verificar navegación, búsqueda y detalles" --type task
```

### 3. Listar Memorias

```bash
# Listar todas
./synkro list

# Filtrar por tipo
./synkro list --type decision --limit 10

# Filtrar por estado
./synkro list --status active --limit 20
```

### 4. Buscar Memorias

```bash
# Buscar por término
./synkro search "Bubble Tea"

# Eliminar memoria
./synkro delete <id>

# Buscar en contenido y título
./synkro search "proyecto"
```

## 🖥️ TUI Interactiva

### Lanzar TUI

```bash
./synkro tui
```

### Navegación

| Tecla | Acción |
|-------|--------|
| `↑` o `k` | Mover arriba |
| `↓` o `j` | Mover abajo |
| `/` o `s` | Buscar |
| `g` | Ver grafo |
| `Enter` | Ver detalles/alternar grafo |
| `Esc` | Volver atrás/salir del grafo |
| `Ctrl+C` | Salir |

### Paneles

**Panel Izquierdo (Sidebar):**
- Filtros: All, Decisions, Tasks, Notes, Archive
- Atajos visibles

**Panel Central (Content):**
- Lista de memorias
- Búsqueda en tiempo real
- Indicador de selección (●)

**Panel Derecho (Detail):**
- Tipo, ID, Fuente, Estado
- Etiquetas (si tiene)
- Contenido completo
- Relaciones explicadas

## 🔍 Búsqueda en Tiempo Real

### Cómo Funciona

1. Presiona `/` o `s` - Aparece caja de búsqueda
2. Escribe mientras buscas - Filtra instantáneamente
3. Busca en TÍTULO y CONTENIDO
4. Presiona `Esc` o `Enter` - Cierra búsqueda

### Ejemplo

```
Presiona "/"
Escribe: "TUI"

Resultado:
● [Decision] Usar Bubble Tea
○ [Task] Pruebas TUI
```

## 🔗 Relaciones

### Tipos de Relaciones

- **Related to**: Conectado con otra memoria
- **Part of**: Parte de un grupo más grande
- **Extends**: Extiende otra memoria
- **Conflicts with**: En conflicto con otra
- **Depends on**: Depende de otra
- **Example of**: Ejemplo de otra memoria

### Ver Relaciones

1. Selecciona una memoria con `↑/↓`
2. Presiona `g` - Muestra panel de relaciones
3. Lee explicación de cada relación
4. Presiona `g` o `Esc` para ocultar

## 🤖 Configuración MCP

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

### Claude Desktop (macOS)

Editar `~/Library/Application Support/Claude/claude_desktop_config.json`:

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

### Continue.dev

Editar `.continuerc.json`:

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

Editar settings JSON con configuración MCP

## 📊 Estado del Sistema

### Verificar Instalación

```bash
# Versión
./synkro version

# Ayuda
./synkro --help

# Comandos disponibles
./synkro --help
```

### Verificar Base de Datos

```bash
# Listar memorias
./synkro list

# Buscar
./synkro search "test"

# Contar memorias
./synkro list | wc -l
```

### Verificar TUI

```bash
# Lanzar TUI
./synkro tui

# Verificar que cargue memorias
# Presiona "/" para buscar
# Presiona "g" para ver grafo
# Presiona "Ctrl+C" para salir
```

## 🧪 Testing

### Pruebas Manuales

1. **Test 1**: Agregar memorias
   ```bash
   ./synkro add --title "Test 1" --content "Contenido" --type note
   ./synkro add --title "Test 2" --content "Contenido" --type decision
   ```

2. **Test 2**: Listar memorias
   ```bash
   ./synkro list
   ```

3. **Test 3**: Buscar
   ```bash
   ./synkro search "Test"
   ```

4. **Test 4**: TUI
   ```bash
   ./synkro tui
   ```
   - Navega con `↑/↓`
   - Busca con `/`
   - Ver grafo con `g`
   - Sal con `Ctrl+C`

## ⚠️ Troubleshooting

### TUI No Inicia

**Problema**: Terminal no soporta AltScreen

**Solución**:
```bash
export TERM=xterm-256color
./synkro tui
```

### Búsqueda No Encuentra

**Problema**: Término de búsqueda incorrecto

**Solución**:
- Verificar ortografía
- Usar términos más cortos
- Busca en título Y contenido

### Detalles No Muestran

**Problema**: No hay memoria seleccionada

**Solución**:
- Navegar primero con `↑/↓`
- Seleccionar una memoria
- Ver detalles en panel derecho

### Relaciones Vacías

**Problema**: La memoria no tiene relaciones

**Solución**:
- Normal para memorias nuevas
- Se crean cuando el sistema detecta similitudes
- O cuando se conectan manualmente

### MCP No Funciona

**Problema**: Configuración incorrecta

**Solución**:
- Verificar comando: `synkro` (ruta completa si es necesario)
- Verificar args: `["mcp"]`
- Reiniciar IDE después de configurar

## 📚 Documentación Adicional

- [README.md](./README.md) - Documentación completa
- [INSTALL.md](./INSTALL.md) - Guía de instalación MCP
- [EMBEDDINGS.md](./EMBEDDINGS.md) - Modelos de embeddings
- [TUI.md](./TUI.md) - Guía detallada de la TUI
## 🎯 Tips de Uso

1. **Búsqueda eficiente**: Usa `/` para filtrar rápidamente
2. **Navegación rápida**: Usa `j/k` estilo Vi
3. **Contexto completo**: Presiona `g` para ver relaciones
4. **Organización**: Usa tipos (decision, task, note) para filtrar
5. **Etiquetas**: Agrega tags para agrupar memorias relacionadas

## 🚀 Próximos Pasos

1. **Explorar la TUI**: Probar navegación y búsqueda
2. **Agregar memorias**: Crear ejemplos reales
3. **Configurar MCP**: Integrar con tu IDE preferido
4. **Probar el sistema**: Usar con proyectos reales

---

**¡Listo para usar Synkro v2!** 🚀

Para soporte, issues o sugerencias, visita:
https://github.com/rodascaar/synkro
