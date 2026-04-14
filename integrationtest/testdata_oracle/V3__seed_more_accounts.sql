-- +migrate Up
INSERT INTO accounts (id, name) VALUES (2, 'bob');

-- +migrate Down
DELETE FROM accounts WHERE id = 2;
