# API de base de datos

`db` expone `dialect`, `query(sql, params?)` y `exec(sql, params?)` síncronos. Descarga los [tipos](../assets/migration-api.d.ts).

```js
export function up(db) {
  const rows = db.query("SELECT id, name FROM users WHERE name=?", ["Alice"]);
  for (const row of rows) {
    const result = db.exec("UPDATE users SET name=? WHERE id=?", ["alice", row.id]);
    // result.rowsAffected es una cadena decimal, por ejemplo "1".
  }
}
```

`query` carga el primer conjunto completo en memoria; usa alias para columnas duplicadas. `exec` devuelve `{ rowsAffected: string }`, manteniendo el conteo del controlador, incluso `"0"`.

## Parámetros y valores

Pasa un array. PostgreSQL/Redshift usan `$1`; MySQL/MariaDB/SQLite/TiDB/ClickHouse `?`; Oracle `:1`; MSSQL `@p1`. `db.dialect` siempre devuelve el nombre canónico.

Enteros y decimales exactos son cadenas, NULL es `null`, los binarios son arrays copiados y las fechas cadenas UTC precisas. Evita pérdidas de precisión con JavaScript number y `JSON.parse`. Se rechazan decimales SQLite guardados como flotantes. Consulta los [tipos particulares](./databases#value-conversion).

## SQL y errores

El SQL se envía sin dividir ni interpretar `GO`. Conserva terminadores, incluido el punto y coma de MERGE. Los lotes MySQL pueden necesitar opciones DSN.

Los errores del controlador se filtran y pueden capturarse, pero eso no repara una transacción. Usa savepoints o rollback y relanza si no es recuperable. Retornar normalmente acepta el resultado. La cancelación, el límite de tiempo global, valores inválidos y límites de recursos siguen siendo fatales aunque se capturen, e impiden nuevas llamadas DB. Un timeout solo del controlador puede capturarse. Consulta [transacciones](./script-transactions).
