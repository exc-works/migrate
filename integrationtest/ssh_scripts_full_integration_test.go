//go:build integration && full

package integrationtest

import "github.com/exc-works/migrate/internal/source"

func sshScriptMigrations(table string) []source.Migration {
	return []source.Migration{
		{Filename: "V2__ssh.ts", Source: "export function up(db) { db.exec('INSERT INTO " + table + " (id) VALUES (1)'); } export function down(db) { db.exec('DELETE FROM " + table + " WHERE id = 1'); }"},
		{Filename: "V3__ssh.js", Source: "export function up(db) { db.exec('INSERT INTO " + table + " (id) VALUES (2)'); } export function down(db) { db.exec('DELETE FROM " + table + " WHERE id = 2'); }"},
	}
}
