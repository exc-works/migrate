---
layout: home
hero:
  name: migrate
  text: Database migrations, kept in order.
  tagline: A Go CLI and library for SQL, with an optional embedded JavaScript and TypeScript runtime.
  actions:
    - theme: brand
      text: Start with SQLite
      link: ./quick-start
    - theme: alt
      text: Read the guide
      link: ./guide
features:
  - title: SQL first
    details: Versioned migrations, dry runs, history validation and explicit rollback.
  - title: Scripts when needed
    details: The full build adds JS, TS and MJS with shared local and installed npm modules. No Node.js runtime is needed to execute migrations.
  - title: Know the boundaries
    details: Inspect verification status, aliases and transaction limits with migrate capabilities before choosing a database.
---

## Choose your next step

- [Quick start](./quick-start): run a reversible SQLite migration without a database server.
- [User guide](./guide): installation, configuration, migration history and Go integration.
- [Commands](./commands): find the right command and its reference section.
- [Database capabilities](./databases): SQL/full builds, verification and transaction limits.
- [Script migrations](./scripting) and [runnable examples](./examples).
- [Release notes](./releases), [changelog](./changelog) and [other languages](./languages).

::: info Version availability
This site follows the repository's main branch. Unreleased changes may appear here;
check the release notes for your installed version with `migrate version`.
:::
