# Gemeinsame Module

Statische Importe und Re-Exporte von `.js`, `.ts`, `.mjs` sind möglich. Hilfsdateien werden nicht als eigene Migration ausgeführt.

## Relative Importe

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

Pfade gelten ab der importierenden Datei. `normalize` in `shared/names.ts` definieren und `users` vorher erstellen. Siehe [vollständiges Beispiel](./examples#local).

## npm-Pakete

Vor migrate anhand des Lockfiles installieren. Unterstützt werden exports, Unterpfade, verschachtelte Abhängigkeiten und statisches CommonJS, mit `neutral`, danach `module` und `main`. Keine Node-APIs, nativen Erweiterungen oder dynamischen import/require-Aufrufe. migrate installiert keine Pakete.

## Eingebettete Quellen

`FSSource` löst innerhalb seines `fs.FS` auf, auch `embed.FS`, mit Erweiterungssuche und `index`. Virtuelle `package.json` oder `node_modules` werden nicht durchsucht. Pakete benötigen ein reales Verzeichnis:

```go
source := migrate.WithModules{
    Source: migrate.FSSource{
        FS:   embeddedFiles,
        Root: "migrations",
    },
    DependencyDir: projectDirectory,
}
```

Relative Module bleiben virtuell. Ohne `DependencyDir` keine Suche im Hostverzeichnis. `StringSource` kann `WithModules` mit `FS`/`Root` oder `Directory` nutzen. Eigene Quellen können `MigrationEntriesSource` implementieren. Siehe [Go](./go-library).

## Reproduzierbarkeit {#checksums}

Nur die Einstiegsdatei wird gehasht. Änderungen an Helfern, Paketen oder Lockfile erzeugen weder Abweichungen noch Wiederholung angewandter Migrationen. Abhängigkeiten fixieren und altes Verhalten erhalten. Alle ausgewählten Abhängigkeiten werden vor Schreibzugriffen kompiliert; spätere Fehler blockieren auch frühere Schritte.
