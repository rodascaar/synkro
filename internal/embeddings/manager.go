package embeddings

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

type ModelType string

const (
	ModelTypeTFIDF  ModelType = "tfidf"
	ModelTypeMiniLM ModelType = "minilm"
)

type Config struct {
	ModelType   ModelType
	ModelPath   string
	EmbedOnInit bool
	EmbedDim    int
	DB          *sql.DB
}

type EmbeddingManager struct {
	config  Config
	tfidf   *TFIDFEmbeddingGenerator
	enabled bool
	mu      sync.RWMutex
	cache   *Cache
}

func NewEmbeddingManager(config Config) (*EmbeddingManager, error) {
	mgr := &EmbeddingManager{
		config: config,
	}

	if config.ModelType == "" {
		config.ModelType = ModelTypeTFIDF
	}

	if config.EmbedDim == 0 {
		config.EmbedDim = EmbeddingDimension
	}

	if config.DB != nil {
		mgr.cache = NewCache(config.DB, 1000)
	}

	switch config.ModelType {
	case ModelTypeTFIDF:
		mgr.tfidf = NewTFIDFEmbeddingGenerator(mgr.cache)
		mgr.enabled = true
		return mgr, nil

	case ModelTypeMiniLM:
		if !config.EmbedOnInit {
			mgr.enabled = false
			return mgr, nil
		}

		fmt.Println("MiniLM embeddings require external dependencies")
		fmt.Println("For now, falling back to TF-IDF")
		mgr.tfidf = NewTFIDFEmbeddingGenerator(mgr.cache)
		mgr.enabled = true
		return mgr, nil

	default:
		return nil, fmt.Errorf("unsupported model type: %s", config.ModelType)
	}
}

func (m *EmbeddingManager) Enable() error {
	if m.enabled {
		return nil
	}

	m.mu.Lock()
	m.enabled = true
	m.mu.Unlock()

	return nil
}

func (m *EmbeddingManager) Generate(ctx context.Context, text string) ([]float32, error) {
	if !m.enabled {
		return nil, fmt.Errorf("embedding manager not enabled")
	}

	switch m.config.ModelType {
	case ModelTypeTFIDF:
		return m.tfidf.Generate(ctx, text)
	case ModelTypeMiniLM:
		return m.tfidf.Generate(ctx, text)
	default:
		return nil, fmt.Errorf("unsupported model type: %s", m.config.ModelType)
	}
}

func (m *EmbeddingManager) GenerateBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if !m.enabled {
		return nil, fmt.Errorf("embedding manager not enabled")
	}

	switch m.config.ModelType {
	case ModelTypeTFIDF:
		return m.tfidf.GenerateBatch(ctx, texts)
	case ModelTypeMiniLM:
		return m.tfidf.GenerateBatch(ctx, texts)
	default:
		return nil, fmt.Errorf("unsupported model type: %s", m.config.ModelType)
	}
}

func (m *EmbeddingManager) Dimension() int {
	switch m.config.ModelType {
	case ModelTypeTFIDF:
		return m.tfidf.Dimension()
	case ModelTypeMiniLM:
		return m.config.EmbedDim
	default:
		return EmbeddingDimension
	}
}

func (m *EmbeddingManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
	return nil
}
