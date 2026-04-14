-- +migrate Up
INSERT INTO accounts (id, name) VALUES (2, 'bob');

-- +migrate Down
ALTER TABLE accounts DELETE WHERE id = 2 SETTINGS mutations_sync = 2;
