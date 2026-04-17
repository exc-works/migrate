# Guía de Usuario de migrate

Esta guía es para usuarios que lo usan por primera vez. Los comandos y flags se basan en la implementación actual (`cmd/migrate`).

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
migrate new config dev.json
migrate new config --force
```

Plantilla de configuración por defecto:

```json
{
  "schema_name": "migration_schema",
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD:change_me} dbname=${DB_NAME:postgres} sslmode=disable",
  "working_directory": "",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
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
migrate new version init_users
```

### 3.2 Versión explícita

```bash
migrate new version add_email -v 202604140002
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

Nota: `migrate down <to-version>` y `migrate down --all` son mutuamente excluyentes.

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

## 7. Actualizar o degradar la herramienta en sí

### 7.1 Actualizar versión de la herramienta

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

### 7.2 Degradar versión de la herramienta

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

Ejemplo:

```bash
go install github.com/exc-works/migrate/cmd/migrate@v0.2.3
```

Si el repo es privado y `go install github.com/...@...` no está disponible, haz checkout de la versión objetivo en el código fuente
y ejecuta:

```bash
go install ./cmd/migrate
```

### 7.3 Verificar versión actual de la herramienta

```bash
migrate version
```

Nota: los artefactos de release muestran la versión de release; las builds locales desde `go install ./cmd/migrate` normalmente muestran `dev`.

## 8. Plantillas de variables de entorno

`data_source_name` soporta:

- `${KEY}`: obligatorio, debe existir
- `${KEY:default}`: usa `default` si falta `KEY`

Ejemplo:

```json
{
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
}
```

Asegúrate de que `DB_PASSWORD` ya esté definido en tu entorno, luego ejecuta:

```bash
migrate status
```

## 9. Demo de primera ejecución en 10 minutos (SQLite)

### 9.1 Preparar directorio y configuración

Primero verifica la disponibilidad del comando:

```bash
migrate --help
```

Crear directorio de demo (macOS/Linux):

```bash
mkdir -p ./migrate-demo
cd ./migrate-demo
migrate new config
```

Equivalente en Windows PowerShell:

```powershell
mkdir .\migrate-demo
cd .\migrate-demo
migrate new config
```

Actualiza `migration_config.json` a:

```json
{
  "schema_name": "migration_schema",
  "dialect": "sqlite",
  "data_source_name": "./demo.sqlite",
  "working_directory": ".",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 9.2 Inicializar y crear archivos de migración

```bash
migrate create
migrate new version init_users -v 202604140001
migrate new version add_email -v 202604140002
```

Edita `migrations/V202604140001__init_users.sql`:

```sql
-- +migrate Up
CREATE TABLE users
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS users;
```

Edita `migrations/V202604140002__add_email.sql`:

```sql
-- +migrate Up
ALTER TABLE users
    ADD COLUMN email TEXT;

-- +migrate Down
CREATE TABLE users_tmp
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
INSERT INTO users_tmp (id, name)
SELECT id, name
FROM users;
DROP TABLE users;
ALTER TABLE users_tmp
    RENAME TO users;
```

### 9.3 Aplicar, verificar estado y revertir

```bash
migrate up --dry-run
migrate up
migrate status
migrate down 202604140001 --dry-run
migrate down 202604140001
migrate status
migrate down --all
migrate status
```

Esperado:

- después de `up`: ambas versiones son `applied`
- después de `down 202604140001`: `202604140001=applied`, `202604140002=pending`
- después de `down --all`: ambas versiones son `pending`

## 10. Flags globales

Usar archivo de configuración específico:

```bash
migrate -c ./configs/dev.json status
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

- asegúrate de que `migration_config.json` exista en el directorio actual
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

- no edites archivos de migración ya aplicados
- crea una nueva migración de versión superior para los cambios
