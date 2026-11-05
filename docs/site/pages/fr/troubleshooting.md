# Dépannage

Commencez avec `migrate version`, `migrate capabilities --output json`, `migrate status --output json`. Seul status exige une connexion.

## Commande ou configuration absente

Vérifiez PATH selon l’[installation](./installation). Générez la configuration dans le projet ou utilisez `-c` / `-w`. Pour YAML, suivez la position indiquée, vérifiez valeurs et indentation avec des espaces, pas des tabulations. Consultez l’[ordre de recherche](./configuration#configuration-file-search-order).

## Variable manquante

`can't find env value for XXX` exige de définir `${XXX}` ; un défaut `${XXX:default}` convient seulement aux valeurs appropriées non secrètes. Les champs SSH `password_env` contiennent le nom de variable.

## Format, base ou module non pris en charge

Vérifiez `capabilities` et mettez à jour les anciennes versions. `unsupported dialect` : contrôlez les [alias](./databases#aliases). Vérifiez les imports relatifs et installez npm avant migrate. Une source embarquée doit préciser `DependencyDir` pour npm. Les dépendances Node ou asynchrones restent incompatibles.

## Historique différent ou aucun changement après up

Restaurez fichiers et noms appliqués depuis Git ; ajoutez ensuite une version supérieure. Ne contournez pas l’erreur par baseline sur un état inconnu. Une commande silencieuse peut réussir : vérifiez applied, baseline ou outOfOrder avec status.

## Annulation ou connexion échouée

`to-version must be set, or use --all` exige la cible ou `--all`. La cible reste appliquée. Un script sans down ne s’annule pas.

Inspectez données et historique avant toute relance : DDL implicite, opérations non transactionnelles et coupures autour du commit peuvent laisser des changements. Baseline ne répare pas cela. Vérifiez les secrets SSH, known_hosts, proxy et accès du bastion à la base. Pas de reconnexion ni de rejeu automatique. Consultez les journaux serveur contrôlés sans partager d’identifiants ni de données.
