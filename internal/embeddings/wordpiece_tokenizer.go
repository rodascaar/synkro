package embeddings

import (
	"os"
	"strings"
	"unicode"
)

type WordPieceTokenizer struct {
	vocab     map[string]int
	unkToken  string
	clsToken  string
	sepToken  string
	padToken  string
	unkID     int
	clsID     int
	sepID     int
	padID     int
	maxSeqLen int
}

func NewWordPieceTokenizer(vocabPath string, maxSeqLen int) (*WordPieceTokenizer, error) {
	data, err := readFileLines(vocabPath)
	if err != nil {
		return nil, err
	}

	vocab := make(map[string]int, len(data))
	for idx, line := range data {
		token := strings.TrimSpace(line)
		if token != "" {
			vocab[token] = idx
		}
	}

	t := &WordPieceTokenizer{
		vocab:     vocab,
		unkToken:  "[UNK]",
		clsToken:  "[CLS]",
		sepToken:  "[SEP]",
		padToken:  "[PAD]",
		maxSeqLen: maxSeqLen,
	}

	if id, ok := vocab["[UNK]"]; ok {
		t.unkID = id
	}
	if id, ok := vocab["[CLS]"]; ok {
		t.clsID = id
	}
	if id, ok := vocab["[SEP]"]; ok {
		t.sepID = id
	}
	if id, ok := vocab["[PAD]"]; ok {
		t.padID = id
	}

	if t.maxSeqLen <= 0 {
		t.maxSeqLen = 128
	}

	return t, nil
}

func (t *WordPieceTokenizer) Encode(text string) (inputIDs []int64, attentionMask []int64, tokenTypeIDs []int64) {
	text = strings.ToLower(text)
	text = strings.TrimSpace(text)

	text = cleanText(text)

	words := whitespaceTokenize(text)

	var tokens []string
	tokens = append(tokens, t.clsToken)

	for _, word := range words {
		subTokens := t.wordPiece(word)
		tokens = append(tokens, subTokens...)
	}

	tokens = append(tokens, t.sepToken)

	maxTokens := t.maxSeqLen - 1
	if len(tokens) > maxTokens {
		tokens = tokens[:maxTokens-1]
		tokens = append(tokens, t.sepToken)
	}

	inputIDs = make([]int64, len(tokens))
	attentionMask = make([]int64, len(tokens))
	tokenTypeIDs = make([]int64, len(tokens))

	for i, token := range tokens {
		if id, ok := t.vocab[token]; ok {
			inputIDs[i] = int64(id)
		} else {
			inputIDs[i] = int64(t.unkID)
		}
		attentionMask[i] = 1
		tokenTypeIDs[i] = 0
	}

	return inputIDs, attentionMask, tokenTypeIDs
}

func (t *WordPieceTokenizer) wordPiece(word string) []string {
	if len(word) > 200 {
		return []string{t.unkToken}
	}

	tokens := []string{}
	start := 0

	for start < len(word) {
		end := len(word)
		curSubstr := ""
		found := false

		for start < end {
			substr := word[start:end]
			if start > 0 {
				substr = "##" + substr
			}

			if _, ok := t.vocab[substr]; ok {
				curSubstr = substr
				found = true
				break
			}
			end--
		}

		if !found {
			return []string{t.unkToken}
		}

		tokens = append(tokens, curSubstr)
		start = end
	}

	return tokens
}

func (t *WordPieceTokenizer) MaxSeqLen() int {
	return t.maxSeqLen
}

func (t *WordPieceTokenizer) PadID() int64 {
	return int64(t.padID)
}

func cleanText(text string) string {
	var b strings.Builder
	b.Grow(len(text))

	for _, r := range text {
		if r == 0 || r == 0xFFFD || unicode.IsControl(r) {
			continue
		}
		if unicode.Is(unicode.Zs, r) {
			b.WriteRune(' ')
			continue
		}
		if !unicode.IsPrint(r) && !unicode.IsSpace(r) {
			continue
		}
		b.WriteRune(r)
	}

	return b.String()
}

func whitespaceTokenize(text string) []string {
	return strings.Fields(text)
}

func readFileLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}
