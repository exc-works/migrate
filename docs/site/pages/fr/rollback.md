# Annuler les migrations

L’annulation exécute Down ou `down(db)`. Elle ne restaure que ce que ce code prévoit, sans récupération automatique des données supprimées.

## Conserver une version cible

```sh
migrate down 202611030001 --dry-run
migrate down 202611030001
migrate status
```

Seules les versions supérieures à la cible sont annulées. Avec 1, 2 et 3 appliquées, `down 2` annule 3 et conserve 1 et 2.

## Tout annuler

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

La cible et `--all` sont incompatibles. Les versions annulées redeviennent `pending` et peuvent être réappliquées.

Vérifiez tout le périmètre avant l’exécution. Un script sans `down(db)` bloque l’ensemble sélectionné ; une fonction vide n’annule rien. DDL et opérations non transactionnelles peuvent subsister après un échec. Vérifiez données et historique, puis les [transactions](./script-transactions).
