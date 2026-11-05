# Datenbank-API

`db` bietet synchron `dialect`, `query(sql, params?)` und `exec(sql, params?)`. [Typen herunterladen](../assets/migration-api.d.ts).

```js
export function up(db) {
  const rows = db.query("SELECT id, name FROM users WHERE name=?", ["Alice"]);
  for (const row of rows) {
    const result = db.exec("UPDATE users SET name=? WHERE id=?", ["alice", row.id]);
    // result.rowsAffected ist eine Dezimalzeichenfolge, beispielsweise "1".
  }
}
```

`query` lädt den ersten vollständigen Ergebnissatz in den Speicher; doppelte Spaltennamen mit Alias versehen. `exec` liefert `{ rowsAffected: string }` und bewahrt den Treiberwert, auch `"0"`.

## Parameter und Werte

Parameter als Array: PostgreSQL/Redshift `$1`, MySQL/MariaDB/SQLite/TiDB/ClickHouse `?`, Oracle `:1`, MSSQL `@p1`. `db.dialect` ist stets kanonisch.

Ganzzahlen und exakte Dezimalwerte sind Zeichenfolgen, NULL wird `null`, Binärwerte kopierte Bytearrays und Zeitstempel präzise UTC-Zeichenfolgen. JavaScript number und `JSON.parse` können Genauigkeit verlieren. SQLite-Dezimalwerte als Fließkomma werden abgelehnt. Siehe [spezielle Typen](./databases#value-conversion).

## SQL und Fehler

SQL wird unverändert gesendet, ohne Batchaufteilung oder `GO`-Auswertung. Terminatoren wie das MERGE-Semikolon erhalten. MySQL-Batches benötigen eventuell DSN-Optionen.

Bereinigte Treiberfehler sind abfangbar, reparieren dadurch aber keine Transaktion. Savepoints oder Rollback nutzen und nicht reparierbare Fehler erneut werfen. Normale Rückkehr akzeptiert das Ergebnis. Abbruch, Gesamttimeout, ungültige Werte und Ressourcenlimits bleiben auch nach catch fatal und sperren weitere DB-Aufrufe. Ein nur vom Treiber gemeldeter Timeout ist abfangbar. Siehe [Transaktionen](./script-transactions).
