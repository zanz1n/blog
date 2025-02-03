-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE articles (
    id integer PRIMARY KEY,
    created_at integer NOT NULL,
    updated_at integer NOT NULL,
    user_id integer NOT NULL DEFAULT 0,
    title text NOT NULL,
    description text,
    indexing text NOT NULL,
    content text NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(id)
        ON DELETE SET DEFAULT ON UPDATE CASCADE
);

CREATE INDEX articles_user_id_idx ON articles(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE IF EXISTS articles;
-- +goose StatementEnd
