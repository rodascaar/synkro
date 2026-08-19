package stopwords

// words is the shared stopword list used by the embedding generator and the
// context pruner. All entries are lowercase; callers must lowercase input
// before looking up.
var words = []string{
	"el", "la", "de", "que", "y", "a", "en", "un", "por",
	"con", "no", "una", "su", "para", "es", "del", "los",
	"the", "a", "an", "and", "are", "as", "at", "be", "by",
	"for", "from", "has", "he", "in", "is", "it", "its", "of",
	"on", "that", "the", "to", "was", "were", "will", "with",
}

// NewSet returns a map of lowercase stopwords.
func NewSet() map[string]bool {
	set := make(map[string]bool, len(words))
	for _, w := range words {
		set[w] = true
	}
	return set
}