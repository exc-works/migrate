# Общие модули

Статический import и реэкспорт `.js`, `.ts`, `.mjs`. Вспомогательные файлы не являются отдельными миграциями.

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

Путь от импортирующего файла. Определите `normalize` в `shared/names.ts` и заранее создайте `users`. См. [пример](./examples#local).

npm устанавливайте по lockfile заранее. Поддерживаются exports, подпути, вложенные зависимости и статический CommonJS, условия `neutral`, затем `module` и `main`. Без Node API, нативных расширений, динамических import/require. migrate пакеты не устанавливает.

## Встроенные источники

`FSSource` разрешает пути внутри `fs.FS`, включая `embed.FS`, расширения и `index`. Виртуальные `package.json` и `node_modules` не ищет. npm нужен реальный каталог:

```go
source := migrate.WithModules{
    Source: migrate.FSSource{
        FS:   embeddedFiles,
        Root: "migrations",
    },
    DependencyDir: projectDirectory,
}
```

Относительные модули остаются виртуальными. Без `DependencyDir` поиска на хосте нет. `StringSource` использует `WithModules` с `FS`/`Root` или `Directory`. Свой источник может реализовать `MigrationEntriesSource`. См. [Go](./go-library).

## Воспроизводимость {#checksums}

Хешируется только входной файл. Правки модулей, пакетов или lockfile не вызывают несовпадение и повтор применённого. Фиксируйте зависимости, сохраняйте старое поведение. До записи компилируются зависимости всего диапазона; ошибка поздней миграции блокирует и ранние.
