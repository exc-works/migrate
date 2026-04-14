-- +migrate Up
INSERT INTO accounts (id, name) VALUES (1, 'alice');

-- +migrate Down
ALTER TABLE accounts DELETE WHERE id = 1 SETTINGS mutations_sync = 2;
