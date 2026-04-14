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
	assert.Contains(t, vec, float32(0))
}

func TestTFIDFEmbeddingGenerator_Cache(t *testing.T) {
	gen := embeddings.NewTFIDFEmbeddingGenerator(nil)

	text := "Test caching"

	vec1, err1 := gen.Generate(context.Background(), text)
	require.NoError(t, err1)

	vec2, err2 := gen.Generate(context.Background(), text)
	require.NoError(t, err2)

	assert.Equal(t, vec1, vec2)
}

func TestCosineSimilarity(t *testing.T) {
	gen := embeddings.NewTFIDFEmbeddingGenerator(nil)

	text1 := "Similar text"
	text2 := "Similar text"
	text3 := "Completely different"

	vec1, _ := gen.Generate(context.Background(), text1)
	vec2, _ := gen.Generate(context.Background(), text2)
	vec3, _ := gen.Generate(context.Background(), text3)

	sim1 := embeddings.CosineSimilarity(vec1, vec2)
	sim2 := embeddings.CosineSimilarity(vec1, vec3)

	assert.InDelta(t, 1.0, sim1, 0.01)
	assert.Less(t, sim2, sim1)
}

func TestSerializeDeserializeEmbedding(t *testing.T) {
	vec := []float32{0.1, 0.2, 0.3, 0.4, 0.5}

	serialized := embeddings.SerializeEmbedding(vec)
	assert.NotNil(t, serialized)

	deserialized := embeddings.DeserializeEmbedding(serialized)
	assert.NotNil(t, deserialized)
	assert.Equal(t, vec, deserialized)
}
