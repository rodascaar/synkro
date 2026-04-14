package embeddings

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
)

type ONNXEmbeddingGenerator struct {
	mu           sync.RWMutex
	session      *ort.DynamicAdvancedSession
	modelInfo    *ModelInfo
	cache        *Cache
	modelManager *ModelManager
	dimension    int
	inputNames   []string
	outputNames  []string
	tokenizer    *Tokenizer
}

func NewONNXEmbeddingGenerator(modelManager *ModelManager, cache *Cache) (*ONNXEmbeddingGenerator, error) {
	ort.SetSharedLibraryPath("")

	if err := ort.InitializeEnvironment(); err != nil {
		return nil, fmt.Errorf("failed to initialize ONNX Runtime: %w", err)
	}

	gen := &ONNXEmbeddingGenerator{
		modelManager: modelManager,
		cache:        cache,
		tokenizer: &Tokenizer{
			vocabulary:   make(map[string]int),
			reverseVocab: make(map[int]string),
		},
	}

	if err := gen.loadModel(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to load model: %w", err)
	}

	return gen, nil
}

func (g *ONNXEmbeddingGenerator) loadModel(ctx context.Context) error {
	modelInfo, err := g.modelManager.GetPreferredModel()
	if err != nil {
		return fmt.Errorf("failed to get preferred model: %w", err)
	}

	if !modelInfo.Downloaded {
		if err := g.modelManager.AutoDownloadPreferredModel(ctx); err != nil {
			return fmt.Errorf("failed to auto-download model: %w", err)
		}

		modelInfo, err = g.modelManager.GetPreferredModel()
		if err != nil {
			return fmt.Errorf("failed to get preferred model after download: %w", err)
		}
	}

	if err := g.modelManager.ValidateModel(modelInfo.Name); err != nil {
		return fmt.Errorf("model validation failed: %w", err)
	}

	inputInfo, outputInfo, err := ort.GetInputOutputInfo(modelInfo.DownloadPath)
	if err != nil {
		return fmt.Errorf("failed to get input/output info: %w", err)
	}

	if len(inputInfo) == 0 {
		return fmt.Errorf("model has no inputs")
	}
	if len(outputInfo) == 0 {
		return fmt.Errorf("model has no outputs")
	}

	g.inputNames = make([]string, len(inputInfo))
	g.outputNames = make([]string, len(outputInfo))
	for i, info := range inputInfo {
		g.inputNames[i] = info.Name
	}
	for i, info := range outputInfo {
		g.outputNames[i] = info.Name
	}

	session, err := ort.NewDynamicAdvancedSession(modelInfo.DownloadPath, g.inputNames, g.outputNames, nil)
	if err != nil {
		return fmt.Errorf("failed to create ONNX session: %w", err)
	}

	if g.session != nil {
		g.session.Destroy()
	}

	g.mu.Lock()
	g.session = session
	g.modelInfo = modelInfo

	outputShape := outputInfo[0].Dimensions
	if len(outputShape) >= 2 {
		g.dimension = int(outputShape[1])
	} else {
		g.dimension = modelInfo.Dimension
	}
	g.mu.Unlock()

	log.Printf("Successfully loaded model: %s (dimension: %d)", modelInfo.Name, g.dimension)

	return nil
}

func (g *ONNXEmbeddingGenerator) Generate(ctx context.Context, text string) ([]float32, error) {
	if g.cache != nil {
		if cached, exists := g.cache.Get(ctx, text); exists {
			return cached, nil
		}
	}

	g.mu.RLock()
	if g.session == nil {
		g.mu.RUnlock()
		return nil, fmt.Errorf("no model loaded")
	}
	dim := g.dimension
	g.mu.RUnlock()

	tokens := g.tokenize(text)
	if len(tokens) == 0 {
		return make([]float32, dim), nil
	}

	inputShape := ort.NewShape(1, int64(len(tokens)))
	inputTensor, err := ort.NewTensor(inputShape, g.tokensToFloat32(tokens))
	if err != nil {
		return nil, fmt.Errorf("failed to create input tensor: %w", err)
	}
	defer inputTensor.Destroy()

	outputShape := ort.NewShape(1, int64(dim))
	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		return nil, fmt.Errorf("failed to create output tensor: %w", err)
	}
	defer outputTensor.Destroy()

	inputs := []ort.Value{inputTensor}
	outputs := []ort.Value{outputTensor}

	if err := g.session.Run(inputs, outputs); err != nil {
		return nil, fmt.Errorf("failed to run inference: %w", err)
	}

	embedding := outputTensor.GetData()
	if len(embedding) == 0 {
		return nil, fmt.Errorf("no output from model")
	}

	if g.cache != nil {
		_ = g.cache.Set(ctx, text, embedding, g.modelInfo.Name)
	}

	return embedding, nil
}

