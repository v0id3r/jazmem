-- +goose Up
CREATE TABLE chunk_embeddings (
	body_hash TEXT NOT NULL,
	model TEXT NOT NULL,
	embedding BLOB NOT NULL,
	PRIMARY KEY (body_hash, model)
);

-- +goose Down
DROP TABLE chunk_embeddings;
