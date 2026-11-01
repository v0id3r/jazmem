package jazmem

import "context"

// embedBatch bounds one scheduler tick: bge-m3 on CPU takes 2–10 s per chunk.
const embedBatch = 8

// embedChunks embeds chunk bodies that lack a vector and drops stale vectors.
// It runs outside maintenanceMu because it touches only chunk_embeddings, and
// sends one chunk per request so query embeddings never queue behind a batch.
func (m *Memory) embedChunks(ctx context.Context) error {
	pending, err := m.store.PendingEmbeddings(ctx, m.embedder.Model, embedBatch)
	if err != nil {
		return err
	}
	vectors := make(map[string][]float32, len(pending))
	for _, chunk := range pending {
		embedded, err := m.embedder.Embed(ctx, []string{chunk.Body})
		if err != nil {
			return err
		}
		vectors[chunk.BodyHash] = embedded[0]
	}
	return m.store.SaveEmbeddings(ctx, m.embedder.Model, vectors)
}
