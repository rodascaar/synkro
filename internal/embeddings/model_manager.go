package embeddings

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type ModelInfo struct {
	Name         string
	Dimension    int
	URL          string
	VocabURL     string
	Language     string
	Downloaded   bool
	DownloadPath string
	License      string
	FileSize     int64
	LastUpdated  string
	Params       string
	MaxSeqLen    int
	Benchmarks   map[string]float64
	Description  string
}

type ModelManager struct {
	config      *ManagerConfig
	models      map[string]*ModelInfo
	mu          sync.RWMutex
	httpClient  *http.Client
	downloadDir string
}

type ManagerConfig struct {
	DownloadDir    string
	CacheDir       string
	MaxModels      int
	AutoDownload   bool
	PreferredModel string
	TimeoutSeconds int
}

func NewModelManager(config *ManagerConfig) *ModelManager {
	if config == nil {
		config = &ManagerConfig{
			DownloadDir:    "models",
			CacheDir:       "cache",
			MaxModels:      5,
			AutoDownload:   true,
			TimeoutSeconds: 300,
		}
	}

	mm := &ModelManager{
		config:      config,
		models:      make(map[string]*ModelInfo),
		httpClient:  &http.Client{Timeout: time.Duration(config.TimeoutSeconds) * time.Second},
		downloadDir: config.DownloadDir,
	}

	mm.initializeModels()
	return mm
}

func (mm *ModelManager) initializeModels() {
	defaultModels := []*ModelInfo{
		{
			Name:        "all-MiniLM-L6-v2",
			Dimension:   384,
			URL:         "https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2/resolve/main/onnx/model.onnx",
			VocabURL:    "https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2/resolve/main/vocab.txt",
			Language:    "en",
			License:     "Apache 2.0",
			LastUpdated: "2024-01-15",
			Params:      "22.7M",
			MaxSeqLen:   256,
			Benchmarks: map[string]float64{
				"ArguAna": 50.17,
				"Speed":   1.0,
			},
			Description: "Best balance of speed and accuracy for English text",
		},
		{
			Name:        "paraphrase-multilingual-MiniLM-L12-v2",
			Dimension:   384,
			URL:         "https://huggingface.co/sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2/resolve/main/onnx/model.onnx",
			VocabURL:    "https://huggingface.co/sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2/resolve/main/vocab.txt",
			Language:    "multilingual",
			License:     "Apache 2.0",
			LastUpdated: "2024-01-15",
			Params:      "118M",
			MaxSeqLen:   128,
			Benchmarks: map[string]float64{
				"Speed": 0.7,
			},
			Description: "Supports 50+ languages, slightly slower",
		},
		{
			Name:        "stsb-roberta-base-v2",
			Dimension:   768,
			URL:         "https://huggingface.co/sentence-transformers/stsb-roberta-base-v2/resolve/main/onnx/model.onnx",
			VocabURL:    "https://huggingface.co/sentence-transformers/stsb-roberta-base-v2/resolve/main/vocab.txt",
			Language:    "en",
			License:     "Apache 2.0",
			LastUpdated: "2024-01-15",
			Params:      "110M",
			MaxSeqLen:   128,
			Benchmarks: map[string]float64{
				"STS-B": 85.0,
				"Speed": 0.5,
			},
			Description: "Higher accuracy but slower, English only",
		},
	}

	for _, model := range defaultModels {
		mm.models[model.Name] = model
	}

	mm.checkDownloadedModels()
}

func (mm *ModelManager) checkDownloadedModels() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	for name, model := range mm.models {
		modelDir := filepath.Join(mm.downloadDir, name)
		modelPath := filepath.Join(modelDir, name+".onnx")
		if info, err := os.Stat(modelPath); err == nil {
			model.Downloaded = true
			model.DownloadPath = modelPath
			model.FileSize = info.Size()
		}
	}
}

func (mm *ModelManager) ListModels() []*ModelInfo {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	models := make([]*ModelInfo, 0, len(mm.models))
	for _, model := range mm.models {
		models = append(models, model)
	}
	return models
}

