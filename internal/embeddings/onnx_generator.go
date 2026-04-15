package embeddings

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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
	tokenizer    *WordPieceTokenizer
}

func findONNXRuntimePath() string {
	paths := []string{}

	switch runtime.GOOS {
	case "darwin":
		paths = []string{
			"/opt/homebrew/lib/libonnxruntime.dylib",
			"/usr/local/lib/libonnxruntime.dylib",
		}
	case "linux":
		paths = []string{
			"/usr/lib/libonnxruntime.so",
			"/usr/local/lib/libonnxruntime.so",
		}
	case "windows":
		paths = []string{
			"C:\\onnxruntime\\lib\\onnxruntime.dll",
		}
	}

	for _, p := range paths {
		if fileExists(p) {
			return p
		}
	}

	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func NewONNXEmbeddingGenerator(modelManager *ModelManager, cache *Cache) (*ONNXEmbeddingGenerator, error) {
	libPath := findONNXRuntimePath()
	if libPath == "" {
		return nil, fmt.Errorf("ONNX Runtime library not found. Install with: brew install onnxruntime (macOS) or apt install libonnxruntime-dev (Linux)")
	}

	ort.SetSharedLibraryPath(libPath)

	if err := ort.InitializeEnvironment(); err != nil {
		return nil, fmt.Errorf("failed to initialize ONNX Runtime: %w", err)
	}

	gen := &ONNXEmbeddingGenerator{
		modelManager: modelManager,
		cache:        cache,
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
	if len(outputShape) >= 3 {
		g.dimension = int(outputShape[2])
	} else if len(outputShape) >= 2 {
		g.dimension = int(outputShape[1])
	} else {
		g.dimension = modelInfo.Dimension
	}
	g.mu.Unlock()

	vocabPath, err := g.modelManager.GetVocabularyPath(modelInfo.Name)
	if err == nil {
		tok, err := NewWordPieceTokenizer(vocabPath, modelInfo.MaxSeqLen)
		if err != nil {
			log.Printf("Warning: failed to load tokenizer: %v", err)
		} else {
			g.tokenizer = tok
		}
	}
	if g.tokenizer == nil {
		return fmt.Errorf("failed to load tokenizer from %s", vocabPath)
	}

	log.Printf("Successfully loaded model: %s (dimension: %d, max_seq: %d)", modelInfo.Name, g.dimension, modelInfo.MaxSeqLen)

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
	if g.tokenizer == nil {
		g.mu.RUnlock()
		return nil, fmt.Errorf("no tokenizer loaded")
	}
	dim := g.dimension
	g.mu.RUnlock()

	inputIDs, attentionMask, tokenTypeIDs := g.tokenizer.Encode(text)
	seqLen := int64(len(inputIDs))

	inputIDsTensor, err := ort.NewTensor(ort.NewShape(1, seqLen), inputIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create input_ids tensor: %w", err)
	}
	defer inputIDsTensor.Destroy()

	attentionMaskTensor, err := ort.NewTensor(ort.NewShape(1, seqLen), attentionMask)
	if err != nil {
		return nil, fmt.Errorf("failed to create attention_mask tensor: %w", err)
	}
	defer attentionMaskTensor.Destroy()

	tokenTypeIDsTensor, err := ort.NewTensor(ort.NewShape(1, seqLen), tokenTypeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create token_type_ids tensor: %w", err)
	}
	defer tokenTypeIDsTensor.Destroy()

	outputShape := ort.NewShape(1, seqLen, int64(dim))
	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		return nil, fmt.Errorf("failed to create output tensor: %w", err)
	}
	defer outputTensor.Destroy()

	inputs := []ort.Value{inputIDsTensor, attentionMaskTensor, tokenTypeIDsTensor}
	outputs := []ort.Value{outputTensor}

	if err := g.session.Run(inputs, outputs); err != nil {
		return nil, fmt.Errorf("failed to run inference: %w", err)
	}

	allHidden := outputTensor.GetData()
	embedding := meanPool(allHidden, int(seqLen), dim)

	if g.cache != nil {
		_ = g.cache.Set(ctx, text, embedding, g.modelInfo.Name)
	}

	return embedding, nil
}

func meanPool(hidden []float32, seqLen, dim int) []float32 {
	embedding := make([]float32, dim)
	count := float32(0)

	for i := 0; i < seqLen; i++ {
		if i == 0 || i == seqLen-1 {
			continue
		}
		offset := i * dim
		for j := 0; j < dim; j++ {
			embedding[j] += hidden[offset+j]
		}
		count++
	}

	if count > 0 {
		for j := 0; j < dim; j++ {
			embedding[j] /= count
		}
	}

	return embedding
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

	tok := NewTokenizer()
	vocab := make(map[string]int)

	lines := strings.Split(string(data), "\n")
	for idx, line := range lines {
		token := strings.TrimSpace(line)
		if token != "" {
			vocab[token] = idx
		}
	}

	tok.SetVocabulary(vocab)
	return tok, nil
}

func DownloadVocabularyFromHuggingFace(modelName, downloadDir string) error {
	modelDir := filepath.Join(downloadDir, modelName)
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return fmt.Errorf("failed to create model directory: %w", err)
	}

	vocabURL := fmt.Sprintf("https://huggingface.co/sentence-transformers/%s/resolve/main/vocab.txt", modelName)
	vocabPath := filepath.Join(modelDir, "vocab.txt")

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", vocabURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download vocabulary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download vocabulary: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(vocabPath)
	if err != nil {
		return fmt.Errorf("failed to create vocab file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		os.Remove(vocabPath)
		return fmt.Errorf("failed to write vocab file: %w", err)
	}

	return nil
}
