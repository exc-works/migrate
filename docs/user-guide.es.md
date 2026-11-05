# Guía de Usuario de migrate

## 1. Instalación

### 1.1 Instalar desde módulo

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

Instala una versión específica:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

Reemplaza `vX.Y.Z` por una versión real, por ejemplo `v0.2.3`.

### 1.2 Instalar desde código fuente local (repo privado o red interna)

Ejecuta en la raíz del repositorio:

```bash
go install ./cmd/migrate
```

### 1.3 Verificar la instalación

```bash
migrate --help
```

Si no se encuentra el comando:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Si ves `Repository not found`, usa la ruta de instalación desde código fuente local indicada arriba.

## 2. Inicialización

### 2.1 Generar archivo de configuración

```bash
migrate new config
```

Opcional:

```bash
migrate new config dev.yml
migrate new config --force
```

Plantilla de configuración por defecto:

```yaml
schema_name: migration_schema
dialect: postgres
data_source_name: "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:postgres} sslmode=disable"
working_directory: ""
migrate_out_of_order: false
logger_level: info
migration_source: migrations
```

### 2.2 Actualizar campos clave de configuración

- `dialect`: `postgres`, `mysql`, `mariadb`, `oracle`, `sqlite`, `mssql`, `clickhouse`, `tidb`, `redshift`
- `data_source_name`: cadena de conexión de la DB
- `migration_source`: directorio de migraciones (por defecto: `migrations`)

### 2.3 Inicializar la tabla de historial de migraciones

```bash
migrate create
```

`create` puede completarse sin salida. Confírmalo con:

```bash
migrate status
```

Si ya tienes un esquema existente y no quieres reejecutar SQL antiguo, usa:

```bash
migrate baseline
```

## 3. Crear archivos de versión de migración

### 3.1 Versión autogenerada

```bash
migrate new version "init users"
```

### 3.2 Versión explícita

```bash
migrate new version "add email" -v 202604140002
```

Formato de nombre de archivo generado:

```text
V<version>__<description>.sql
```

Plantilla de archivo por defecto:

```sql
-- +migrate Up

-- +migrate Down
```

## 4. Actualización (aplicar migraciones)

Primero ejecuta dry run:

```bash
migrate up --dry-run
```

Aplicar de forma real:

```bash
migrate up
```

Luego verifica el estado:

```bash
migrate status
```

`up` puede completarse sin salida. Usa `status` como fuente de verdad.

## 5. Reversión

### 5.1 Volver a una versión objetivo (la versión objetivo se mantiene)

```bash
migrate down 202604140001
```

Semántica: solo se revierten las versiones aplicadas mayores que `202604140001`.

### 5.2 Revertir todas las versiones aplicadas

```bash
migrate down --all
```


### 5.3 Reversión en dry-run

```bash
migrate down 202604140001 --dry-run
migrate down --all --dry-run
```

`down` puede completarse sin salida. Ejecuta `migrate status` para verificar.

## 6. Verificar estado

```bash
migrate status
```

Salida legible por máquina (recomendada para scripts y agentes de IA):

```bash
migrate status --output json
```

Columnas de salida: `Version`, `Filename`, `Hash`, `Status`.

Estados comunes:

- `pending`
- `applied`
- `baseline`
- `outOfOrder`
- `hashMismatch`
- `filenameMismatch`

## 7. Consultar la versión

```bash
migrate version
```

## 8. Plantillas de variables de entorno

`data_source_name` soporta:

- `${KEY}`: obligatorio, debe existir
- `${KEY:default}`: usa `default` si falta `KEY`

Ejemplo:

```yaml
dialect: postgres
data_source_name: "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
```

Asegúrate de que `DB_PASSWORD` ya esté definido en tu entorno, luego ejecuta:

```bash
migrate status
```

## 9. Demo de primera ejecución en 10 minutos (SQLite)

[Demo de primera ejecución en 10 minutos (SQLite) · English](site/pages/quick-start.md)

## 10. Flags globales

Usar archivo de configuración específico:

