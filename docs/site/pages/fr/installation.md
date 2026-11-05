# Installation

Installez avec Go ou choisissez un binaire pour votre système et architecture dans les [releases](https://github.com/exc-works/migrate/releases). Vérifiez le SHA-256 publié et placez `migrate` (`migrate.exe` sous Windows) sur PATH.

## Installer depuis le module

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

Installer une version spécifique :

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

## Installer depuis le code source local (repo privé ou réseau interne)

Exécutez dans la racine du dépôt :

```bash
go install ./cmd/migrate
```

## Vérifier l’installation

```sh
migrate --help
migrate version
migrate capabilities --output json
```

Si la commande est introuvable :

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Les compilations locales affichent `dev`. Pour une ancienne installation Go affichant `dev`, utilisez `go version -m <binary>`. `capabilities` ne se connecte pas à la base. Continuez avec le [démarrage rapide](./quick-start).
