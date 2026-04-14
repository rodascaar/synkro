package embeddings_test

import (
	"context"
	"testing"

	"github.com/rodascaar/synkro/internal/embeddings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTFIDFEmbeddingGenerator_Generate(t *testing.T) {
	gen := embeddings.NewTFIDFEmbeddingGenerator(nil)

	text := "This is a test document about embeddings"
	vec, err := gen.Generate(context.Background(), text)

	assert.NoError(t, err)
	assert.Len(t, vec, embeddings.EmbeddingDimension)
	assert.NotNil(t, vec)
}

func TestTFIDFEmbeddingGenerator_Cache(t *testing.T) {
	gen := embeddings.NewTFIDFEmbeddingGenerator(nil)

	text := "Test caching"

	vec1, err1 := gen.Generate(context.Background(), text)
	require.NoError(t, err1)

	vec2, err2 := gen.Generate(context.Background(), text)
	require.NoError(t, err2)

	assert.Len(t, vec1, embeddings.EmbeddingDimension)
	assert.Len(t, vec2, embeddings.EmbeddingDimension)
}

func TestCosineSimilarity(t *testing.T) {
	gen1 := embeddings.NewTFIDFEmbeddingGenerator(nil)
	gen2 := embeddings.NewTFIDFEmbeddingGenerator(nil)

	textSimilar := "database connection pooling"
	textDifferent := "cooking recipe ingredients"

	vec1, _ := gen1.Generate(context.Background(), textSimilar)
	vec2, _ := gen2.Generate(context.Background(), textSimilar)
	vec3, _ := gen1.Generate(context.Background(), textDifferent)

	simSame := embeddings.CosineSimilarity(vec1, vec2)
	simDiff := embeddings.CosineSimilarity(vec1, vec3)

	assert.InDelta(t, 1.0, simSame, 0.01)
	assert.Less(t, simDiff, simSame)
}

func TestSerializeDeserializeEmbedding(t *testing.T) {
	vec := []float32{0.1, 0.2, 0.3, 0.4, 0.5}

	serialized := embeddings.SerializeEmbedding(vec)
	assert.NotNil(t, serialized)

	deserialized := embeddings.DeserializeEmbedding(serialized)
	assert.NotNil(t, deserialized)
	assert.Equal(t, vec, deserialized)
}
