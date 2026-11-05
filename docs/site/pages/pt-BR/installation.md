# Instalação

Use Go ou baixe o binário do seu sistema e arquitetura em [Releases](https://github.com/exc-works/migrate/releases). Confira o SHA-256 e coloque `migrate` (`migrate.exe` no Windows) no PATH.

## Instalar a partir do módulo

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

Instale uma versão específica:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

## Instalar a partir do código-fonte local (repositório privado ou rede interna)

Execute na raiz do repositório:

```bash
go install ./cmd/migrate
```

## Verificar instalação

```sh
migrate --help
migrate version
migrate capabilities --output json
```

Se o comando não for encontrado:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Builds locais mostram `dev`; instalações Go antigas podem ser identificadas com `go version -m <binary>`. `capabilities` não conecta ao banco. Siga para o [início rápido](./quick-start).
