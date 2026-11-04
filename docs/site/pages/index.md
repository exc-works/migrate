---
layout: home
hero:
  name: migrate
  text: Manage database changes.
  tagline: Version SQL migrations and transform data with JavaScript or TypeScript.
  actions:
    - theme: brand
      text: Get started
      link: ./quick-start
    - theme: alt
      text: Explore the documentation
      link: ./guide
features:
  - title: SQL first
    details: Versioned migrations, dry runs, history validation and explicit rollback.
    link: ./sql-migrations
    linkText: Write a migration
  - title: Scripts when needed
    details: The full build adds JS, TS and MJS with shared local and installed npm modules. No Node.js runtime is needed to execute migrations.
    link: ./scripting
    linkText: Use JavaScript or TypeScript
  - title: Know the boundaries
    details: Inspect verification status, aliases and transaction limits with migrate capabilities before choosing a database.
    link: ./databases
    linkText: Check database support
---

## What do you want to do?

| Your task | Start here |
| --- | --- |
| Try migrate for the first time | [SQLite quick start](./quick-start) |
| Set up a project | [Installation](./installation) · [Configuration](./configuration) |
| Deploy database changes | [Apply migrations](./applying) · [Rollback](./rollback) |
| Adopt an existing database | [History and baseline](./history) |
| Reuse script logic | [Shared modules](./modules) · [Runnable examples](./examples) |
| Integrate with an application or network | [Go library](./go-library) · [SSH tunnels](./ssh-tunnels) |
| Find a flag or resolve an error | [CLI reference](./commands) · [Troubleshooting](./troubleshooting) |

::: info Version availability
This site follows the repository's main branch. Unreleased changes may appear here;
check the release notes for your installed version with `migrate version`.
:::