func (g *ONNXEmbeddingGenerator) GenerateBatch(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := g.Generate(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

func (g *ONNXEmbeddingGenerator) Dimension() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.dimension
}

func (g *ONNXEmbeddingGenerator) tokensToFloat32(tokens []string) []float32 {
	data := make([]float32, len(tokens))
	for i, token := range tokens {
		if idx, exists := g.tokenizer.vocabulary[token]; exists {
			data[i] = float32(idx)
		} else {
			data[i] = 0
		}
	}
	return data
}

func (g *ONNXEmbeddingGenerator) tokenize(text string) []string {
	text = strings.ToLower(text)
	text = strings.TrimSpace(text)

	var tokens []string
	words := strings.Fields(text)

	for _, word := range words {
		word = strings.Trim(word, ".,!?;:\"'()[]{}")
		if len(word) == 0 {
			continue
		}

		if _, exists := g.tokenizer.vocabulary[word]; exists {
			tokens = append(tokens, word)
		} else {
			tokens = append(tokens, "[UNK]")
		}
	}

	if len(tokens) == 0 {
		tokens = append(tokens, "[UNK]")
	}

	if len(tokens) > 512 {
		tokens = tokens[:512]
	}

	return tokens
}

func (g *ONNXEmbeddingGenerator) SwitchModel(ctx context.Context, modelName string) error {
	if err := g.modelManager.ValidateModel(modelName); err != nil {
		return fmt.Errorf("model validation failed: %w", err)
	}

	if !g.modelManager.models[modelName].Downloaded {
		if err := g.modelManager.DownloadModel(ctx, modelName, nil); err != nil {
			return fmt.Errorf("failed to download model: %w", err)
		}
	}

	g.modelManager.config.PreferredModel = modelName

	return g.loadModel(ctx)
}

func (g *ONNXEmbeddingGenerator) GetModelInfo() *ModelInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.modelInfo
}

func (g *ONNXEmbeddingGenerator) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.session != nil {
		g.session.Destroy()
		g.session = nil
	}

	return nil
}

func (g *ONNXEmbeddingGenerator) ModelType() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.modelInfo != nil {
		return "onnx:" + g.modelInfo.Name
	}
	return "onnx"
}

func LoadVocabularyFromJSON(jsonPath string) (*Tokenizer, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read vocabulary file: %w", err)
	}

	tokenizer := &Tokenizer{
		vocabulary:   make(map[string]int),
		reverseVocab: make(map[int]string),
	}

	lines := strings.Split(string(data), "\n")
	for idx, line := range lines {
		token := strings.TrimSpace(line)
		if token == "" {
			continue
		}
		tokenizer.vocabulary[token] = idx
		tokenizer.reverseVocab[idx] = token
	}

	return tokenizer, nil
}

func (g *ONNXEmbeddingGenerator) LoadTokenizer(vocabularyPath string) error {
	tokenizer, err := LoadVocabularyFromJSON(vocabularyPath)
	if err != nil {
		return fmt.Errorf("failed to load tokenizer: %w", err)
	}

	g.mu.Lock()
	g.tokenizer = tokenizer
	g.mu.Unlock()

	return nil
}

func DownloadVocabularyFromHuggingFace(modelName, downloadDir string) error {
	modelPath := filepath.Join(downloadDir, modelName)
	vocabPath := filepath.Join(modelPath, "vocab.txt")

	if err := os.MkdirAll(modelPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if _, err := os.Stat(vocabPath); err == nil {
		return nil
	}

	return nil
}