```bash
migrate -c ./configs/dev.yml status
```

Usar directorio de trabajo específico:

```bash
migrate -w ./deploy create
migrate -w ./deploy up
```

## 11. Errores comunes y solución de problemas

### 11.1 No se encontró el archivo de configuración

Error: `config file ... no such file or directory`

Solución:

- asegúrate de que `migration_config.yml` exista en el directorio actual
- o pasa la ruta de configuración con `-c`

### 11.2 Falta variable de entorno

Error: `can't find env value for XXX`

Solución:

- `export XXX=...`
- o usa `${XXX:default}`

### 11.3 Argumentos `down` incompletos

Error: `to-version must be set, or use --all`

Solución:

- usa `migrate down <version>`
- o usa `migrate down --all`

### 11.4 Dialect no soportado

Error: `unsupported dialect: xxx`

Solución: usa uno de:

- `postgres` `mysql` `mariadb` `oracle` `sqlite` `mssql` `clickhouse` `tidb` `redshift`

### 11.5 Incompatibilidad de metadatos de migración

Error: `hash mismatch` o `filename mismatch`

Solución:

- restaura el archivo original desde el control de versiones
- crea una nueva migración de versión superior para los cambios

## 12. Usar migrate como librería Go

Además del CLI, `github.com/exc-works/migrate` puede importarse directamente desde el código de tu servicio para ejecutar migraciones — útil para pruebas unitarias, hooks de arranque o paneles de administración.

### 12.1 Instalación

```bash
go get github.com/exc-works/migrate
```

Importa el driver de base de datos que necesites (la librería no fija ninguno):

```go
import (
    _ "modernc.org/sqlite"             // sqlite
    _ "github.com/jackc/pgx/v5/stdlib" // postgres
    _ "github.com/go-sql-driver/mysql" // mysql / mariadb / tidb
    // ...
)
```

### 12.2 Ejemplo mínimo

```go
package main

import (
    "context"
    "database/sql"

    _ "modernc.org/sqlite"

    "github.com/exc-works/migrate"
)

func main() {
    db, err := sql.Open("sqlite", "./app.sqlite")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    svc, err := migrate.NewService(context.Background(), migrate.Config{
        Dialect:         migrate.NewSQLiteDialect(),
        DB:              db,
        MigrationSource: migrate.DirectorySource{Directory: "./migrations"},
    })
    if err != nil {
        panic(err)
    }

    if err := svc.Create(); err != nil { // idempotente: crea la tabla de historial si no existe
        panic(err)
    }
    if err := svc.Up(); err != nil {
        panic(err)
    }
}
```

### 12.3 API principal

- `migrate.NewService(ctx, migrate.Config)` construye un ejecutor de migraciones
- `svc.Create()` crea la tabla de historial `migration_schema` (idempotente)
- `svc.Up()` aplica todas las migraciones pendientes
- `svc.Down(toVersion, all)` revierte hasta una versión objetivo o todo
- `svc.Status()` devuelve `[]migrate.MigrationStatus`
- `svc.Baseline()` marca los archivos pendientes existentes como `baseline`

Tipos comunes:

- Dialectos (prefiere los constructores — devuelven la interfaz `Dialect`): `migrate.NewPostgresDialect()`, `NewMySQLDialect()`, `NewSQLiteDialect()`, `NewMSSQLDialect()`, `NewOracleDialect()`, `NewClickHouseDialect()`, `NewMariaDBDialect()`, `NewTiDBDialect()`, `NewRedshiftDialect()`, o `migrate.DialectFromName("postgres")` para resolución por nombre
- Fuentes: `DirectorySource` (sistema de archivos), `StringSource` (slice en memoria), `FSSource` (cualquier `fs.FS`, p. ej. `//go:embed` o `os.DirFS`), `CombinedSource` (combina varias fuentes)
- Loggers: `migrate.NoopLogger{}` (por defecto), `migrate.NewStdLogger("info", os.Stdout)` o cualquier tipo que implemente `migrate.Logger`

[Go API](../migrate.go) · [Go example](../example_test.go)
