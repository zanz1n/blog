-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE articles (
    id bigint PRIMARY KEY,
    created_at bigint NOT NULL,
    updated_at bigint NOT NULL,
    user_id bigint NOT NULL DEFAULT 0,
    title text NOT NULL,
    description text,
    indexing jsonb NOT NULL,
    content text NOT NULL,
    raw_content text NOT NULL
);

ALTER TABLE articles ADD CONSTRAINT articles_user_id_fkey
FOREIGN KEY (user_id) REFERENCES users(id)
ON DELETE SET DEFAULT ON UPDATE CASCADE;

CREATE INDEX articles_user_id_idx ON articles(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE IF EXISTS articles;
-- +goose StatementEnd
