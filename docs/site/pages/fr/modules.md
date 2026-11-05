# Modules partagés

Importez ou réexportez statiquement `.js`, `.ts`, `.mjs`. Les fichiers utilitaires ne s’exécutent pas comme migrations indépendantes.

## Imports relatifs

```text
project/
  migrations/V2__normalize.ts
  shared/names.ts
```

```ts
import { normalize } from "../shared/names.ts";

export function up(db) {
  db.exec("UPDATE users SET name=? WHERE id=?", [normalize(" Alice "), "1"]);
}
```

Les chemins partent du fichier importeur. Définissez `normalize` dans `shared/names.ts` et créez d’abord `users`. Voir l’[exemple complet](./examples#local).

## Paquets npm

Installez-les à partir du lockfile avant migrate. La résolution prend en charge exports, sous-chemins, dépendances imbriquées et CommonJS statique, avec les conditions `neutral`, puis les entrées `module` et `main`. Pas d’API Node, d’extensions natives ni d’import/require dynamique. migrate n’installe pas les paquets.

## Sources embarquées

`FSSource` résout les utilitaires dans son `fs.FS`, y compris `embed.FS`. Extensions et `index` sont reconnus, mais pas les `package.json` ou `node_modules` virtuels. Pour npm, indiquez le dossier réel :

```go
source := migrate.WithModules{
    Source: migrate.FSSource{
        FS:   embeddedFiles,
        Root: "migrations",
    },
    DependencyDir: projectDirectory,
}
```

Les imports relatifs restent dans le système virtuel. Sans `DependencyDir`, aucun dossier hôte n’est recherché. `StringSource` accepte `WithModules` avec `FS` et `Root`, ou `Directory`. Une source personnalisée peut implémenter `MigrationEntriesSource`. Voir [Go](./go-library).

## Reproductibilité {#checksums}

Seul le fichier d’entrée est contrôlé. Modifier un utilitaire, un paquet ou le lockfile ne déclenche ni divergence ni réexécution. Figez les dépendances et préservez l’ancien comportement. Toutes les dépendances sélectionnées sont compilées avant les écritures ; une erreur tardive empêche aussi les premières étapes.
