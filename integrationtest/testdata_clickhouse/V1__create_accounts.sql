-- +migrate Up
CREATE TABLE accounts (
    id UInt64,
    name String
) ENGINE = MergeTree()
ORDER BY id;

-- +migrate Down
DROP TABLE accounts;
