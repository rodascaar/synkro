package pruner

import (
	"strings"

	"github.com/rodascaar/synkro/internal/memory"
	"github.com/rodascaar/synkro/internal/stopwords"
)

type ContextPruner struct {
	maxTokens int
	stopWords map[string]bool
}

func NewContextPruner() *ContextPruner {
	return &ContextPruner{
		maxTokens: 4000,
		stopWords: stopwords.NewSet(),
	}
}

func (p *ContextPruner) Prune(results []*memory.HybridSearchResult, query string) []*memory.HybridSearchResult {
	if len(results) == 0 {
		return results
	}

	pruned := make([]*memory.HybridSearchResult, 0)
	tokens := 0

	for _, result := range results {
		if result.VectorScore > 0 && result.VectorScore < 0.3 {
			continue
		}

		content := result.Memory.Title + " " + result.Memory.Content
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

func (p *ContextPruner) isLowContent(content, query string) bool {
	words := strings.Fields(strings.ToLower(content))
	queryWords := strings.Fields(strings.ToLower(query))

	var meaningfulQueryWords []string
	for _, qw := range queryWords {
		if !p.stopWords[qw] {
			meaningfulQueryWords = append(meaningfulQueryWords, qw)
		}
	}
	if len(meaningfulQueryWords) == 0 {
		meaningfulQueryWords = queryWords
	}

	matches := 0
	for _, qword := range meaningfulQueryWords {
		for _, word := range words {
			if p.similarWords(word, qword) {
				matches++
				break
			}
		}
	}

	threshold := len(meaningfulQueryWords) / 3
	if threshold < 2 {
		threshold = 2
	}
	return matches < threshold
}

func (p *ContextPruner) similarWords(a, b string) bool {
	a = strings.TrimSpace(strings.ToLower(a))
	b = strings.TrimSpace(strings.ToLower(b))

	if a == b {
		return true
	}

	if len(a) >= 4 && len(b) >= 4 && (strings.Contains(a, b) || strings.Contains(b, a)) {
		return true
	}

	if len(a) >= 4 && len(b) >= 4 && (strings.HasPrefix(a, b) || strings.HasPrefix(b, a)) {
		return true
	}

	return false
}

func (p *ContextPruner) countTokens(text string) int {
	words := strings.Fields(text)
	return len(words)
}
