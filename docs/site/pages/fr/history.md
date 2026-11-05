# Historique et référence initiale

La table conserve versions, noms et empreintes. Comparez-la au schéma réel.

```sh
migrate status
migrate status --output json
```

| État | Signification | Action |
| --- | --- | --- |
| `pending` | Non appliqué | Vérifier puis `up` |
| `applied` | Appliqué avec succès | Ne pas modifier |
| `baseline` | Enregistré sans exécution | Conserver l’historique adopté |
| `outOfOrder` | Version en attente plus ancienne | Vérifier la politique d’ordre |
| `hashMismatch` | Contenu modifié | Restaurer le fichier d’origine |
| `filenameMismatch` | Nom modifié | Restaurer le nom d’origine |

Les divergences bloquent l’exécution. Ajoutez une version pour tout nouveau changement. L’empreinte des scripts ne couvre pas les [modules importés](./modules#checksums).

## Adopter une base existante {#baseline}

Uniquement si son schéma et ses données correspondent déjà aux fichiers :

1. Comparez les fichiers et l’état réel.
2. Exécutez `migrate create` si nécessaire.
3. Lancez `migrate baseline`, puis `migrate status`.
4. Ajoutez ensuite des versions supérieures.

Baseline enregistre sans exécuter. Il ne répare pas une migration partiellement échouée.
