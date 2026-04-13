SHELL := /bin/bash

GO ?= go
APP_BIN ?= sql-migrate
APP_PKG ?= ./cmd/sql-migrate

export GOCACHE ?= $(CURDIR)/.cache/go-build
export GOMODCACHE ?= $(CURDIR)/.cache/go-mod

.PHONY: help prepare-cache hooks fmt tidy test test-integration vet build build-cli install ci-check

help:
	@echo "Available targets:"
	@echo "  make hooks             # install git hooks (.githooks)"
	@echo "  make fmt               # gofmt all Go files"
	@echo "  make tidy              # go mod tidy"
	@echo "  make test              # run unit tests"
	@echo "  make test-integration  # run integration tests"
	@echo "  make vet               # run go vet"
	@echo "  make build             # build all packages"
	@echo "  make build-cli         # build CLI binary"
	@echo "  make install           # go install CLI"
	@echo "  make ci-check          # test + integration + vet + build"

prepare-cache:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"

hooks:
	@./scripts/install-git-hooks.sh

fmt:
	@$(GO)fmt ./...
	@gofmt -w $$(find cmd internal integrationtest -name '*.go' | sort)

tidy: prepare-cache
	@$(GO) mod tidy

test: prepare-cache
	@$(GO) test ./...

test-integration: prepare-cache
	@$(GO) test -tags=integration ./integrationtest/...

vet: prepare-cache
	@$(GO) vet ./...

build: prepare-cache
	@$(GO) build ./...

build-cli: prepare-cache
	@$(GO) build $(APP_PKG)

install: prepare-cache
	@$(GO) install $(APP_PKG)

ci-check: test test-integration vet build build-cli
