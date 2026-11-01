package sqlite

import (
	"github.com/gluonfield/jazmem/pkg/terms"
	"modernc.org/sqlite"
)

func init() {
	if err := sqlite.RegisterFTS5Tokenizer("jazmem", func([]string) (sqlite.FTS5Tokenizer, error) {
		return terms.FTS5Tokenizer{}, nil
	}); err != nil {
		panic(err)
	}
}
