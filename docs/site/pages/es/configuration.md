# Configuración

Desde el directorio del proyecto, genera y edita `migration_config.yml`:

```sh
migrate new config
```

## Conectar a la base

PostgreSQL con la conexión en `DATABASE_DSN`:

```yaml
# Tipo de base de datos
dialect: postgres
# Define DATABASE_DSN antes de ejecutar
data_source_name: "${DATABASE_DSN}"
# Directorio de los archivos de migración
migration_source: migrations
# Tabla de versiones aplicadas
schema_name: migration_schema
```

| Campo | Uso | Valor predeterminado |
| --- | --- | --- |
| `dialect` | Tipo de base | `postgres` |
| `data_source_name` | Conexión única; incompatible con `data_sources` | Obligatorio en modo único |
| `data_sources` | Conexiones con nombre | Desactivado |
| `migration_source` | Archivos de migración | `migrations` |
| `schema_name` | Tabla del historial | `migration_schema` |
| `working_directory` | Base para buscar archivos | Directorio actual |
| `migrate_out_of_order` | Aplicar versiones pendientes más antiguas | `false` |
| `logger_level` | `debug`, `info`, `warn`, `error` | `info` |

### Ejemplos de conexión

Cambia el host, puerto, base o servicio Oracle y define `DB_USER` y `DB_PASSWORD`. SQLite utiliza una ruta de archivo.

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

En las URL, codifica caracteres especiales del usuario y contraseña (`@`, `:`, `/`); la sustitución de variables no los codifica. El formato `user:password@tcp(...)` de MySQL/MariaDB/TiDB no usa codificación URL. Configura TLS/SSL según tu servidor.

### Niveles de registro {#logging}

| Valor | Mensajes |
| --- | --- |
| `debug` | Depuración, información, advertencias y errores |
| `info` | Información, advertencias y errores |
| `warn` | Advertencias y errores; también acepta `warning` |
| `error` | Solo errores |

```yaml
logger_level: debug # Mostrar mensajes de depuración
```

## Variables de entorno {#environment-variables}

`${DATABASE_DSN}` exige una variable definida; `${DB_NAME:app}` usa `app` si falta. Comparte el YAML entre entornos y proporciona credenciales mediante variables o el almacén de secretos. El host y usuario SSH y la dirección del proxy también admiten plantillas.

## Orden de búsqueda {#configuration-file-search-order}

En el directorio de trabajo se usa el primer archivo existente:

1. `migration_config.yml`
2. `migration_config.yaml`
3. `migration_config.json`

No se combinan. Si el seleccionado es inválido o no puede leerse, se devuelve un error. Con `-c` se usa únicamente el archivo indicado, aunque no exista.

```sh
migrate -c ./configs/dev.yml status
```

## Generar según el escenario

| Escenario | Comando |
| --- | --- |
| Una base | `migrate new config` |
| Nombre personalizado | `migrate new config dev.yml` |
| Varias bases | `migrate new config --multi` |
| Cuatro bases | `migrate new config --multi --shards 4` |
| SSH Agent | `migrate new config --ssh-tunnel` |
| Clave privada | `migrate new config --ssh-tunnel --ssh-auth private_key` |
| Contraseña SSH | `migrate new config --ssh-tunnel --ssh-auth password` |
| SOCKS5 | `migrate new config --ssh-tunnel --proxy socks5` |
| HTTP CONNECT | `migrate new config --ssh-tunnel --proxy http` |

Combina las opciones. Para sobrescribir un archivo usa `--force`. Consulta `migrate new config --help`.

```sh
migrate new config shards.yml --multi --shards 4 --ssh-tunnel --proxy socks5
```

## Varias bases {#multiple-databases}

Sustituye `data_source_name` por `data_sources`. Comparten tipo, directorio y nombre de tabla del historial. Los nombres generados empiezan en `shard_0`, con variables `SHARD_0_DSN`, etc.

```yaml
dialect: postgres
migration_source: migrations
schema_name: migration_schema
data_sources:
  shard_0:
    data_source_name: "${SHARD_0_DSN}" # Conexión de la primera base
  shard_1:
    data_source_name: "${SHARD_1_DSN}" # Conexión de la segunda base
```

![Archivos → shard_0 / shard_1 → Resultados](../assets/diagrams/multiple-databases-es.svg)

```sh
migrate create
migrate up --dry-run
migrate up
migrate status
```

Cada base selecciona sus versiones pendientes según su propio historial y las ejecuta en orden. No repite las aplicadas. `new version` crea un solo archivo compartido. El máximo simultáneo predeterminado es 8; `--parallel 1` ejecuta secuencialmente.

```sh
migrate up --parallel 4
migrate status --output json
```

Un fallo no detiene otras bases. Se resumen los resultados por nombre; cualquier fallo devuelve un código no nulo. No hay transacción entre bases ni se revierten cambios ajenos al fallo. Ctrl+C deja de iniciar bases y cancela operaciones activas.

El `ssh_tunnel` global se comparte. Un objeto local lo sustituye completamente; ausente o `null`, lo hereda. Consulta [SSH](./ssh-tunnels). Configura cada base real una sola vez y serializa las ejecuciones que la afectan. No cambies archivos ni dependencias durante la ejecución; inspecciona datos e historial antes de reintentar.

## Directorio de trabajo

```sh
migrate -w ./deploy up
migrate -w ./deploy -c configs/dev.yml status
```

`-w` determina la búsqueda de configuración y reemplaza `working_directory`. `migration_source` relativo parte de ese directorio; usar solo `-c` no lo cambia. Las rutas SSH relativas parten del directorio del archivo de configuración.
