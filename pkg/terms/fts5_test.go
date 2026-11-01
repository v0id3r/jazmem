package terms

import (
	"fmt"
	"slices"
	"testing"

	"modernc.org/sqlite"
)

func TestFTS5TokenizerColocatesStems(t *testing.T) {
	tokens := func(flags int) []string {
		var out []string
		if err := (FTS5Tokenizer{}).Tokenize("Тёплого пола, v2", flags, func(token string, tflags, start, end int) error {
			out = append(out, fmt.Sprintf("%s/%d/%d-%d", token, tflags, start, end))
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		return out
	}
	if got, want := tokens(sqlite.FTS5TokenizeDocument), []string{"теплого/0/0-14", "тепл/1/0-14", "пола/0/15-23", "пол/1/15-23", "v2/0/25-27"}; !slices.Equal(got, want) {
		t.Errorf("document tokens = %q, want %q", got, want)
	}
	if got, want := tokens(sqlite.FTS5TokenizeQuery|sqlite.FTS5TokenizePrefix), []string{"теплого/0/0-14", "пола/0/15-23", "v2/0/25-27"}; !slices.Equal(got, want) {
		t.Errorf("prefix tokens = %q, want %q", got, want)
	}
}
