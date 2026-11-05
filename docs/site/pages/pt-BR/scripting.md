# Escrever scripts

Transforme dados com JS, TS ou MJS, sem Node.js na execução.

```sh
migrate new version "normalize names" --format ts
```

`V<version>__<description>.<format>` compartilha versões com SQL. MJS tem os mesmos limites de JS.

```ts
export function up(db) {
  for (const row of db.query("SELECT id, name FROM users")) {
    db.exec("UPDATE users SET name=? WHERE id=?", [row.name.trim(), row.id]);
  }
}
```

O exemplo usa `users.name` em SQLite. O up gerado lança erro até ser implementado. Funções devem ser síncronas: sem async, geradores ou Promise; retornos síncronos comuns são ignorados. [Parâmetros SQL](./script-api) variam por banco.

Remover espaços perde informação, portanto o exemplo não tem down. Exporte `down(db)` apenas com a operação inversa implementada e preserve valores originais quando necessário.

```sh
migrate up --dry-run
migrate up
migrate status
```

Dry-run não executa scripts. Operações DB ficam em up/down, não no topo do módulo. Não dependa de estado global entre migrações. Há tipos nos modelos e em [migration-api.d.ts](../assets/migration-api.d.ts), mas a transpilação não verifica tipos. Não há API Node, rede, arquivos, ambiente, temporizadores ou carregamento dinâmico; confira também npm. Veja [módulos](./modules) e [transações](./script-transactions).
