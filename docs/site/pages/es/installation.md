# Instalación

Instala con Go o descarga el binario de tu sistema y arquitectura desde [Releases](https://github.com/exc-works/migrate/releases). Verifica el SHA-256 y coloca `migrate` (`migrate.exe` en Windows) en PATH.

## Instalar desde módulo

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

Instala una versión específica:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

## Instalar desde código fuente local (repo privado o red interna)

Ejecuta en la raíz del repositorio:

```bash
go install ./cmd/migrate
```

## Verificar la instalación

```sh
migrate --help
migrate version
migrate capabilities --output json
```

Si no se encuentra el comando:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Las compilaciones locales muestran `dev`. Para instalaciones antiguas de Go que muestran `dev`, usa `go version -m <binary>`. `capabilities` no conecta a la base. Continúa con el [inicio rápido](./quick-start).
