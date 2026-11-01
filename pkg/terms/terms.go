// Package terms gives words the form full-text indexes store and match, so ё/е
// spellings and Russian word forms find each other.
package terms

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/kljensen/snowball/russian"
)

var word = regexp.MustCompile(`[\p{L}\p{M}\p{N}_]+`)

var yoFold = strings.NewReplacer("ё", "е")

// Words splits text into words.
func Words(text string) []string {
	return word.FindAllString(text, -1)
}

// Fold lowercases a word and folds ё to е.
func Fold(w string) string {
	return yoFold.Replace(strings.ToLower(w))
}

// Term returns the search form of a word: folded, and reduced to its Snowball
// stem when Cyrillic. The stemmer must see the folded word: it stems ё and е
// spellings differently.
func Term(w string) string {
	w = Fold(w)
	if !strings.ContainsFunc(w, isCyrillic) {
		return w
	}
	return russian.Stem(w, false)
}

// Text returns the search forms of the words in text, space-separated.
func Text(text string) string {
	words := Words(text)
	for i, w := range words {
		words[i] = Term(w)
	}
	return strings.Join(words, " ")
}

func isCyrillic(r rune) bool {
	return unicode.Is(unicode.Cyrillic, r)
}
