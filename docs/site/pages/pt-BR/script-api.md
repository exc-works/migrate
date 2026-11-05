# API de banco de dados

`db` fornece `dialect`, `query(sql, params?)`, `exec(sql, params?)` síncronos. Baixe os [tipos](../assets/migration-api.d.ts).

```js
export function up(db) {
  const rows = db.query("SELECT id, name FROM users WHERE name=?", ["Alice"]);
  for (const row of rows) {
    const result = db.exec("UPDATE users SET name=? WHERE id=?", ["alice", row.id]);
    // result.rowsAffected é uma string decimal, como "1".
  }
}
```

`query` carrega o primeiro resultado completo em memória; dê aliases a colunas repetidas. `exec` retorna `{ rowsAffected: string }` sem alterar o valor do driver, inclusive `"0"`.

Parâmetros são arrays: PostgreSQL/Redshift `$1`, MySQL/MariaDB/SQLite/TiDB/ClickHouse `?`, Oracle `:1`, MSSQL `@p1`. `db.dialect` é canônico.

Inteiros e decimais exatos são strings, NULL vira `null`, binários são arrays copiados e horários strings UTC precisas. number e `JSON.parse` podem perder precisão. Decimais SQLite armazenados como float são rejeitados. Veja [tipos específicos](./databases#value-conversion).

SQL segue sem divisão nem interpretação de `GO`; mantenha o ponto e vírgula de MERGE. Lotes MySQL podem exigir opção DSN. Erros de driver são filtrados e capturáveis, mas capturar não repara a transação: use savepoints/rollback ou relance. Retornar normalmente aceita o resultado. Cancelamento, timeout global, valores inválidos e limites continuam fatais mesmo com catch, bloqueando novas chamadas DB. Timeout apenas do driver pode ser capturado. Veja [transações](./script-transactions).
