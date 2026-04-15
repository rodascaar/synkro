package embeddings

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func getProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

func TestONNXInference(t *testing.T) {
	root := getProjectRoot()
	modelMgr := NewModelManager(&ManagerConfig{
		DownloadDir:    filepath.Join(root, "models"),
		PreferredModel: "all-MiniLM-L6-v2",
	})

	gen, err := NewONNXEmbeddingGenerator(modelMgr, nil)
	if err != nil {
		t.Skipf("ONNX Runtime not available or model not downloaded: %v", err)
		return
	}
	defer func() { _ = gen.Close() }()

	ctx := context.Background()

	texts := []string{
		"Hello world",
		"This is a test of the ONNX embedding model",
		"The quick brown fox jumps over the lazy dog",
	}

	for _, text := range texts {
		emb, err := gen.Generate(ctx, text)
		if err != nil {
			t.Fatalf("Failed to generate embedding for '%s': %v", text, err)
		}
		if len(emb) != 384 {
			t.Fatalf("Expected dimension 384, got %d", len(emb))
		}

		var norm float64
		for _, x := range emb {
			norm += float64(x * x)
		}
		if norm == 0 {
			t.Fatalf("Zero embedding for '%s'", text)
		}
		t.Logf("'%s' -> dim=%d, L2=%.4f, first5=%v", text, len(emb), math.Sqrt(norm), emb[:5])
	}

	t.Run("SemanticSimilarity", func(t *testing.T) {
		emb1, _ := gen.Generate(ctx, "The cat sat on the mat")
		emb2, _ := gen.Generate(ctx, "A kitten was sitting on a rug")
		emb3, _ := gen.Generate(ctx, "The stock market crashed today")

		sim12 := cosineSim(emb1, emb2)
		sim13 := cosineSim(emb1, emb3)

		t.Logf("Similarity (cat/kitten): %.4f", sim12)
		t.Logf("Similarity (cat/stock):  %.4f", sim13)

		if sim12 <= sim13 {
			t.Logf("Warning: semantically similar texts should have higher similarity (got %.4f <= %.4f)", sim12, sim13)
		} else {
			t.Logf("Semantic similarity ordering correct!")
		}
	})
}

func cosineSim(a, b []float32) float64 {
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func TestMain(m *testing.M) {
	if os.Getenv("TEST_ONNX") == "" {
		fmt.Println("Skipping ONNX tests (set TEST_ONNX=1 to enable)")
		os.Exit(0)
	}
	os.Exit(m.Run())
}
