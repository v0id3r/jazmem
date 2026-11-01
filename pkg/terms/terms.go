// Package terms gives words the form full-text indexes store and match, so ё/е
// spellings and Russian and English word forms find each other.
package terms

import (
	"strings"
	"unicode"

	"github.com/blevesearch/snowballstem"
	"github.com/blevesearch/snowballstem/english"
	"github.com/blevesearch/snowballstem/russian"
)

var yoFold = strings.NewReplacer("ё", "е")

// Words splits text into words: runs of letters, marks, digits and underscores.
func Words(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsMark(r) && !unicode.IsNumber(r) && r != '_'
	})
}

// Fold lowercases a word and folds ё to е.
func Fold(w string) string {
	return yoFold.Replace(strings.ToLower(w))
}

// Term returns the search form of a word: folded and reduced to its Snowball
// stem, Russian for Cyrillic words and English for the rest.
func Term(w string) string {
	w = Fold(w)
	env := snowballstem.NewEnv(w)
	if strings.ContainsFunc(w, isCyrillic) {
		russian.Stem(env)
	} else {
		english.Stem(env)
	}
	return env.Current()
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
