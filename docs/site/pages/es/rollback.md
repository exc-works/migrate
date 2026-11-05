# Revertir migraciones

Ejecuta Down o `down(db)`. Solo deshace lo que implementa ese código; no recupera automáticamente datos eliminados.

## Conservar una versión

```sh
migrate down 202611030001 --dry-run
migrate down 202611030001
migrate status
```

Solo revierte las superiores. Con 1, 2 y 3 aplicadas, `down 2` revierte 3 y conserva 1 y 2.

## Revertir todas

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

El destino y `--all` son excluyentes. Tras el éxito, los archivos vuelven a `pending` y pueden aplicarse otra vez.

Revisa todo el rango y las [transacciones](./script-transactions). Un script sin `down(db)` bloquea el rango; un down vacío no deshace nada. DDL y operaciones no transaccionales pueden quedar tras un fallo. Inspecciona antes de recuperar.
