-- +goose Up
DROP TABLE chunks_fts;
CREATE VIRTUAL TABLE chunks_fts USING fts5(
	slug UNINDEXED,
	chunk_index UNINDEXED,
	title,
	body,
	terms
);

-- +goose Down
DROP TABLE chunks_fts;
CREATE VIRTUAL TABLE chunks_fts USING fts5(
	slug UNINDEXED,
	chunk_index UNINDEXED,
	title,
	body
);
