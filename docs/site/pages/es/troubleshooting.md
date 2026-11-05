# Solución de problemas

Empieza por `migrate version`, `migrate capabilities --output json` y `migrate status --output json`. Solo status necesita conexión.

## Comando o configuración ausente

Revisa PATH en la [instalación](./installation). Genera la configuración en el proyecto o usa `-c` / `-w`. En YAML comprueba la posición del error, los valores y la sangría con espacios, no tabulaciones. Consulta el [orden de búsqueda](./configuration#configuration-file-search-order).

## Variable ausente

`can't find env value for XXX` pide definir `${XXX}`. Usa `${XXX:default}` solo para valores no secretos apropiados. `password_env` de SSH contiene el nombre de la variable.

## Formato, base o módulo incompatible

Comprueba `capabilities` y actualiza versiones antiguas. Para `unsupported dialect`, verifica los [alias](./databases#aliases). Revisa rutas relativas e instala npm de antemano. Las fuentes embebidas requieren `DependencyDir` para paquetes. Los paquetes dependientes de Node o asincronía siguen siendo incompatibles.

## Historial distinto o up sin cambios

Restaura archivos y nombres aplicados desde Git y añade versiones superiores. No evites errores usando baseline sobre un estado desconocido. Un comando silencioso puede funcionar: comprueba applied, baseline y outOfOrder.

## Fallo al revertir o conectar

`to-version must be set, or use --all` requiere destino o `--all`. El destino se conserva. Un script sin down no puede revertirse.

Inspecciona datos e historial antes de reintentar: DDL implícito, operaciones no transaccionales y cortes alrededor de COMMIT pueden dejar cambios. Baseline no los repara. Revisa credenciales SSH, known_hosts, proxy y conectividad desde el bastión. No hay reconexión ni repetición automática. Consulta registros controlados del servidor sin compartir credenciales ni datos.
