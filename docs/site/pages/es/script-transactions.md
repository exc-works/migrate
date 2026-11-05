# Transacciones y límites

Cuando se admite, cada script y su historial comparten transacción: commit tras el éxito, rollback tras fallo. No emitas BEGIN/COMMIT/ROLLBACK manualmente en ese modo. Los commits DDL implícitos o las bases sin transacciones pueden dejar cambios.

## Desactivar la transacción automática

```js
export const config = { transaction: false };

export function up(db) {
  // PostgreSQL: esta sentencia debe ejecutarse fuera de una transacción.
  db.exec("CREATE INDEX CONCURRENTLY users_name_idx ON users (name)");
}

export function down(db) {
  db.exec("DROP INDEX CONCURRENTLY users_name_idx");
}
```

`transaction` es booleano, predeterminado `true`, para up y down. status, baseline y dry-run no evalúan esa configuración en ejecución.

Una migración comparte conexión dedicada. Finaliza tus transacciones antes de retornar y restaura los ajustes de sesión. No dependas de su persistencia entre migraciones. Un fallo no transaccional puede dejar cambios sin historial exitoso; inspecciona antes de reintentar.

## Limitar recursos

```yaml
script_timeout: 5m # Tiempo máximo de cada script
script_max_rows: 1000 # Máximo de filas por consulta
script_max_bytes: 16777216 # Máximo de bytes por consulta (16 MiB)
```

Superar límites falla sin truncar. Pagina las lecturas grandes; no produce commits por lotes ni reanudación automática. Cancelar solicita detener las operaciones DB, lo que puede tardar. En Go usa `Config.ScriptOptions`. Consulta la [tabla de bases](./databases#script-support).
