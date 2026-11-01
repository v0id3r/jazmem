package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Ollama embeds text through an Ollama server's /api/embed endpoint.
type Ollama struct {
	Endpoint string
	Model    string
}

func (o *Ollama) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	body, err := json.Marshal(map[string]any{"model": o.Model, "input": inputs})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(o.Endpoint, "/")+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("ollama embed: %s: %s", resp.Status, strings.TrimSpace(string(detail)))
	}
	var out struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("ollama embed: decode: %w", err)
	}
	if len(out.Embeddings) != len(inputs) {
		return nil, fmt.Errorf("ollama embed: got %d vectors for %d inputs", len(out.Embeddings), len(inputs))
	}
	return out.Embeddings, nil
}
