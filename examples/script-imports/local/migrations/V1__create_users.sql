-- +migrate Up
CREATE TABLE users (id INTEGER PRIMARY KEY, original_name TEXT NOT NULL, display_name TEXT NOT NULL);
INSERT INTO users VALUES (1, ' Alice ', ' Alice ');

-- +migrate Down
DROP TABLE users;
