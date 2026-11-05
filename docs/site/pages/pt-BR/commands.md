# Referência da CLI

Confira `migrate --help`, `migrate <command> --help` e `migrate version`.

| Opção global | Efeito |
| --- | --- |
| `-c`, `--config` | Arquivo específico; senão [busca automática](./configuration#configuration-file-search-order) |
| `-w`, `--working-dir` | Substitui o diretório de trabalho |
| `--non-interactive` | Falha em vez de pedir segredo ou confiar em chave SSH desconhecida |

## Gerar

```sh
migrate new config [filename] [--format yaml|json] [--multi] [--shards N] [--ssh-tunnel] [--force]
migrate new config shards.yml --multi --shards 4
migrate new config shards-ssh.yml --multi --ssh-tunnel --proxy socks5
```

Padrão `migration_config.yml`. `--shards N` exige `--multi` e número positivo, a partir de 0. `--proxy` exige `--ssh-tunnel`. Autenticação com `--ssh-auth agent|private_key|password`. `--force` sobrescreve; `.json` ou `--format json` escolhe JSON.

```sh
migrate new version <description> [-v <version>] [--format sql|js|ts|mjs]
```

Padrão SQL e versão automática ou `-v`. Descrições como `"init users"` usam aspas; espaços viram `_` no nome.

## Banco de dados

```sh
migrate create
```

Cria só o histórico, de forma idempotente.

```sh
migrate baseline
```

Registra sem executar, apenas se o banco já corresponde aos arquivos. Veja [baseline](./history#baseline).

```sh
migrate up [--dry-run]
```

Aplica pendentes em ordem. Dry-run pode ler histórico, sem aplicar mudanças.

```sh
migrate down <to-version> [--dry-run]
migrate down --all [--dry-run]
```

Reverte superiores ao destino, preservando-o, ou todas. Escolha destino ou `--all`, não ambos.

```sh
migrate status [--output table|json]
```

Tabela por padrão. JSON multibanco ordenado por nome contém `data_source`, `status`, `migrations` no sucesso ou `error` no erro. create/baseline/status/up/down percorrem todas as conexões, com `--parallel N` (8). Um erro não impede os outros e produz saída não zero.

## Informações sem conexão

```sh
migrate capabilities [--output text|json]
```

Formatos, aliases, scripts e transações; padrão `text`.

```sh
migrate version
```

Versão da ferramenta, não das migrações.
