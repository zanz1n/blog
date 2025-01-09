-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE users (
    id bigint PRIMARY KEY,
    created_at bigint NOT NULL,
    updated_at bigint NOT NULL,
    permission integer NOT NULL,
    email varchar(128) NOT NULL,
    nickname varchar(32) NOT NULL,
    name text,
    password bytea NOT NULL
);

CREATE UNIQUE INDEX users_email_idx ON users(email);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE IF EXISTS users;
-- +goose StatementEnd
