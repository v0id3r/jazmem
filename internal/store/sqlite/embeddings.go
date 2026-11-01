package sqlite

import (
	"context"
	"encoding/binary"
	"math"

	"github.com/gluonfield/jazmem/internal/store/sqlite/generated/indexdb"
	"github.com/gluonfield/jazmem/internal/store/sqlite/generated/searchdb"
)

type PendingEmbedding struct {
	BodyHash string
	Body     string
}

// PendingEmbeddings lists chunk bodies that have no embedding for model yet.
func (s *Store) PendingEmbeddings(ctx context.Context, model string, limit int) ([]PendingEmbedding, error) {
	rows, err := indexdb.New(s.db).PendingChunkEmbeddings(ctx, indexdb.PendingChunkEmbeddingsParams{Model: model, Limit: int64(limit)})
	if err != nil {
		return nil, err
	}
	out := make([]PendingEmbedding, 0, len(rows))
	for _, row := range rows {
		out = append(out, PendingEmbedding{BodyHash: row.BodyHash, Body: row.Body})
	}
	return out, nil
}

// SaveEmbeddings stores vectors by chunk body hash and drops vectors of other
// models or of bodies no longer indexed.
func (s *Store) SaveEmbeddings(ctx context.Context, model string, vectors map[string][]float32) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)
	q := indexdb.New(tx)
	for bodyHash, vector := range vectors {
		if err := q.InsertChunkEmbedding(ctx, indexdb.InsertChunkEmbeddingParams{BodyHash: bodyHash, Model: model, Embedding: vectorBlob(vector)}); err != nil {
			return err
		}
	}
	if err := q.DeleteStaleChunkEmbeddings(ctx, model); err != nil {
		return err
	}
	return tx.Commit()
}

// SearchVector ranks embedded chunks by cosine distance to vector; Score is
// the distance, lower is closer.
func (s *Store) SearchVector(ctx context.Context, model string, vector []float32, limit int) ([]SearchResult, error) {
	rows, err := s.searchQ.SearchVector(ctx, searchdb.SearchVectorParams{QueryVector: vectorBlob(vector), Model: model, Limit: int64(chunkPoolLimit(limit))})
	if err != nil {
		return nil, err
	}
	results := make([]SearchResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, SearchResult{
			Slug:       row.Slug,
			Title:      row.Title,
			ChunkIndex: int(row.ChunkIndex),
			Snippet:    row.Snippet,
			Score:      row.Distance,
		})
	}
	return bestPerPage(results, limit), nil
}

// vectorBlob encodes a vector the way sqlite-vec reads float32 BLOBs.
func vectorBlob(vector []float32) []byte {
	blob := make([]byte, 4*len(vector))
	for i, v := range vector {
		binary.LittleEndian.PutUint32(blob[4*i:], math.Float32bits(v))
	}
	return blob
}
