package jazmem

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestVectorSearchFindsPagesWithoutSharedTermsAndDegradesWhenEmbedderIsDown(t *testing.T) {
	ollama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Input []string }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
		}
		vectors := make([][]float32, len(req.Input))
		for i, input := range req.Input {
			vectors[i] = []float32{0, 1}
			if strings.Contains(input, "putty") || strings.Contains(input, "шпатлевка") {
				vectors[i] = []float32{1, 0}
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": vectors})
	}))
	mem, err := Open(Config{Root: t.TempDir(), DBPath: filepath.Join(t.TempDir(), "index.sqlite"), EmbeddingEndpoint: ollama.URL})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mem.Close() }()
	for slug, body := range map[string]string{
		"projects/walls":  "Two coats of Knauf putty on the hall walls.",
		"projects/garden": "Apple trees along the fence.",
	} {
		if err := mem.fs.WritePage(slug, "# "+slug+"\n\n"+body+"\n"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := mem.Reindex(t.Context(), ReindexOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := mem.embedChunks(t.Context()); err != nil {
		t.Fatal(err)
	}

	response, err := mem.Retrieve(t.Context(), "шпатлевка", SearchOptions{Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Results) == 0 || response.Results[0].Slug != "projects/walls" || response.Results[0].Via != "semantic" || len(response.Warnings) != 0 {
		t.Fatalf("vector retrieve = %#v, want projects/walls first via semantic", response)
	}

	ollama.Close()
	response, err = mem.Retrieve(t.Context(), "apple", SearchOptions{Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Results) != 1 || response.Results[0].Slug != "projects/garden" || len(response.Warnings) != 1 {
		t.Fatalf("retrieve without embedder = %#v, want text hit plus a warning", response)
	}
}
