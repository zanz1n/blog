-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE keyvalue (
    key text PRIMARY KEY,
    value text NOT NULL,
    expiry bigint
);

CREATE INDEX keyvalue_expiry_idx ON keyvalue(expiry);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE IF EXISTS keyvalue;
-- +goose StatementEnd
