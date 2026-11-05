# Konfiguration

Erzeugen und bearbeiten Sie `migration_config.yml` im Projektverzeichnis.

```sh
migrate new config
```

## Datenbank verbinden

PostgreSQL mit einer Verbindungszeichenfolge aus `DATABASE_DSN`:

```yaml
# Datenbanktyp
dialect: postgres
# DATABASE_DSN vor der Ausführung setzen
data_source_name: "${DATABASE_DSN}"
# Verzeichnis der Migrationsdateien
migration_source: migrations
# Tabelle der ausgeführten Versionen
schema_name: migration_schema
```

| Feld | Bedeutung | Standard |
| --- | --- | --- |
| `dialect` | Datenbanktyp | `postgres` |
| `data_source_name` | Einzelverbindung, nicht zusammen mit `data_sources` | Im Einzelmodus erforderlich |
| `data_sources` | Benannte Verbindungen | Deaktiviert |
| `migration_source` | Migrationsverzeichnis | `migrations` |
| `schema_name` | Verlaufstabelle | `migration_schema` |
| `working_directory` | Ausgangspunkt der Dateisuche | Aktuelles Verzeichnis |
| `migrate_out_of_order` | Ältere offene Versionen ausführen | `false` |
| `logger_level` | `debug`, `info`, `warn`, `error` | `info` |

### Verbindungsbeispiele

Ersetzen Sie Host, Port, Datenbank beziehungsweise Oracle-Dienst und setzen Sie `DB_USER` und `DB_PASSWORD`. SQLite verwendet einen Dateipfad.

::: code-group

```yaml [PostgreSQL]
dialect: postgres
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@db.example:5432/app?sslmode=require"
```

```yaml [MySQL]
dialect: mysql
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [MariaDB]
dialect: mariadb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [SQLite]
dialect: sqlite
data_source_name: "./demo.sqlite"
```

```yaml [Oracle]
dialect: oracle
data_source_name: "oracle://${DB_USER}:${DB_PASSWORD}@db.example:1521/FREEPDB1"
```

```yaml [SQL Server]
dialect: mssql
data_source_name: "sqlserver://${DB_USER}:${DB_PASSWORD}@db.example:1433?database=app"
```

```yaml [ClickHouse]
dialect: clickhouse
data_source_name: "clickhouse://${DB_USER}:${DB_PASSWORD}@db.example:9000/app"
```

```yaml [TiDB]
dialect: tidb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:4000)/app?parseTime=true"
```

```yaml [Redshift]
dialect: redshift
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@cluster.example:5439/app?sslmode=require"
```

:::

Sonderzeichen in Benutzernamen und Passwörtern von URLs (`@`, `:`, `/`) müssen URL-kodiert sein; die Variablenersetzung kodiert nicht automatisch. MySQL/MariaDB/TiDB verwenden mit `user:password@tcp(...)` kein URL-Format. TLS/SSL richten Sie nach den Serveranforderungen ein.

### Protokollstufen {#logging}

| Wert | Ausgaben |
| --- | --- |
| `debug` | Debug, Information, Warnung, Fehler |
| `info` | Information, Warnung, Fehler |
| `warn` | Warnung und Fehler; auch `warning` |
| `error` | Nur Fehler |

```yaml
logger_level: debug # Debug-Meldungen zur Fehlersuche anzeigen
```

## Umgebungsvariablen {#environment-variables}

`${DATABASE_DSN}` erfordert eine gesetzte Variable. `${DB_NAME:app}` verwendet andernfalls `app`. Dieselbe YAML-Datei kann verschiedene Umgebungen bedienen; Zugangsdaten kommen aus Variablen oder der Geheimnisverwaltung. SSH-Host, Benutzer und Proxyadresse unterstützen ebenfalls Vorlagen.

## Suchreihenfolge {#configuration-file-search-order}

Im Arbeitsverzeichnis wird die erste vorhandene Datei verwendet:

1. `migration_config.yml`
2. `migration_config.yaml`
3. `migration_config.json`

Dateien werden nicht kombiniert. Eine ungültige oder unlesbare gefundene Datei führt zum Fehler. Mit `-c` wird ausschließlich die angegebene Datei verwendet.

```sh
migrate -c ./configs/dev.yml status
```

## Konfiguration erzeugen

| Fall | Befehl |
| --- | --- |
| Eine Datenbank | `migrate new config` |
| Eigener Dateiname | `migrate new config dev.yml` |
| Mehrere Datenbanken | `migrate new config --multi` |
| Vier Datenbanken | `migrate new config --multi --shards 4` |
| SSH Agent | `migrate new config --ssh-tunnel` |
| Privater Schlüssel | `migrate new config --ssh-tunnel --ssh-auth private_key` |
| SSH-Passwort | `migrate new config --ssh-tunnel --ssh-auth password` |
| SOCKS5 | `migrate new config --ssh-tunnel --proxy socks5` |
| HTTP CONNECT | `migrate new config --ssh-tunnel --proxy http` |

Optionen sind kombinierbar. Überschreiben erfordert `--force`; alle Optionen stehen unter `migrate new config --help`.

```sh
migrate new config shards.yml --multi --shards 4 --ssh-tunnel --proxy socks5
```

## Mehrere Datenbanken {#multiple-databases}

Verwenden Sie `data_sources` anstelle von `data_source_name`. Typ, Dateien und Name der Verlaufstabelle sind gemeinsam. Erzeugte Namen beginnen mit `shard_0` und verwenden `SHARD_0_DSN` usw.

```yaml
dialect: postgres
migration_source: migrations
schema_name: migration_schema
data_sources:
  shard_0:
    data_source_name: "${SHARD_0_DSN}" # Verbindung der ersten Datenbank
  shard_1:
    data_source_name: "${SHARD_1_DSN}" # Verbindung der zweiten Datenbank
```

![Migrationsdateien → shard_0 / shard_1 → Ergebnisse](../assets/diagrams/multiple-databases-de.svg)

```sh
migrate create
migrate up --dry-run
migrate up
migrate status
```

Jede Datenbank ermittelt offene Versionen anhand ihres eigenen Verlaufs und führt sie geordnet aus. Bereits erfolgreiche Versionen werden übersprungen. `new version` erzeugt eine gemeinsame Datei. Standardmäßig laufen höchstens 8 Datenbanken gleichzeitig; `--parallel 1` arbeitet nacheinander.

```sh
migrate up --parallel 4
migrate status --output json
```

Fehler stoppen andere Datenbanken nicht. Die Ergebnisse werden je Verbindung zusammengefasst; ein Fehler ergibt einen Exitcode ungleich null. Es gibt keine datenbankübergreifende Transaktion oder Rücknahme anderer Erfolge. Ctrl+C stoppt neue Aufträge und bricht aktive Operationen ab.

Ein globales `ssh_tunnel` dient allen Verbindungen. Ein lokales Objekt ersetzt es vollständig; fehlt es oder ist es `null`, wird das globale geerbt. Siehe [SSH](./ssh-tunnels). Jede tatsächliche Datenbank nur einmal eintragen und überlappende Aufrufe serialisieren. Dateien und Abhängigkeiten während des Laufs unverändert lassen; vor Wiederholungen Daten und Verlauf prüfen.

## Arbeitsverzeichnis

```sh
migrate -w ./deploy up
migrate -w ./deploy -c configs/dev.yml status
```

`-w` bestimmt die Konfigurationssuche und überschreibt `working_directory`. Ein relatives `migration_source` bezieht sich darauf; `-c` allein ändert diesen Bezug nicht. Relative SSH-Dateipfade beziehen sich auf das Verzeichnis der Konfigurationsdatei.
