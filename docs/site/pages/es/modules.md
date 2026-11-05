# Módulos compartidos

Importa o reexporta estáticamente `.js`, `.ts` y `.mjs`. Los archivos auxiliares no son migraciones independientes.

## Importaciones relativas

```text
project/
  migrations/V2__normalize.ts
  shared/names.ts
```

```ts
import { normalize } from "../shared/names.ts";

export function up(db) {
  db.exec("UPDATE users SET name=? WHERE id=?", [normalize(" Alice "), "1"]);
}
```

La ruta parte del archivo importador. Define `normalize` en `shared/names.ts` y crea `users` antes. Consulta el [ejemplo completo](./examples#local).

## Paquetes npm

Instala desde el lockfile antes de ejecutar migrate. Se admiten exports, subrutas, dependencias anidadas y CommonJS estático, con condiciones `neutral`, luego `module` y `main`. No se admiten API Node, extensiones nativas ni import/require dinámico. migrate no instala paquetes.

## Archivos embebidos

`FSSource` resuelve dentro de `fs.FS`, incluido `embed.FS`, con búsqueda de extensión e `index`. No busca `package.json` ni `node_modules` virtuales. Para paquetes, indica un directorio real:

```go
source := migrate.WithModules{
    Source: migrate.FSSource{
        FS:   embeddedFiles,
        Root: "migrations",
    },
    DependencyDir: projectDirectory,
}
```

Los archivos relativos siguen en el sistema virtual. Sin `DependencyDir`, no se busca en el host. `StringSource` acepta `WithModules` con `FS`/`Root` o `Directory`. Las fuentes propias pueden implementar `MigrationEntriesSource`. Consulta [Go](./go-library).

## Reproducibilidad {#checksums}

Solo se verifica el archivo de entrada. Cambiar auxiliares, paquetes o lockfile no causa divergencia ni reejecuta versiones aplicadas. Fija dependencias y preserva el comportamiento histórico. Todas las dependencias del rango se compilan antes de escribir; un error posterior bloquea los primeros pasos.
