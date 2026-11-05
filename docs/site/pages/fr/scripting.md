# Écrire des scripts

Utilisez JS, TS ou MJS pour les transformations de données, sans Node.js à l’exécution.

```sh
migrate new version "normalize names" --format ts
```

Les fichiers `V<version>__<description>.<format>` partagent les versions SQL. MJS a les mêmes restrictions que JS.

## Fonctions d’entrée

Exemple sur une table SQLite `users` avec une colonne texte `name` :

```ts
export function up(db) {
  for (const row of db.query("SELECT id, name FROM users")) {
    db.exec("UPDATE users SET name=? WHERE id=?", [row.name.trim(), row.id]);
  }
}
```

Le modèle `up` généré lève une erreur tant qu’il n’est pas implémenté. Les fonctions doivent être synchrones : pas d’async, de générateur ni de Promise. Les retours synchrones ordinaires sont ignorés. Voir les [paramètres SQL](./script-api).

Supprimer les espaces perd de l’information : cet exemple n’a pas de `down`. N’exportez `down(db)` qu’après avoir écrit l’inverse ; conservez les valeurs initiales si vous devez les restaurer.

## Prévisualiser

```sh
migrate up --dry-run
migrate up
migrate status
```

Dry-run n’exécute pas les scripts. Placez les opérations DB dans up/down, jamais au niveau du module. Ne partagez pas d’état global entre migrations.

Les modèles incluent des types ; [migration-api.d.ts](../assets/migration-api.d.ts) peut les remplacer. La transpilation ne vérifie pas les types. Les API Node, réseau, fichiers, environnement, timers et chargements dynamiques sont absents. Vérifiez aussi les dépendances npm. Voir [modules](./modules) et [transactions](./script-transactions).
