package pruner

import (
	"fmt"
	"strings"

	"github.com/rodascaar/synkro/internal/memory"
)

type ContextPruner struct {
	similarityThreshold float64
	maxTokens           int
	stopWords           map[string]bool
}

func NewContextPruner() *ContextPruner {
	stopWords := make(map[string]bool)
	for _, word := range []string{
		"el", "la", "de", "que", "y", "a", "en", "un", "por",
		"con", "no", "una", "su", "para", "es", "del", "los",
		"the", "a", "an", "and", "are", "as", "at", "be", "by",
		"for", "from", "has", "he", "in", "is", "it", "its", "of",
		"on", "that", "the", "to", "was", "were", "will", "with",
	} {
		stopWords[strings.ToLower(word)] = true
	}

	return &ContextPruner{
		similarityThreshold: 0.5,
		maxTokens:           4000,
		stopWords:           stopWords,
	}
}

func (p *ContextPruner) Prune(results []*memory.HybridSearchResult, query string) []*memory.HybridSearchResult {
	if len(results) == 0 {
		return results
	}

	pruned := make([]*memory.HybridSearchResult, 0)
	tokens := 0

	for _, result := range results {
		if result.VectorScore < p.similarityThreshold {
			continue
		}

		content := result.Memory.Content
		if p.isLowContent(content, query) {
			continue
		}

		contentTokens := p.countTokens(content)
		if tokens+contentTokens > p.maxTokens {
			break
		}

		tokens += contentTokens
		pruned = append(pruned, result)
	}

	return pruned
}

func (p *ContextPruner) WithGrounding(mem *memory.Memory) string {
	return fmt.Sprintf("[PALACIO MENTAL: %s - %s]\n%s", mem.ID, mem.Title, mem.Content)
}

func (p *ContextPruner) isLowContent(content, query string) bool {
	words := strings.Fields(strings.ToLower(content))
	queryWords := strings.Fields(strings.ToLower(query))

	matches := 0
	for _, qword := range queryWords {
		for _, word := range words {
			if p.similarWords(word, qword) {
				matches++
				break
			}
		}
	}

	threshold := len(queryWords) / 2
	return matches < threshold
}

func (p *ContextPruner) similarWords(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)

	if a == b {
		return true
	}

	return strings.HasPrefix(a, b) || strings.HasPrefix(b, a)
}

func (p *ContextPruner) countTokens(text string) int {
	words := strings.Fields(text)
	return len(words)
}
