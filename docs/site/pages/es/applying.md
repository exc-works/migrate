# Aplicar migraciones

Comprueba la base de destino y los archivos antes de ejecutar.

## Inicializar el historial

```sh
migrate create
```

Crea la tabla si falta, sin ejecutar los archivos. Si los cambios ya existen, consulta la [línea base](./history#baseline).

## Previsualizar y aplicar

```sh
migrate up --dry-run
migrate up
migrate status
```

Aplica versiones pendientes en orden y registra los éxitos. Repetir no vuelve a aplicarlas. Puede no haber salida; comprueba `status`.

Dry-run puede conectar y abrir SSH para leer el historial, pero no aplica cambios. Comprueba scripts y dependencias sin evaluarlos y no predice SQL dinámico.

## Versiones antiguas

Se omiten por defecto si son inferiores a versiones aplicadas. Activa `migrate_out_of_order` solo después de comprobar su compatibilidad con el esquema actual.

## Desplegar

```sh
migrate --non-interactive up
migrate --non-interactive status --output json
```

Proporciona variables y secretos y serializa las ejecuciones contra una base. El historial no garantiza una única ejecución ante concurrencia, cierres inesperados o commits implícitos. Tras fallos, inspecciona datos e historial antes de reintentar. Consulta [problemas](./troubleshooting).
