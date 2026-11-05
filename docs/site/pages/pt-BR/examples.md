# Executar exemplos

Na cópia do repositório, todos usam SQLite e validam seus dados internamente.

## CLI temporária

Mantenha o mesmo shell aberto.

```sh
example_workspace=$(mktemp -d)
go build -o "$example_workspace/migrate" ./cmd/migrate
```

## Auxiliar local {#local}

[Fontes](https://github.com/exc-works/migrate/tree/main/examples/script-imports/local): duas migrações compartilham TypeScript.

```sh
export MIGRATE_EXAMPLE_DB="$example_workspace/local.sqlite"
"$example_workspace/migrate" -w examples/script-imports/local create
"$example_workspace/migrate" -w examples/script-imports/local up --dry-run
"$example_workspace/migrate" -w examples/script-imports/local up
"$example_workspace/migrate" -w examples/script-imports/local status
"$example_workspace/migrate" -w examples/script-imports/local down 2
"$example_workspace/migrate" -w examples/script-imports/local down --all
```

Após up há três versões e nomes `alice`, `bob`. `down 2` mantém 1 e 2; `down --all` apaga a tabela.

## Pacote npm

Node/npm só é usado para instalar. O lockfile referencia um pacote local, sem registro externo.

```sh
npm --prefix examples/script-imports/npm ci --ignore-scripts --no-audit --no-fund --offline
export MIGRATE_EXAMPLE_DB="$example_workspace/npm.sqlite"
"$example_workspace/migrate" -w examples/script-imports/npm create
"$example_workspace/migrate" -w examples/script-imports/npm up
"$example_workspace/migrate" -w examples/script-imports/npm status
"$example_workspace/migrate" -w examples/script-imports/npm down --all
```

A segunda migração importa `@migrate-example/normalizer/names` e verifica `alice`. Duas versões são aplicadas.

## Embutir em Go

```sh
go run ./examples/script-imports/embedded
```

Usa `embed.FS` e banco privado em memória, sem npm. Saída:

```text
after up: alice
after down: all migrations pending
```

## Limpeza

```sh
rm -r "$example_workspace"
```

npm deixa `node_modules` ignorado. Execute `./scripts/test-script-examples.sh` na raiz para validar cópias temporárias. Veja [reprodutibilidade](./modules#checksums).
