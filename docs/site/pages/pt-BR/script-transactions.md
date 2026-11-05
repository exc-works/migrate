# Transações e limites

Quando suportado, script e histórico compartilham uma transação: commit no sucesso, rollback no erro. Não execute BEGIN/COMMIT/ROLLBACK manualmente. DDL com commit implícito e bancos sem transação podem deixar mudanças.

## Desativar a transação automática

```js
export const config = { transaction: false };

export function up(db) {
  // PostgreSQL: esta instrução deve executar fora de uma transação.
  db.exec("CREATE INDEX CONCURRENTLY users_name_idx ON users (name)");
}

export function down(db) {
  db.exec("DROP INDEX CONCURRENTLY users_name_idx");
}
```

`transaction` é booleano, padrão `true`, para up e down. status, baseline e dry-run não avaliam essa configuração em execução. Uma migração compartilha conexão dedicada. Termine suas transações e restaure ajustes de sessão antes de retornar; não dependa da sessão entre migrações. Uma falha não transacional pode deixar mudanças sem histórico de sucesso.

## Limitar recursos

```yaml
script_timeout: 5m # Tempo máximo por script
script_max_rows: 1000 # Máximo de linhas por consulta
script_max_bytes: 16777216 # Máximo de bytes por consulta (16 MiB)
```

Exceder limites falha sem truncar. Pagine leituras grandes, sem presumir commits por lote ou retomada. Cancelar solicita a interrupção DB, que pode levar tempo. Em Go use `Config.ScriptOptions`. Consulte [cada banco](./databases#script-support).
