package embeddings

import (
	"context"
	"encoding/binary"
	"math"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"github.com/rodascaar/synkro/internal/stopwords"
)

type EmbeddingGenerator interface {
	Generate(ctx context.Context, text string) ([]float32, error)
	GenerateBatch(ctx context.Context, texts []string) ([][]float32, error)
	Dimension() int
}

type TFIDFEmbeddingGenerator struct {
	mu                sync.RWMutex
	dimension         int
	cache             *Cache
	vocabulary        map[string]int
	documentFrequency map[string]int
	totalDocuments    int
	ngramSize         int
	stopWords         map[string]bool
	modelType         string
}

const (
	EmbeddingDimension = 384
)

// defaultStopwords es la lista de stopwords compartida usada por el
// tokenizador exportado TokenizeText.
var defaultStopwords = stopwords.NewSet()

// TokenizeText devuelve los tokens base de un texto: minúsculas, sin
// stopwords y sin duplicados. No incluye n-gramas. Se usa para comparaciones
// por solapamiento (e.g. conflict detection), independientes del estado
// del generador de embeddings.
func TokenizeText(text string) []string {
	text = normalizeText(text)
	text = strings.ToLower(text)
	re := regexp.MustCompile(`[^\w\s]`)
	text = re.ReplaceAllString(text, " ")
	words := strings.Fields(text)

	seen := make(map[string]bool, len(words))
	tokens := make([]string, 0, len(words))
	for _, word := range words {
		if len(word) < 2 || defaultStopwords[word] {
			continue
		}
		if seen[word] {
			continue
		}
		seen[word] = true
		tokens = append(tokens, word)
	}
	return tokens
}

func NewTFIDFEmbeddingGenerator(cache *Cache) *TFIDFEmbeddingGenerator {
	return &TFIDFEmbeddingGenerator{
		dimension:         EmbeddingDimension,
		cache:             cache,
		vocabulary:        make(map[string]int),
		documentFrequency: make(map[string]int),
		totalDocuments:    0,
		ngramSize:         3,
		stopWords:         stopwords.NewSet(),
		modelType:         "tfidf",
	}
}

func (g *TFIDFEmbeddingGenerator) Generate(ctx context.Context, text string) ([]float32, error) {
	if g.cache != nil {
		if cached, exists := g.cache.Get(ctx, text); exists {
			return cached, nil
		}
	}

	tokens := g.tokenize(text)
	embedding := g.generateEmbedding(tokens)

	g.mu.Lock()
	g.updateVocabulary(tokens)
	g.totalDocuments++
	g.mu.Unlock()

	if g.cache != nil {
		_ = g.cache.Set(ctx, text, embedding, g.modelType)
	}

	return embedding, nil
}

func (g *TFIDFEmbeddingGenerator) GenerateBatch(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := g.Generate(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

func (g *TFIDFEmbeddingGenerator) Dimension() int {
	return g.dimension
}

func (g *TFIDFEmbeddingGenerator) tokenize(text string) []string {
	text = normalizeText(text)
	text = strings.ToLower(text)
	re := regexp.MustCompile(`[^\w\s]`)
	text = re.ReplaceAllString(text, " ")
	words := strings.Fields(text)

	var tokens []string
	seen := make(map[string]bool)

	for _, word := range words {
		if len(word) < 2 || g.stopWords[strings.ToLower(word)] {
			continue
		}

		word = strings.ToLower(strings.TrimSpace(word))

		if seen[word] {
			continue
		}
		seen[word] = true
		tokens = append(tokens, word)
	}

	return g.generateNgrams(tokens, g.ngramSize)
}

func (g *TFIDFEmbeddingGenerator) generateNgrams(tokens []string, n int) []string {
	if n <= 1 || len(tokens) < n {
		return tokens
	}

	ngrams := make([]string, 0, len(tokens)+(len(tokens)-1)*(n-1))
	ngrams = append(ngrams, tokens...)

	for i := 0; i < len(tokens)-n+1; i++ {
		ngram := strings.Join(tokens[i:i+n], "_")
		ngrams = append(ngrams, ngram)
	}

	return ngrams
}

func (g *TFIDFEmbeddingGenerator) generateEmbedding(tokens []string) []float32 {
	embedding := make([]float32, g.dimension)

	tokenCounts := make(map[string]int)
	for _, token := range tokens {
		tokenCounts[token]++
	}

	for _, token := range tokens {
		tf := float32(tokenCounts[token]) / float32(len(tokens))
		df := float32(g.documentFrequency[token])
		if df == 0 {
			df = 1
		}
		idf := float32(math.Log(float64(g.totalDocuments+1)/float64(df+1))) + 1
		score := tf * idf

		dim := int(g.hashString(token) % uint32(g.dimension))
		sign := float32(1.0)
		if g.hashSign(token)%2 == 1 {
			sign = -1.0
		}
		embedding[dim] += sign * score
	}

	norm := float32(math.Sqrt(float64(dotProduct(embedding, embedding))))
	if norm > 1e-6 {
		for i := range embedding {
			embedding[i] /= norm
		}
	}

	return embedding
}

func dotProduct(a, b []float32) float32 {
	var sum float32
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func (g *TFIDFEmbeddingGenerator) hashString(s string) uint32 {
	hash := uint32(2166136261)
	for _, c := range s {
		hash ^= uint32(c)
		hash *= uint32(16777619)
	}
	return hash
}

// hashSign devuelve un hash decorrelacionado de hashString para la dirección
// del signo del hashing trick.
func (g *TFIDFEmbeddingGenerator) hashSign(s string) uint32 {
	hash := uint32(0x9E3779B9)
	for _, c := range s {
		hash ^= uint32(c)
		hash *= uint32(0x01000193)
	}
	return hash
}

func (g *TFIDFEmbeddingGenerator) updateVocabulary(tokens []string) {
	for _, token := range tokens {
		if _, exists := g.vocabulary[token]; !exists {
			g.vocabulary[token] = len(g.vocabulary)
		}
		g.documentFrequency[token]++
	}
}

func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct float32
	var normA float32
	var normB float32

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	normA = float32(math.Sqrt(float64(normA)))
	normB = float32(math.Sqrt(float64(normB)))

	if normA < 1e-6 || normB < 1e-6 {
		return 0.0
	}

	return dotProduct / (normA * normB)
}

func SerializeEmbedding(embedding []float32) []byte {
	buf := make([]byte, len(embedding)*4)

	for i, val := range embedding {
		bits := math.Float32bits(val)
		binary.LittleEndian.PutUint32(buf[i*4:], bits)
	}

	return buf
}

func DeserializeEmbedding(data []byte) []float32 {
	if len(data)%4 != 0 {
		return nil
	}

	embedding := make([]float32, len(data)/4)

	for i := 0; i < len(embedding); i++ {
		bits := binary.LittleEndian.Uint32(data[i*4 : (i+1)*4])
		embedding[i] = math.Float32frombits(bits)
	}

	return embedding
}

func normalizeText(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range norm.NFKD.String(s) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
