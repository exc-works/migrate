# Transactions et limites

Quand la base le permet, chaque script et son historique partagent une transaction. Succès : commit ; échec : rollback. N’exécutez pas manuellement BEGIN/COMMIT/ROLLBACK dans ce mode. Les commits DDL implicites et les bases sans transactions peuvent laisser des changements.

## Désactiver la transaction automatique

```js
export const config = { transaction: false };

export function up(db) {
  // PostgreSQL : cette instruction exige une exécution hors transaction.
  db.exec("CREATE INDEX CONCURRENTLY users_name_idx ON users (name)");
}

export function down(db) {
  db.exec("DROP INDEX CONCURRENTLY users_name_idx");
}
```

`transaction` est booléen, vaut `true` par défaut et s’applique à up et down. status, baseline et dry-run n’évaluent pas la configuration exportée à l’exécution.

Les appels d’une migration partagent une connexion dédiée. Terminez vos propres transactions avant le retour et restaurez les réglages de session. Ne comptez pas sur leur persistance entre migrations. En cas d’échec non transactionnel, des changements peuvent rester sans historique réussi ; inspectez avant de relancer.

## Limiter les ressources

```yaml
script_timeout: 5m # Durée maximale de chaque script
script_max_rows: 1000 # Nombre maximal de lignes par requête
script_max_bytes: 16777216 # Octets maximaux par requête (16 MiB)
```

Les dépassements échouent sans tronquer. Paginez les lectures ; cela ne crée pas de commits par lots ni de reprise automatique. L’annulation demande l’arrêt des opérations DB, qui peut prendre du temps. En Go, utilisez `Config.ScriptOptions`. Voir les [transactions par base](./databases#script-support).
