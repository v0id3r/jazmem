-- +goose Up
-- The jazmem tokenizer stores each word's stem at the word's position, so the
-- terms column goes; the indexer version bump refills the table.
DROP TABLE chunks_fts;
CREATE VIRTUAL TABLE chunks_fts USING fts5(
	slug UNINDEXED,
	chunk_index UNINDEXED,
	title,
	body,
	tokenize = 'jazmem'
);

-- +goose Down
DROP TABLE chunks_fts;
CREATE VIRTUAL TABLE chunks_fts USING fts5(
	slug UNINDEXED,
	chunk_index UNINDEXED,
	title,
	body,
	terms
);
