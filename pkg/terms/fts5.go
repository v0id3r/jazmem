package terms

import (
	"unicode"

	"modernc.org/sqlite"
)

// FTS5Tokenizer is a modernc FTS5 tokenizer that emits each word folded and,
// at the same position, its stem, so FTS5 matches word forms in the text
// column and highlight() and snippet() mark the source words. Prefix query
// terms are emitted folded only: a stem is no prefix of what was typed.
// Words are runs of letters and digits.
//
// Indexes rebuild only when their migrations say so, so a change in its
// output needs a migration that rebuilds every index created with it.
type FTS5Tokenizer struct{}

func (FTS5Tokenizer) Tokenize(text string, flags int, emit func(token string, tflags, start, end int) error) error {
	start := -1
	for i, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if start < 0 {
				start = i
			}
			continue
		}
		if start >= 0 {
			if err := emitWord(text, start, i, flags, emit); err != nil {
				return err
			}
			start = -1
		}
	}
	if start >= 0 {
		return emitWord(text, start, len(text), flags, emit)
	}
	return nil
}

func emitWord(text string, start, end, flags int, emit func(string, int, int, int) error) error {
	word := text[start:end]
	folded := Fold(word)
	if err := emit(folded, 0, start, end); err != nil {
		return err
	}
	if stem := Term(word); flags&sqlite.FTS5TokenizePrefix == 0 && stem != folded {
		return emit(stem, sqlite.FTS5TokenColocated, start, end)
	}
	return nil
}
