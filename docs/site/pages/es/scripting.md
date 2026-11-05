# Escribir scripts

Transforma datos con JS, TS o MJS sin Node.js durante la ejecución.

```sh
migrate new version "normalize names" --format ts
```

Los archivos `V<version>__<description>.<format>` comparten versiones con SQL. MJS tiene las mismas restricciones que JS.

## Funciones de entrada

Ejemplo para SQLite con una tabla `users` y columna de texto `name`:

```ts
export function up(db) {
  for (const row of db.query("SELECT id, name FROM users")) {
    db.exec("UPDATE users SET name=? WHERE id=?", [row.name.trim(), row.id]);
  }
}
```

El `up` generado lanza un error hasta implementarlo. No se admiten async, generadores ni Promise; los retornos síncronos ordinarios se ignoran. Consulta los [parámetros SQL](./script-api).

Recortar espacios pierde datos originales, así que el ejemplo no tiene down. Exporta `down(db)` solo cuando hayas implementado la inversa. Guarda los valores originales si necesitas recuperarlos.

## Previsualizar

```sh
migrate up --dry-run
migrate up
migrate status
```

Dry-run no ejecuta scripts. Coloca operaciones DB dentro de up/down, no al nivel del módulo. No dependas de variables globales entre migraciones.

Los modelos incluyen tipos; puedes usar [migration-api.d.ts](../assets/migration-api.d.ts). La transpilación no comprueba tipos. No existen API Node, red, archivos, entorno, temporizadores ni carga dinámica. Verifica también los paquetes npm. Consulta [módulos](./modules) y [transacciones](./script-transactions).
