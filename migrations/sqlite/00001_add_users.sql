-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE users(
    id integer PRIMARY KEY,
    created_at integer NOT NULL,
    updated_at integer NOT NULL,
    permission integer NOT NULL,
    email text NOT NULL,
    nickname text NOT NULL,
    name text NOT NULL,
    password blob NOT NULL
) STRICT;

CREATE UNIQUE INDEX users_email_idx ON users(email);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE IF EXISTS users;
-- +goose StatementEnd
