# Referencia CLI

Comprueba tu versión con `migrate version` y los argumentos con `migrate --help` o `migrate <command> --help`.

## Opciones globales

| Opción | Efecto |
| --- | --- |
| `-c`, `--config` | Selecciona archivo; si se omite usa el [orden de búsqueda](./configuration#configuration-file-search-order) |
| `-w`, `--working-dir` | Reemplaza el directorio de trabajo |
| `--non-interactive` | Falla sin pedir credenciales ni confianza para claves SSH nuevas |

## Generar archivos

```sh
migrate new config [filename] [--format yaml|json] [--multi] [--shards N] [--ssh-tunnel] [--force]
migrate new config shards.yml --multi --shards 4
migrate new config shards-ssh.yml --multi --ssh-tunnel --proxy socks5
```

Predeterminado `migration_config.yml`. `--shards N` requiere `--multi` y un número positivo desde 0; `--proxy` requiere `--ssh-tunnel`. Autenticación con `--ssh-auth agent|private_key|password`. `--force` sobrescribe; `.json` o `--format json` produce JSON.

```sh
migrate new version <description> [-v <version>] [--format sql|js|ts|mjs]
```

SQL por defecto y versión automática o `-v`. Usa `"init users"` para descripciones con espacios, sustituidos por `_` en el archivo.

## Operaciones

```sh
migrate create
```

Crea solo el historial, de forma idempotente.

```sh
migrate baseline
```

Registra sin ejecutar, solo si la base ya corresponde a los archivos. Consulta [baseline](./history#baseline).

```sh
migrate up [--dry-run]
```

Aplica pendientes en orden; dry-run puede leer el historial, sin cambiar datos.

```sh
migrate down <to-version> [--dry-run]
migrate down --all [--dry-run]
```

Revierte por encima del destino, conservándolo, o todo. Elige destino o `--all`, nunca ambos.

```sh
migrate status [--output table|json]
```

Tabla por defecto. JSON multibase ordenado por nombre: `data_source`, `status`, y `migrations` en éxito o `error` en fallo.

create/baseline/status/up/down recorren todas las bases y aceptan `--parallel N` (8). Un fallo no detiene otras bases y produce código no nulo.

## Información sin conexión

```sh
migrate capabilities [--output text|json]
```

Formatos, alias, compatibilidad de scripts y transacciones; predeterminado `text`.

```sh
migrate version
```

Versión de la herramienta, no de las migraciones.
