package terms

import "testing"

func TestTermMatchesSpellingsAndWordForms(t *testing.T) {
	for _, group := range [][]string{
		{"тёплый", "теплый", "Тёплого", "теплом"},
		{"кабель", "кабеля", "КАБЕЛЕМ"},
		{"шпаклёвка", "шпаклевку"},
	} {
		want := Term(group[0])
		for _, w := range group[1:] {
			if got := Term(w); got != want {
				t.Errorf("Term(%q) = %q, want %q like %q", w, got, want, group[0])
			}
		}
	}
	if got := Text("Jaz MCP_server v2"); got != "jaz mcp_server v2" {
		t.Errorf("Text(latin) = %q, want lowercase words unchanged", got)
	}
}
