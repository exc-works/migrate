# API de base de données

`db` fournit `dialect`, `query(sql, params?)` et `exec(sql, params?)`, synchrones. Voir les [types](../assets/migration-api.d.ts).

```js
export function up(db) {
  const rows = db.query("SELECT id, name FROM users WHERE name=?", ["Alice"]);
  for (const row of rows) {
    const result = db.exec("UPDATE users SET name=? WHERE id=?", ["alice", row.id]);
    // result.rowsAffected est une chaîne décimale, par exemple "1".
  }
}
```

`query` charge le premier résultat complet en mémoire ; donnez des alias aux colonnes de même nom. `exec` retourne `{ rowsAffected: string }`, sans changer le compte du pilote, même `"0"`.

## Paramètres et valeurs

Passez un tableau. PostgreSQL/Redshift utilisent `$1`, MySQL/MariaDB/SQLite/TiDB/ClickHouse `?`, Oracle `:1`, MSSQL `@p1`. `db.dialect` retourne toujours le nom canonique.

Entiers et décimaux exacts sont des chaînes, NULL devient `null`, les binaires sont copiés en tableaux d’octets, les dates sont des chaînes UTC précises. Évitez les pertes de précision de JavaScript number et `JSON.parse`. Les décimaux SQLite stockés en flottants sont refusés. Voir les [types propres aux bases](./databases#value-conversion).

## SQL et erreurs

Le SQL est transmis tel quel, sans découpage ni interprétation de `GO`. Conservez les terminaisons comme le point-virgule de MERGE. MySQL peut exiger une option DSN pour les lots.

Les erreurs de pilote sont filtrées et interceptables. Les intercepter ne répare pas la transaction : utilisez points de sauvegarde ou rollback, puis relancez l’erreur si nécessaire. Un retour normal accepte le résultat. Annulation, délai global, valeurs invalides et limites dépassées restent fatals même après interception et interdisent les appels DB suivants. Un délai signalé uniquement par le pilote reste interceptable. Voir [transactions](./script-transactions).
