# Transaktionen und Grenzen

Wo unterstützt, teilen Skript und Erfolgsverlauf eine Transaktion: Commit bei Erfolg, Rollback bei Fehler. BEGIN/COMMIT/ROLLBACK nicht selbst ausführen. Implizite DDL-Commits und nichttransaktionale Datenbanken können Änderungen hinterlassen.

## Automatische Transaktion deaktivieren

```js
export const config = { transaction: false };

export function up(db) {
  // PostgreSQL: Diese Anweisung muss außerhalb einer Transaktion laufen.
  db.exec("CREATE INDEX CONCURRENTLY users_name_idx ON users (name)");
}

export function down(db) {
  db.exec("DROP INDEX CONCURRENTLY users_name_idx");
}
```

`transaction` ist boolesch, standardmäßig `true`, für up und down. status, baseline und dry-run werten die exportierte Laufzeitkonfiguration nicht aus.

Alle Aufrufe einer Migration nutzen eine eigene gemeinsame Verbindung. Eigene Transaktionen vor Rückkehr beenden und Sitzungseinstellungen zurücksetzen. Nicht auf fortbestehende Sitzungen zwischen Migrationen verlassen. Ohne Transaktion können Änderungen ohne Erfolgsverlauf bleiben; vor Wiederholung prüfen.

## Ressourcen begrenzen

```yaml
script_timeout: 5m # Zeitlimit je Skript
script_max_rows: 1000 # Maximale Zeilen je Abfrage
script_max_bytes: 16777216 # Maximale Bytes je Abfrage (16 MiB)
```

Überschreitungen führen zum Fehler statt Abschneiden. Große Abfragen paginieren; das erzeugt keine Teilcommits oder automatische Wiederaufnahme. Ein Abbruch fordert das Stoppen der DB an, was Zeit benötigen kann. Go verwendet `Config.ScriptOptions`. Siehe [Datenbanktabelle](./databases#script-support).
