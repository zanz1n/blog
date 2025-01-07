-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE users(
    id integer PRIMARY KEY,
    created_at text NOT NULL,
    updated_at text NOT NULL,
    permission integer NOT NULL,
    email text NOT NULL,
    nickname text NOT NULL,
    name text,
    password blob NOT NULL
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE IF EXISTS users;
-- +goose StatementEnd
