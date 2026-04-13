-- +migrate Up
INSERT INTO accounts (id, name) VALUES (1, 'alice');

-- +migrate Down
DELETE FROM accounts WHERE id = 1;
