-- +migrate Up
CREATE TABLE accounts (
    id NUMBER(19) PRIMARY KEY,
    name VARCHAR2(64 CHAR) NOT NULL
);

-- +migrate Down
DROP TABLE accounts;
