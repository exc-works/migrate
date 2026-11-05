---
layout: home
hero:
  name: migrate
  text: "Gérez les changements de vos bases de données."
  tagline: "Des migrations ordonnées, pour une base ou plusieurs."
  actions:
    - theme: brand
      text: "Commencer"
      link: ./quick-start
    - theme: alt
      text: "Lire la documentation"
      link: ./guide
features:
  - title: "De nombreuses bases, un seul outil"
    details: "PostgreSQL, MySQL, MariaDB, SQLite, Oracle, SQL Server, ClickHouse, TiDB et Redshift."
    link: "./databases"
    linkText: "Voir les bases compatibles"
  - title: "Migrations SQL et scripts"
    details: "Écrivez vos migrations en SQL, TS, JS ou MJS."
    link: "./scripting"
    linkText: "Écrire un script"
  - title: "Mettre à jour plusieurs bases"
    details: "Partagez les fichiers, exécutez les migrations en parallèle et suivez chaque historique séparément."
    link: "./configuration#multiple-databases"
    linkText: "Configurer plusieurs bases"
  - title: "Variables d’environnement"
    details: "Conservez la configuration en YAML et choisissez les connexions selon votre environnement."
    link: "./configuration#environment-variables"
    linkText: "Configurer les connexions"
  - title: "Tunnels SSH et proxys"
    details: "Utilisez un bastion avec mot de passe, clé privée ou SSH Agent, et un proxy SOCKS5 ou HTTP CONNECT facultatif."
    link: "./ssh-tunnels"
    linkText: "Connexion distante"
  - title: "Prévisualiser, vérifier et annuler"
    details: "Vérifiez le plan et les fichiers déjà appliqués, puis revenez à une version avec Down."
    link: "./applying"
    linkText: "Exécuter les changements"
---
