package pruner

import (
	"testing"

	"github.com/rodascaar/synkro/internal/memory"
	"github.com/stretchr/testify/assert"
)

func TestNewContextPruner(t *testing.T) {
	p := NewContextPruner()
	assert.NotNil(t, p)
	assert.Equal(t, 0.5, p.similarityThreshold)
	assert.Equal(t, 4000, p.maxTokens)
	assert.NotEmpty(t, p.stopWords)
	assert.True(t, p.stopWords["the"])
	assert.True(t, p.stopWords["el"])
}

func TestPrune_Empty(t *testing.T) {
	p := NewContextPruner()
	result := p.Prune(nil, "test")
	assert.Empty(t, result)

	result = p.Prune([]*memory.HybridSearchResult{}, "test")
	assert.Empty(t, result)
}

func TestPrune_FiltersBySimilarity(t *testing.T) {
	p := NewContextPruner()

	results := []*memory.HybridSearchResult{
		{Memory: &memory.Memory{Content: "hello world test query"}, VectorScore: 0.9},
		{Memory: &memory.Memory{Content: "unrelated content"}, VectorScore: 0.3},
		{Memory: &memory.Memory{Content: "test query match"}, VectorScore: 0.7},
	}

	pruned := p.Prune(results, "test query")
	assert.Len(t, pruned, 2)
	assert.Equal(t, 0.9, pruned[0].VectorScore)
	assert.Equal(t, 0.7, pruned[1].VectorScore)
}

func TestPrune_FiltersByTokenBudget(t *testing.T) {
	p := NewContextPruner()

	bigContent := make([]byte, 0)
	for i := 0; i < 5000; i++ {
		bigContent = append(bigContent, "word "...)
	}

	results := []*memory.HybridSearchResult{
		{Memory: &memory.Memory{Content: "test query match one"}, VectorScore: 0.9},
		{Memory: &memory.Memory{Content: string(bigContent)}, VectorScore: 0.8},
		{Memory: &memory.Memory{Content: "test query match two"}, VectorScore: 0.7},
	}

	pruned := p.Prune(results, "test query")
	assert.NotEmpty(t, pruned)
}

func TestPrune_FiltersByContentRelevance(t *testing.T) {
	p := NewContextPruner()

	results := []*memory.HybridSearchResult{
		{Memory: &memory.Memory{Content: "database architecture design patterns"}, VectorScore: 0.9},
		{Memory: &memory.Memory{Content: "the weather is nice today"}, VectorScore: 0.9},
	}

	pruned := p.Prune(results, "database design")
	assert.Len(t, pruned, 1)
	assert.Contains(t, pruned[0].Memory.Content, "database")
}

func TestWithGrounding(t *testing.T) {
	p := NewContextPruner()
	mem := &memory.Memory{
		ID:      "abc123",
		Title:   "Test Title",
		Content: "Test content",
	}

	result := p.WithGrounding(mem)
	assert.Contains(t, result, "abc123")
	assert.Contains(t, result, "Test Title")
	assert.Contains(t, result, "Test content")
	assert.Contains(t, result, "PALACIO MENTAL")
}

func TestIsLowContent_Matches(t *testing.T) {
	p := NewContextPruner()

	assert.False(t, p.isLowContent("database design architecture", "database design"))
	assert.True(t, p.isLowContent("the weather is nice today", "database architecture"))
}

func TestSimilarWords(t *testing.T) {
	p := NewContextPruner()

	assert.True(t, p.similarWords("database", "database"))
	assert.True(t, p.similarWords("database", "data"))
	assert.True(t, p.similarWords("data", "database"))
	assert.False(t, p.similarWords("apple", "orange"))
	assert.True(t, p.similarWords("testing", "test"))
}

func TestCountTokens(t *testing.T) {
	p := NewContextPruner()

	assert.Equal(t, 3, p.countTokens("one two three"))
	assert.Equal(t, 0, p.countTokens(""))
	assert.Equal(t, 5, p.countTokens("hello  world   test  here  now"))
}
