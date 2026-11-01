package terms

import "testing"

func TestTermMatchesSpellingsAndWordForms(t *testing.T) {
	for _, group := range [][]string{
		{"тёплый", "теплый", "Тёплого", "теплом"},
		{"кабель", "кабеля", "КАБЕЛЕМ"},
		{"шпаклёвка", "шпаклевку"},
		{"cache", "caches", "Caching"},
	} {
		want := Term(group[0])
		for _, w := range group[1:] {
			if got := Term(w); got != want {
				t.Errorf("Term(%q) = %q, want %q like %q", w, got, want, group[0])
			}
		}
	}
	if got := Text("Jaz v2 2026"); got != "jaz v2 2026" {
		t.Errorf("Text = %q, want names and numbers unchanged", got)
	}
}
