---
layout: home
hero:
  name: migrate
  text: Manage database changes.
  tagline: Keep database changes in order, across one database or many.
  actions:
    - theme: brand
      text: Get started
      link: ./quick-start
    - theme: alt
      text: Explore the documentation
      link: ./guide
features:
  - title: Many databases, one tool
    details: Use PostgreSQL, MySQL, MariaDB, SQLite, Oracle, SQL Server, ClickHouse, TiDB and Redshift.
    link: ./databases
    linkText: Explore database support
  - title: SQL and script migrations
    details: Write database migrations in SQL, TS, JS or MJS.
    link: ./scripting
    linkText: Write migration scripts
  - title: Update multiple databases
    details: Share one migration directory across databases, run concurrently and track history and results for each database.
    link: ./configuration#multiple-databases
    linkText: Configure multiple databases
  - title: Environment variables
    details: Switch database connections and deployment environments through environment variables, with project settings in YAML.
    link: ./configuration#environment-variables
    linkText: Configure your environment
  - title: SSH tunnels and proxies
    details: Reach databases through an SSH bastion using passwords, private keys or SSH Agent, with optional SOCKS5 or HTTP CONNECT proxies.
    link: ./ssh-tunnels
    linkText: Connect to remote databases
  - title: Preview, validate and roll back
    details: Preview changes before execution, detect edits to applied files and roll back versions using their Down operations.
    link: ./applying
    linkText: Apply database changes
---
