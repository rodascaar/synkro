package embeddings

type Tokenizer struct {
	vocabulary   map[string]int
	reverseVocab map[int]string
	maxSeqLength int
	unkToken     string
	padToken     string
}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{
		vocabulary:   make(map[string]int),
		reverseVocab: make(map[int]string),
		maxSeqLength: 512,
		unkToken:     "[UNK]",
		padToken:     "[PAD]",
	}
}

func (t *Tokenizer) SetVocabulary(vocab map[string]int) {
	t.vocabulary = vocab
	t.reverseVocab = make(map[int]string)
	for token, idx := range vocab {
		t.reverseVocab[idx] = token
	}
}

func (t *Tokenizer) GetVocabulary() map[string]int {
	return t.vocabulary
}

func (t *Tokenizer) GetTokenIndex(token string) (int, bool) {
	idx, exists := t.vocabulary[token]
	return idx, exists
}

func (t *Tokenizer) GetToken(idx int) (string, bool) {
	token, exists := t.reverseVocab[idx]
	return token, exists
}

func (t *Tokenizer) GetVocabSize() int {
	return len(t.vocabulary)
}

func (t *Tokenizer) Encode(text string) []int {
	encoded := []int{}
	words := []rune(text)

	for _, word := range words {
		token := string(word)
		if idx, exists := t.vocabulary[token]; exists {
			encoded = append(encoded, idx)
		} else {
			if idx, exists := t.vocabulary[t.unkToken]; exists {
				encoded = append(encoded, idx)
			}
		}
	}

	return encoded
}
