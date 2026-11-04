//go:build integration && full

package integrationtest

import "github.com/exc-works/migrate/internal/source"

func sshScriptMigrations(table string, flavors ...string) []source.Migration {
	down := func(id string) string {
		if len(flavors) > 0 && flavors[0] == "clickhouse" {
			return "ALTER TABLE " + table + " DELETE WHERE id = " + id + " SETTINGS mutations_sync=1"
		}
		return "DELETE FROM " + table + " WHERE id = " + id
	}
	return []source.Migration{
		{Filename: "V2__ssh.ts", Source: "export function up(db) { db.exec('INSERT INTO " + table + " (id) VALUES (1)'); } export function down(db) { db.exec('" + down("1") + "'); }"},
		{Filename: "V3__ssh.js", Source: "export function up(db) { db.exec('INSERT INTO " + table + " (id) VALUES (2)'); } export function down(db) { db.exec('" + down("2") + "'); }"},
	}
}
