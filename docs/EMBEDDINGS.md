# Modelos de Embeddings

Synkro soporta dos tipos de generadores de embeddings:

## 1. TF-IDF + N-grams (Por defecto)

- **Ligero**: Sin dependencias externas
- **Rápido**: Generación instantánea
- **Precisión moderada**: Adecuado para búsquedas simples
- **Tamaño**: 0 MB (embebido en el código)

Este es el modo por defecto y funciona sin configuración adicional.

## 2. MiniLM (all-minilm-l6-v2)

- **Transformador real**: Modelo de HuggingFace
- **Alta precisión**: Mejor comprensión semántica
- **384 dimensiones**: Balance entre tamaño y rendimiento
- **Tamaño**: ~40 MB (modelo cuantizado int8)

### Instalación del Modelo

Para usar MiniLM, ejecuta:

```bash
synkro init --with-models
```

Esto configura el sistema para usar embeddings avanzados.

### Descarga del Modelo

Para funcionalidad completa con MiniLM, descarga el modelo:

```bash
# Opción 1: Descargar manualmente
wget https://huggingface.co/ggerganov/all-minilm-l6-v2-ggml/resolve/main/all-minilm-l6-v2.ggml -O internal/embeddings/models/all-minilm-l6-v2.ggml

# Opción 2: Usar huggingface-cli
huggingface-cli download ggerganov/all-minilm-l6-v2-ggml --local-dir internal/embeddings/models/
```

### Uso en el Código

```go
import "github.com/rodascaar/synkro/internal/embeddings"

// Crear manager con TF-IDF
mgr, err := embeddings.NewEmbeddingManager(embeddings.Config{
    ModelType: embeddings.ModelTypeTFIDF,
})

// Crear manager con MiniLM
mgr, err := embeddings.NewEmbeddingManager(embeddings.Config{
    ModelType:   embeddings.ModelTypeMiniLM,
    EmbedOnInit: true,
    EmbedDim:    384,
})
```

## Comparación de Rendimiento

| Característica | TF-IDF | MiniLM |
|--------------|--------|--------|
| Velocidad     | ⚡⚡⚡   | ⚡⚡    |
| Precisión     | ⭐⭐    | ⭐⭐⭐⭐  |
| Tamaño        | 0 MB   | 40 MB  |
| Configuración | Ninguna | Descarga modelo |

## Futuros Modelos

Planes para agregar soporte adicional:
- **all-MiniLM-L12-v2**: Más preciso pero más lento
- **bge-small-en-v1.5**: Modelo multilingüe moderno
- **nomic-embed-text-v1**: Embeddings para RAG

## Problemas Comunes

**Error al cargar modelo MiniLM**: Asegúrate de descargar el archivo .ggml y colocarlo en `internal/embeddings/models/`

**Búsqueda no funciona**: Verifica que el generador de embeddings esté habilitado en el repository

**Rendimiento lento**: Considera usar TF-IDF para datasets pequeños o prototipos rápidos
