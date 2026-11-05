# Exécuter les migrations

Vérifiez la base ciblée et les fichiers à déployer.

## Initialiser l’historique

```sh
migrate create
```

Crée la table si nécessaire, sans exécuter les migrations. Si les changements existent déjà, consultez la [référence initiale](./history#baseline).

## Prévisualiser et appliquer

```sh
migrate up --dry-run
migrate up
migrate status
```

Les versions en attente sont appliquées dans l’ordre et enregistrées après succès. Relancer ne rejoue pas les versions appliquées. Une commande réussie peut rester silencieuse ; vérifiez avec `status`.

Dry-run peut se connecter ou ouvrir SSH pour lire l’historique, mais n’exécute pas les changements. Il vérifie les scripts sans les évaluer et ne prédit donc pas leur SQL dynamique.

## Anciennes versions en attente

Elles sont ignorées par défaut lorsqu’elles sont inférieures à une version appliquée. Activez `migrate_out_of_order` uniquement après vérification de leur compatibilité avec le schéma actuel.

## Déploiement

```sh
migrate --non-interactive up
migrate --non-interactive status --output json
```

Fournissez variables et secrets à l’avance, puis sérialisez les commandes sur une même base. L’historique ne garantit pas une exécution unique face à des appels concurrents, un arrêt brutal ou un commit implicite. Après un échec, inspectez l’état avant de réessayer. Voir [dépannage](./troubleshooting).
