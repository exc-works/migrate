# CLI-Referenz

Version über `migrate version`, Optionen über `migrate --help` oder `migrate <command> --help` prüfen.

## Globale Optionen

| Option | Wirkung |
| --- | --- |
| `-c`, `--config` | Datei wählen; sonst [automatische Suche](./configuration#configuration-file-search-order) |
| `-w`, `--working-dir` | Arbeitsverzeichnis überschreiben |
| `--non-interactive` | Statt Geheimnisabfrage oder SSH-Vertrauensbestätigung fehlschlagen |

## Dateien erzeugen

```sh
migrate new config [filename] [--format yaml|json] [--multi] [--shards N] [--ssh-tunnel] [--force]
migrate new config shards.yml --multi --shards 4
migrate new config shards-ssh.yml --multi --ssh-tunnel --proxy socks5
```

Standard `migration_config.yml`. `--shards N` verlangt `--multi` und eine positive Anzahl, ab 0 nummeriert. `--proxy` verlangt `--ssh-tunnel`. Authentifizierung: `--ssh-auth agent|private_key|password`. `--force` überschreibt; `.json` oder `--format json` erzeugt JSON.

```sh
migrate new version <description> [-v <version>] [--format sql|js|ts|mjs]
```

Standard SQL, Version automatisch oder mit `-v`. Beschreibungen wie `"init users"` zitieren; Leerzeichen werden im Dateinamen `_`.

## Datenbankoperationen

```sh
migrate create
```

Nur Verlaufstabelle idempotent erstellen.

```sh
migrate baseline
```

Ohne Ausführung erfassen, nur wenn die DB den Dateien entspricht. Siehe [Baseline](./history#baseline).

```sh
migrate up [--dry-run]
```

Offene Versionen ausführen. Dry-run liest eventuell Verlauf, ohne Änderungen anzuwenden.

```sh
migrate down <to-version> [--dry-run]
migrate down --all [--dry-run]
```

Versionen oberhalb des Ziels oder alle zurücknehmen. Ziel bleibt erhalten; Ziel und `--all` schließen sich aus.

```sh
migrate status [--output table|json]
```

Standard Tabelle. Mehrdatenbank-JSON ist nach Namen sortiert: `data_source`, `status`, bei Erfolg `migrations`, bei Fehler `error`.

create/baseline/status/up/down bearbeiten alle Verbindungen mit `--parallel N` (Standard 8). Ein Fehler stoppt andere nicht und ergibt einen Exitcode ungleich null.

## Informationen ohne Verbindung

```sh
migrate capabilities [--output text|json]
```

Formate, Aliase, Skriptunterstützung und Transaktionen; Standard `text`.

```sh
migrate version
```

Werkzeugversion, nicht Migrationsversion.
