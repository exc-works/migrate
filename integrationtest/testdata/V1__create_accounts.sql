-- +migrate Up
CREATE TABLE accounts (
    id BIGINT PRIMARY KEY,
    name VARCHAR(64) NOT NULL
);

-- +migrate Down
DROP TABLE accounts;
