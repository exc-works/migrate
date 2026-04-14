# AGENTS Guide

This file is a minimal entrypoint for AI agents.

## Source of Truth

- Start from `docs/migrate-harness-spec.md`.
- Detailed requirements live only under `docs/spec/`.
- Do not duplicate or reinterpret spec details in this file.

## Hard Security Rules

- Never commit absolute local paths.
- Never commit plaintext passwords, tokens, or real credentials.
- Never commit private keys or key material.

## Merge Gates

1. `go test ./...`
2. `go test -tags=integration ./...`
3. `go vet ./...`
4. `go build ./...`
5. `go build ./cmd/migrate`