func (mm *ModelManager) GetModel(name string) (*ModelInfo, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	model, exists := mm.models[name]
	if !exists {
		return nil, fmt.Errorf("model %s not found", name)
	}

	return model, nil
}

func (mm *ModelManager) DownloadModel(ctx context.Context, name string, progressCallback func(float64)) error {
	mm.mu.Lock()
	model, exists := mm.models[name]
	if !exists {
		mm.mu.Unlock()
		return fmt.Errorf("model %s not found", name)
	}
	mm.mu.Unlock()

	if model.Downloaded {
		return nil
	}

	if err := os.MkdirAll(mm.downloadDir, 0755); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	modelDir := filepath.Join(mm.downloadDir, name)
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return fmt.Errorf("failed to create model directory: %w", err)
	}

	tempPath := filepath.Join(modelDir, name+".tmp")
	finalPath := filepath.Join(modelDir, name+".onnx")

	req, err := http.NewRequestWithContext(ctx, "GET", model.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := mm.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download model: HTTP %d", resp.StatusCode)
	}

	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = file.Close() }()

	totalSize := resp.ContentLength
	var downloaded int64

	progressTicker := make(chan int64)
	go func() {
		for size := range progressTicker {
			downloaded = size
			if totalSize > 0 && progressCallback != nil {
				progressCallback(float64(downloaded) / float64(totalSize))
			}
		}
	}()

	_, err = io.Copy(file, io.TeeReader(resp.Body, &progressWriter{
		progress: progressTicker,
	}))
	close(progressTicker)

	if err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to write model file: %w", err)
	}

	_ = file.Close()

	if err := os.Rename(tempPath, finalPath); err != nil {
		return fmt.Errorf("failed to rename model file: %w", err)
	}

	if model.VocabURL != "" {
		if err := mm.downloadVocab(ctx, model, modelDir); err != nil {
			return fmt.Errorf("failed to download vocabulary: %w", err)
		}
	}

	mm.mu.Lock()
	model.Downloaded = true
	model.DownloadPath = finalPath
	model.FileSize = downloaded
	mm.mu.Unlock()

	return nil
}

func (mm *ModelManager) downloadVocab(ctx context.Context, model *ModelInfo, modelDir string) error {
	if model.VocabURL == "" {
		return nil
	}

	vocabPath := filepath.Join(modelDir, "vocab.txt")

	if _, err := os.Stat(vocabPath); err == nil {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", model.VocabURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create vocab request: %w", err)
	}

	resp, err := mm.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download vocabulary: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download vocabulary: HTTP %d", resp.StatusCode)
	}

	file, err := os.Create(vocabPath)
	if err != nil {
		return fmt.Errorf("failed to create vocab file: %w", err)
	}
	defer func() { _ = file.Close() }()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to write vocab file: %w", err)
	}

	return nil
}

func (mm *ModelManager) DownloadVocabulary(ctx context.Context, modelName string) error {
	model, err := mm.GetModel(modelName)
	if err != nil {
		return err
	}

	modelDir := filepath.Join(mm.downloadDir, modelName)
	return mm.downloadVocab(ctx, model, modelDir)
}

type progressWriter struct {
	progress chan<- int64
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	pw.progress <- int64(len(p))
	return len(p), nil
}

func (mm *ModelManager) DeleteModel(name string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	model, exists := mm.models[name]
	if !exists {
		return fmt.Errorf("model %s not found", name)
	}

	if !model.Downloaded {
		return nil
	}

	modelDir := filepath.Join(mm.downloadDir, name)
	if err := os.RemoveAll(modelDir); err != nil {
		return fmt.Errorf("failed to delete model directory: %w", err)
	}

	model.Downloaded = false
	model.DownloadPath = ""
	model.FileSize = 0

	return nil
}

