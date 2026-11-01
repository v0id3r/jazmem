package sqlite

import (
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gluonfield/jazmem/pkg/terms"
)

func millis(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UTC().UnixMilli()
}

// stopWords are function words that match almost every chunk; queries drop
// them unless nothing else is left.
var stopWords = map[string]bool{
	"a": true, "about": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true,
	"by": true, "did": true, "do": true, "does": true, "for": true, "from": true, "how": true, "i": true,
	"in": true, "into": true, "is": true, "it": true, "me": true, "my": true, "of": true, "on": true,
	"or": true, "tell": true, "that": true, "the": true, "this": true, "to": true, "what": true, "when": true,
	"where": true, "which": true, "who": true, "why": true, "with": true,
	"а": true, "в": true, "во": true, "где": true, "для": true, "до": true, "если": true, "же": true,
	"за": true, "и": true, "из": true, "или": true, "как": true, "какая": true, "какие": true, "каким": true,
	"каких": true, "каком": true, "какой": true, "когда": true, "кто": true, "ли": true, "мне": true, "на": true,
	"не": true, "но": true, "о": true, "об": true, "от": true, "по": true, "про": true, "с": true,
	"со": true, "так": true, "то": true, "у": true, "чем": true, "что": true, "это": true,
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	if limit > 50 {
		return 50
	}
	return limit
}

func chunkPoolLimit(pageLimit int) int {
	limit := normalizeLimit(pageLimit) * 8
	if limit < 50 {
		return 50
	}
	if limit > 500 {
		return 500
	}
	return limit
}

func ftsTokens(query string) []string {
	tokens := terms.Words(query)
	content := tokens[:0:0]
	for _, token := range tokens {
		if !stopWords[terms.Fold(token)] {
			content = append(content, token)
		}
	}
	if len(content) > 0 {
		tokens = content
	}
	if len(tokens) > 8 {
		tokens = tokens[:8]
	}
	return tokens
}

// QueryTerms returns the stems of the words a query searches for.
func QueryTerms(query string) []string {
	tokens := ftsTokens(query)
	for i, token := range tokens {
		tokens[i] = terms.Term(token)
	}
	return tokens
}

func titleAliasLookupTerms(query string) []string {
	full := normalizeLookup(query)
	tokens := ftsTokens(query)
	seen := map[string]bool{}
	var terms []string
	add := func(term string) {
		term = normalizeLookup(term)
		if term == "" || seen[term] {
			return
		}
		seen[term] = true
		terms = append(terms, term)
	}
	add(full)
	for _, token := range tokens {
		term := normalizeLookup(token)
		if stopWords[term] {
			continue
		}
		if utf8.RuneCountInString(term) < 2 && len(tokens) > 1 {
			continue
		}
		add(term)
	}
	if len(terms) > 8 {
		terms = terms[:8]
	}
	return terms
}

func normalizeLookup(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(s))), " ")
}

func normalizeEntity(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Trim(s, " \t\r\n?.!,;:\"'`")
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	return strings.Join(strings.Fields(s), " ")
}

func cleanEntityPhrase(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, " \t\r\n?.!,;:\"'`")
	for _, prefix := range []string{"the ", "a ", "an "} {
		if strings.HasPrefix(strings.ToLower(s), prefix) && len(s) > len(prefix) {
			return strings.TrimSpace(s[len(prefix):])
		}
	}
	return s
}

func cleanSlug(s string) string {
	s = filepath.ToSlash(strings.TrimSpace(s))
	s = strings.TrimSuffix(s, ".md")
	s = strings.Trim(s, " \t\r\n?.!,;:\"'`")
	s = strings.Trim(s, "/")
	if strings.Contains(s, "..") {
		return ""
	}
	return strings.ToLower(s)
}

func ftsQueryAll(tokens []string) string {
	return ftsQueryJoin(tokens, " AND ")
}

func ftsQueryAny(tokens []string) string {
	if len(tokens) <= 1 {
		return ""
	}
	return ftsQueryJoin(tokens, " OR ")
}

func ftsQueryJoin(tokens []string, separator string) string {
	if len(tokens) == 0 {
		return ""
	}
	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		parts = append(parts, `"`+strings.ReplaceAll(token, `"`, `""`)+`"`)
	}
	return strings.Join(parts, separator)
}

func isFTSQuerySyntaxError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "fts5") &&
		(strings.Contains(text, "syntax") || strings.Contains(text, "malformed") || strings.Contains(text, "unterminated"))
}

func uniqueNonEmpty(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
