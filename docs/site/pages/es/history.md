# Historial y línea base

El historial guarda versión, nombre y hash. Compáralo con el esquema real.

```sh
migrate status
migrate status --output json
```

| Estado | Significado | Acción |
| --- | --- | --- |
| `pending` | Sin aplicar | Revisar y usar `up` |
| `applied` | Aplicación correcta | Conservar el archivo |
| `baseline` | Registrado sin ejecutar | Mantener el historial adoptado |
| `outOfOrder` | Pendiente más antiguo | Revisar el orden |
| `hashMismatch` | Contenido distinto | Restaurar el original |
| `filenameMismatch` | Nombre distinto | Restaurar el nombre |

Las diferencias bloquean la ejecución. Añade versiones para nuevos cambios. El hash de scripts no incluye los [módulos](./modules#checksums).

## Adoptar una base existente {#baseline}

Úsalo únicamente cuando datos y esquema coincidan con los archivos:

1. Compara los archivos y el estado real.
2. Ejecuta `migrate create` si hace falta.
3. Ejecuta `migrate baseline` y comprueba `migrate status`.
4. Añade versiones superiores para nuevos cambios.

Baseline registra sin ejecutar; no repara una migración parcialmente fallida.