func (mm *ModelManager) GetPreferredModel() (*ModelInfo, error) {
	if mm.config.PreferredModel != "" {
		return mm.GetModel(mm.config.PreferredModel)
	}

	for _, model := range mm.models {
		if model.Downloaded {
			return model, nil
		}
	}

	return nil, fmt.Errorf("no downloaded models available")
}

func (mm *ModelManager) AutoDownloadPreferredModel(ctx context.Context) error {
	if !mm.config.AutoDownload {
		return nil
	}

	modelName := mm.config.PreferredModel
	if modelName == "" {
		modelName = "paraphrase-multilingual-MiniLM-L12-v2"
	}

	model, err := mm.GetModel(modelName)
	if err != nil {
		return err
	}

	if model.Downloaded {
		return nil
	}

	return mm.DownloadModel(ctx, modelName, nil)
}

func (mm *ModelManager) GetTotalDiskUsage() (int64, error) {
	var total int64

	for _, model := range mm.models {
		if model.Downloaded && model.FileSize > 0 {
			total += model.FileSize
		}
	}

	return total, nil
}

func (mm *ModelManager) CleanupOldModels(ctx context.Context, keepCount int) error {
	if keepCount <= 0 {
		keepCount = mm.config.MaxModels
	}

	type modelInfo struct {
		name       string
		downloaded bool
	}

	var downloadedModels []modelInfo
	for name, model := range mm.models {
		if model.Downloaded {
			downloadedModels = append(downloadedModels, modelInfo{
				name:       name,
				downloaded: true,
			})
		}
	}

	if len(downloadedModels) <= keepCount {
		return nil
	}

	for i := 0; i < len(downloadedModels)-keepCount; i++ {
		if err := mm.DeleteModel(downloadedModels[i].name); err != nil {
			return fmt.Errorf("failed to delete model %s: %w", downloadedModels[i].name, err)
		}
	}

	return nil
}

func (mm *ModelManager) ValidateModel(name string) error {
	model, err := mm.GetModel(name)
	if err != nil {
		return err
	}

	if !model.Downloaded {
		return fmt.Errorf("model %s is not downloaded", name)
	}

	modelDir := filepath.Join(mm.downloadDir, name)
	modelPath := filepath.Join(modelDir, name+".onnx")
	vocabPath := filepath.Join(modelDir, "vocab.txt")

	if _, err := os.Stat(modelPath); err != nil {
		mm.mu.Lock()
		model.Downloaded = false
		model.DownloadPath = ""
		model.FileSize = 0
		mm.mu.Unlock()
		return fmt.Errorf("model file not found: %w", err)
	}

	if model.VocabURL != "" {
		if _, err := os.Stat(vocabPath); err != nil {
			return fmt.Errorf("vocabulary file not found: %w", err)
		}
	}

	return nil
}

func (mm *ModelManager) GetVocabularyPath(name string) (string, error) {
	model, err := mm.GetModel(name)
	if err != nil {
		return "", err
	}

	if !model.Downloaded {
		return "", fmt.Errorf("model %s is not downloaded", name)
	}

	modelDir := filepath.Join(mm.downloadDir, name)
	vocabPath := filepath.Join(modelDir, "vocab.txt")

	if model.VocabURL == "" {
		return "", fmt.Errorf("model %s has no vocabulary file", name)
	}

	if _, err := os.Stat(vocabPath); err != nil {
		return "", fmt.Errorf("vocabulary file not found: %w", err)
	}

	return vocabPath, nil
}

func (mm *ModelManager) GetSystemInfo() map[string]interface{} {
	totalSize, _ := mm.GetTotalDiskUsage()
	downloadedCount := 0
	for _, model := range mm.models {
		if model.Downloaded {
			downloadedCount++
		}
	}

	return map[string]interface{}{
		"os":            runtime.GOOS,
		"arch":          runtime.GOARCH,
		"download_dir":  mm.downloadDir,
		"total_models":  len(mm.models),
		"downloaded":    downloadedCount,
		"total_size":    totalSize,
		"auto_download": mm.config.AutoDownload,
	}
}
