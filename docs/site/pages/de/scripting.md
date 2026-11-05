# Skripte schreiben

Daten mit JS, TS oder MJS transformieren; Node.js ist zur Ausführung nicht erforderlich.

```sh
migrate new version "normalize names" --format ts
```

`V<version>__<description>.<format>` teilt die SQL-Versionsfolge. MJS unterliegt denselben Grenzen wie JS.

## Einstiegspunkte

Beispiel für SQLite mit `users` und Textspalte `name`:

```ts
export function up(db) {
  for (const row of db.query("SELECT id, name FROM users")) {
    db.exec("UPDATE users SET name=? WHERE id=?", [row.name.trim(), row.id]);
  }
}
```

Das erzeugte up wirft bis zur Implementierung einen Fehler. Kein async, Generator oder Promise; gewöhnliche synchrone Rückgaben werden ignoriert. [SQL-Platzhalter](./script-api) sind datenbankabhängig.

Das Entfernen von Leerzeichen verliert Information; deshalb hat das Beispiel kein down. `down(db)` erst exportieren, wenn die Umkehr implementiert ist. Für Wiederherstellung Originalwerte sichern.

## Vorschau

```sh
migrate up --dry-run
migrate up
migrate status
```

Dry-run führt Skripte nicht aus. DB-Aufrufe gehören in up/down, nicht auf Modulebene. Keinen globalen Zustand zwischen Migrationen voraussetzen.

Vorlagen enthalten Typen; alternativ [migration-api.d.ts](../assets/migration-api.d.ts) verwenden. Die Übersetzung prüft keine Typen. Node-APIs, Netzwerk, Dateisystem, Umgebungsvariablen, Timer und dynamische Importe stehen nicht bereit. Auch npm-Pakete darauf prüfen. Siehe [Module](./modules) und [Transaktionen](./script-transactions).
