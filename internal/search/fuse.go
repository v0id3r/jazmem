package search

import (
	"sort"
	"strings"

	sqlitestore "github.com/gluonfield/jazmem/internal/store/sqlite"
)

// rrfK damps rank differences: first and fifth place differ by about 6%.
const rrfK = 60

// lanePriors rank raw logs below curated pages that say the same thing.
var lanePriors = map[string]float64{"daily": 0.7, "sources": 0.6}

// arm is one retrieval method's rows, best first. Pages any title or text
// arm found carry no via; others take the via of their largest contribution.
// Strength, when set, scales a row's credit by how fully it matches (0..1),
// which rank alone cannot tell.
type arm struct {
	via      string
	weight   float64
	rows     []sqlitestore.SearchResult
	strength func(sqlitestore.SearchResult) float64
}

type chunkKey struct {
	slug  string
	chunk int
}

// fuse ranks pages by weighted reciprocal rank fusion, so arms with different
// score scales combine by rank alone. A page earns weight·(rrfK+1)/(rrfK+rank)
// per arm (first place in a weight-1 arm is 1.0), times its lane prior. Every
// chunk an arm returned is kept and carries its page's score and via.
func fuse(arms []arm) []sqlitestore.SearchResult {
	type page struct {
		score, best float64
		via         string
		direct      bool
	}
	pages := map[string]*page{}
	seen := map[chunkKey]bool{}
	var rows []sqlitestore.SearchResult
	for _, a := range arms {
		rank := 0
		ranked := map[string]bool{}
		for _, row := range a.rows {
			if !ranked[row.Slug] {
				ranked[row.Slug] = true
				rank++
				contribution := a.weight * (rrfK + 1) / float64(rrfK+rank)
				if a.strength != nil {
					contribution *= a.strength(row)
				}
				p := pages[row.Slug]
				if p == nil {
					p = &page{}
					pages[row.Slug] = p
				}
				p.score += contribution
				p.direct = p.direct || a.via == ""
				if contribution > p.best {
					p.best, p.via = contribution, a.via
				}
			}
			if key := (chunkKey{row.Slug, row.ChunkIndex}); !seen[key] {
				seen[key] = true
				rows = append(rows, row)
			}
		}
	}
	for i := range rows {
		p := pages[rows[i].Slug]
		rows[i].Score = p.score * lanePrior(rows[i].Slug)
		if !p.direct {
			rows[i].Via = p.via
		}
	}
	sort.SliceStable(rows, func(a, b int) bool {
		if rows[a].Score != rows[b].Score {
			return rows[a].Score > rows[b].Score
		}
		return rows[a].Slug < rows[b].Slug
	})
	return rows
}

// titleCoverage is the share of query terms found in a title, at least one
// term for alias matches whose title differs.
func titleCoverage(queryTerms []string, title string) float64 {
	if len(queryTerms) == 0 {
		return 1
	}
	inTitle := map[string]bool{}
	for _, term := range sqlitestore.QueryTerms(title) {
		inTitle[term] = true
	}
	found := 0
	for _, term := range queryTerms {
		if inTitle[term] {
			found++
		}
	}
	return float64(max(found, 1)) / float64(len(queryTerms))
}

func lanePrior(slug string) float64 {
	lane, _, _ := strings.Cut(slug, "/")
	if prior, ok := lanePriors[lane]; ok {
		return prior
	}
	return 1
}
