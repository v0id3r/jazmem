package search

import (
	"testing"

	sqlitestore "github.com/gluonfield/jazmem/internal/store/sqlite"
)

func TestFuseRanksAgreementAndMatchStrengthOverSingleFirstPlaces(t *testing.T) {
	row := func(slug string) sqlitestore.SearchResult { return sqlitestore.SearchResult{Slug: slug, Title: slug} }
	rows := fuse([]arm{
		{weight: titleWeight, rows: []sqlitestore.SearchResult{row("companies/mts")}, strength: func(sqlitestore.SearchResult) float64 { return 0.25 }},
		{weight: textWeight, rows: []sqlitestore.SearchResult{row("daily/2026-10-01"), row("projects/laptop")}},
		{via: "semantic", weight: semanticWeight, rows: []sqlitestore.SearchResult{row("projects/laptop")}},
	})
	var order []string
	for _, r := range rows {
		order = append(order, r.Slug)
	}
	if len(order) != 3 || order[0] != "projects/laptop" || order[1] != "daily/2026-10-01" || order[2] != "companies/mts" {
		t.Fatalf("fuse order = %v, want laptop (two arms), daily (first text place, lane prior), mts (weak title)", order)
	}
	if rows[0].Via != "" || rows[0].Score <= rows[1].Score {
		t.Fatalf("laptop = %+v, want a direct hit scoring above the rest", rows[0])
	}
}
