package embeddings

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
	ready    bool
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
		mgr.ready = true
		return mgr, nil

	case ModelTypeONNX:
		mgr.tfidf = NewTFIDFEmbeddingGenerator(mgr.cache)
		mgr.enabled = true
		mgr.ready = false

		if config.EmbedOnInit {
			mgr.loadONNX(config)
		}

		return mgr, nil

	default:
		return nil, fmt.Errorf("unsupported model type: %s", config.ModelType)
	}
}

func (m *EmbeddingManager) loadONNX(config Config) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.onnx != nil {
		return
	}

	modelMgr := NewModelManager(&ManagerConfig{
		DownloadDir:    config.ModelPath,
		CacheDir:       "cache",
		MaxModels:      5,
		AutoDownload:   false,
		PreferredModel: config.PreferredModel,
	})

	onnxGen, err := NewONNXEmbeddingGenerator(modelMgr, m.cache)
	if err != nil {
		log.Printf("Failed to initialize ONNX embeddings: %v (using TF-IDF)", err)
		return
	}

	m.onnx = onnxGen
	m.modelMgr = modelMgr
	m.ready = true
	log.Println("ONNX embeddings loaded successfully")
}

func (m *EmbeddingManager) LoadONNXAsync(config Config) {
	go func() {
		m.loadONNX(config)
	}()
}

func (m *EmbeddingManager) IsReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ready
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
		m.mu.RLock()
		if m.onnx != nil {
			defer m.mu.RUnlock()
			return m.onnx.Generate(ctx, text)
		}
		m.mu.RUnlock()
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
		m.mu.RLock()
		if m.onnx != nil {
			defer m.mu.RUnlock()
			return m.onnx.GenerateBatch(ctx, texts)
		}
		m.mu.RUnlock()
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
		m.mu.RLock()
		if m.onnx != nil {
			defer m.mu.RUnlock()
			return m.onnx.Dimension()
		}
		m.mu.RUnlock()
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
