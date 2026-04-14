package embeddings

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

type ModelType string

const (
	ModelTypeTFIDF ModelType = "tfidf"
	ModelTypeONNX  ModelType = "onnx"
)

type Config struct {
	ModelType      ModelType
	ModelPath      string
	EmbedOnInit    bool
	EmbedDim       int
	DB             *sql.DB
	PreferredModel string
}

type EmbeddingManager struct {
	config   Config
	tfidf    *TFIDFEmbeddingGenerator
	onnx     *ONNXEmbeddingGenerator
	modelMgr *ModelManager
	enabled  bool
	mu       sync.RWMutex
	cache    *Cache
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

	case ModelTypeONNX:
		modelMgr := NewModelManager(&ManagerConfig{
			DownloadDir:    config.ModelPath,
			CacheDir:       "cache",
			MaxModels:      5,
			AutoDownload:   config.EmbedOnInit,
			PreferredModel: config.PreferredModel,
		})

		onnxGen, err := NewONNXEmbeddingGenerator(modelMgr, mgr.cache)
		if err != nil {
			fmt.Printf("Failed to initialize ONNX embeddings: %v\n", err)
			fmt.Println("Falling back to TF-IDF")
			mgr.tfidf = NewTFIDFEmbeddingGenerator(mgr.cache)
			mgr.enabled = true
			return mgr, nil
		}

		mgr.onnx = onnxGen
		mgr.modelMgr = modelMgr
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
	case ModelTypeONNX:
		if m.onnx != nil {
			return m.onnx.Generate(ctx, text)
		}
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
	case ModelTypeONNX:
		if m.onnx != nil {
			return m.onnx.GenerateBatch(ctx, texts)
		}
		return m.tfidf.GenerateBatch(ctx, texts)
	default:
		return nil, fmt.Errorf("unsupported model type: %s", m.config.ModelType)
	}
}

func (m *EmbeddingManager) Dimension() int {
	switch m.config.ModelType {
	case ModelTypeTFIDF:
		return m.tfidf.Dimension()
	case ModelTypeONNX:
		if m.onnx != nil {
			return m.onnx.Dimension()
		}
		return m.config.EmbedDim
	default:
		return EmbeddingDimension
	}
}

func (m *EmbeddingManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false

	if m.onnx != nil {
		if err := m.onnx.Close(); err != nil {
			return err
		}
	}

	return nil
}
