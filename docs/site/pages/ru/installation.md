# Установка

Используйте Go или загрузите подходящий ОС и архитектуре бинарный файл из [Releases](https://github.com/exc-works/migrate/releases). Сверьте SHA-256 и поместите `migrate` (`migrate.exe` в Windows) в PATH.

## Установка из модуля

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

Установка конкретной версии:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

## Установка из локального исходного кода (приватный репозиторий или внутренняя сеть)

Выполните в корне репозитория:

```bash
go install ./cmd/migrate
```

## Проверка установки

```sh
migrate --help
migrate version
migrate capabilities --output json
```

Если команда не найдена:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Локальные сборки показывают `dev`. Для старых установок Go с `dev` используйте `go version -m <binary>`. capabilities не подключается к базе. Далее — [быстрый старт](./quick-start).
