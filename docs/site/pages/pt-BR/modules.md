# Módulos compartilhados

Importação e reexportação estáticas de `.js`, `.ts`, `.mjs`. Auxiliares não executam como migrações.

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

Caminhos partem do arquivo importador. Defina `normalize` em `shared/names.ts` e crie `users` antes. Veja o [exemplo](./examples#local).

Instale npm pelo lockfile antes de migrate. Há suporte a exports, subcaminhos, dependências aninhadas e CommonJS estático, usando `neutral`, depois `module` e `main`. Não há API Node, extensões nativas ou import/require dinâmico. migrate não instala pacotes.

## Arquivos embutidos

`FSSource` resolve em seu `fs.FS`, inclusive `embed.FS`, com extensões e `index`. Não resolve `package.json` nem `node_modules` virtuais. Para npm, informe um diretório real:

```go
source := migrate.WithModules{
    Source: migrate.FSSource{
        FS:   embeddedFiles,
        Root: "migrations",
    },
    DependencyDir: projectDirectory,
}
```

Importações relativas continuam virtuais. Sem `DependencyDir`, não busca no host. `StringSource` aceita `WithModules` com `FS`/`Root` ou `Directory`. Fontes próprias podem implementar `MigrationEntriesSource`. Veja [Go](./go-library).

## Reprodutibilidade {#checksums}

O hash cobre só o arquivo de entrada. Alterar auxiliar, pacote ou lockfile não acusa divergência nem repete migrações. Fixe dependências e preserve o comportamento antigo. Todas as dependências selecionadas são compiladas antes de escrever; um erro posterior impede etapas anteriores.
