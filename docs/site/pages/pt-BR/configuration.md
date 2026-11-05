# Configuração

Gere e edite `migration_config.yml` no diretório do projeto:

```sh
migrate new config
```

## Conectar ao banco

PostgreSQL com a conexão na variável `DATABASE_DSN`:

```yaml
# Tipo de banco
dialect: postgres
# Defina DATABASE_DSN antes da execução
data_source_name: "${DATABASE_DSN}"
# Diretório dos arquivos de migração
migration_source: migrations
# Tabela das versões aplicadas
schema_name: migration_schema
```

| Campo | Uso | Padrão |
| --- | --- | --- |
| `dialect` | Tipo de banco | `postgres` |
| `data_source_name` | Conexão única; não combina com `data_sources` | Obrigatório no modo único |
| `data_sources` | Conexões nomeadas | Desativado |
| `migration_source` | Diretório dos arquivos | `migrations` |
| `schema_name` | Tabela do histórico | `migration_schema` |
| `working_directory` | Base para buscar arquivos | Diretório atual |
| `migrate_out_of_order` | Aplicar versões pendentes mais antigas | `false` |
| `logger_level` | `debug`, `info`, `warn`, `error` | `info` |

### Exemplos de conexão

Troque host, porta, banco ou serviço Oracle e defina `DB_USER` e `DB_PASSWORD`. SQLite usa um caminho de arquivo.

::: code-group

```yaml [PostgreSQL]
dialect: postgres
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@db.example:5432/app?sslmode=require"
```

```yaml [MySQL]
dialect: mysql
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [MariaDB]
dialect: mariadb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [SQLite]
dialect: sqlite
data_source_name: "./demo.sqlite"
```

```yaml [Oracle]
dialect: oracle
data_source_name: "oracle://${DB_USER}:${DB_PASSWORD}@db.example:1521/FREEPDB1"
```

```yaml [SQL Server]
dialect: mssql
data_source_name: "sqlserver://${DB_USER}:${DB_PASSWORD}@db.example:1433?database=app"
```

```yaml [ClickHouse]
dialect: clickhouse
data_source_name: "clickhouse://${DB_USER}:${DB_PASSWORD}@db.example:9000/app"
```

```yaml [TiDB]
dialect: tidb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:4000)/app?parseTime=true"
```

```yaml [Redshift]
dialect: redshift
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@cluster.example:5439/app?sslmode=require"
```

:::

Em URLs, codifique caracteres especiais do usuário/senha (`@`, `:`, `/`). A expansão de variáveis não faz isso. O formato `user:password@tcp(...)` de MySQL/MariaDB/TiDB não usa codificação URL. Ajuste TLS/SSL conforme o servidor.

### Níveis de log {#logging}

| Valor | Mensagens |
| --- | --- |
| `debug` | Depuração, informação, avisos, erros |
| `info` | Informação, avisos, erros |
| `warn` | Avisos e erros; aceita `warning` |
| `error` | Apenas erros |

```yaml
logger_level: debug # Exibir mensagens de depuração
```

## Variáveis de ambiente {#environment-variables}

`${DATABASE_DSN}` exige uma variável definida; `${DB_NAME:app}` usa `app` se ela faltar. Compartilhe o YAML entre ambientes e forneça credenciais por variáveis ou cofre de segredos. Host e usuário SSH e endereço do proxy também aceitam modelos.

## Ordem de busca {#configuration-file-search-order}

O primeiro arquivo existente no diretório de trabalho é usado:

1. `migration_config.yml`
2. `migration_config.yaml`
3. `migration_config.json`

Não há combinação de arquivos. Um arquivo selecionado inválido ou ilegível causa erro. `-c` usa somente o arquivo indicado, mesmo se estiver ausente.

```sh
migrate -c ./configs/dev.yml status
```

## Gerar por cenário

| Cenário | Comando |
| --- | --- |
| Um banco | `migrate new config` |
| Nome personalizado | `migrate new config dev.yml` |
| Vários bancos | `migrate new config --multi` |
| Quatro bancos | `migrate new config --multi --shards 4` |
| SSH Agent | `migrate new config --ssh-tunnel` |
| Chave privada | `migrate new config --ssh-tunnel --ssh-auth private_key` |
| Senha SSH | `migrate new config --ssh-tunnel --ssh-auth password` |
| SOCKS5 | `migrate new config --ssh-tunnel --proxy socks5` |
| HTTP CONNECT | `migrate new config --ssh-tunnel --proxy http` |

Combine as opções. Para sobrescrever use `--force`; veja todas em `migrate new config --help`.

```sh
migrate new config shards.yml --multi --shards 4 --ssh-tunnel --proxy socks5
```

## Vários bancos {#multiple-databases}

Use `data_sources` no lugar de `data_source_name`. Tipo, diretório e nome da tabela de histórico são comuns. Os nomes gerados começam em `shard_0`, com `SHARD_0_DSN` etc.

```yaml
dialect: postgres
migration_source: migrations
schema_name: migration_schema
data_sources:
  shard_0:
    data_source_name: "${SHARD_0_DSN}" # Conexão do primeiro banco
  shard_1:
    data_source_name: "${SHARD_1_DSN}" # Conexão do segundo banco
```

![Arquivos → shard_0 / shard_1 → Resultados](../assets/diagrams/multiple-databases-pt-BR.svg)

```sh
migrate create
migrate up --dry-run
migrate up
migrate status
```

Cada banco seleciona as versões pendentes pelo próprio histórico e as aplica em ordem. Sucessos anteriores não se repetem. `new version` cria um único arquivo compartilhado. O limite simultâneo padrão é 8; `--parallel 1` executa em sequência.

```sh
migrate up --parallel 4
migrate status --output json
```

Falhas não interrompem outros bancos. O resultado é resumido por conexão e qualquer falha gera saída não zero. Não é uma transação entre bancos: sucessos de outros bancos não são desfeitos. Ctrl+C impede novos trabalhos e cancela os ativos.

`ssh_tunnel` global é compartilhado; um objeto local o substitui inteiro, enquanto ausência ou `null` herda o global. Consulte [SSH](./ssh-tunnels). Cadastre cada banco real uma única vez, serialize comandos sobre ele e não altere arquivos ou dependências durante a execução. Inspecione dados e histórico antes de repetir após uma falha.

## Diretório de trabalho

```sh
migrate -w ./deploy up
migrate -w ./deploy -c configs/dev.yml status
```

`-w` determina a busca e substitui `working_directory`. `migration_source` relativo parte dele; apenas `-c` não muda essa base. Caminhos SSH relativos partem do diretório do arquivo de configuração.
