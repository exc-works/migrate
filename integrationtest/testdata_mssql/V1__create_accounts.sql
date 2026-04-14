-- +migrate Up
CREATE TABLE accounts (
    id BIGINT PRIMARY KEY,
    name NVARCHAR(64) NOT NULL
);

-- +migrate Down
DROP TABLE accounts;
