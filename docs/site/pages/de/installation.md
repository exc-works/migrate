# Installation

Installieren Sie über Go oder laden Sie das passende Betriebssystem-/Architekturpaket aus den [Releases](https://github.com/exc-works/migrate/releases). Prüfen Sie SHA-256 und legen Sie `migrate` beziehungsweise `migrate.exe` unter Windows in PATH ab.

## Aus Modul installieren

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

Eine bestimmte Version installieren:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

## Aus lokalem Quellcode installieren (privates Repo oder internes Netzwerk)

Im Repository-Root ausführen:

```bash
go install ./cmd/migrate
```

## Installation prüfen

```sh
migrate --help
migrate version
migrate capabilities --output json
```

Wenn der Befehl nicht gefunden wird:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Lokale Builds melden `dev`. Bei älteren Go-Installationen mit `dev` hilft `go version -m <binary>`. `capabilities` benötigt keine Verbindung. Weiter zum [Schnellstart](./quick-start).
