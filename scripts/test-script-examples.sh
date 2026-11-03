#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
example_run_dir=$(mktemp -d)
trap 'rm -rf -- "$example_run_dir"' EXIT
cd "$repo_root"

command -v npm >/dev/null
go build -tags=full -o "$example_run_dir/migrate" ./cmd/migrate

# Install only the committed local package, keeping all generated files out of the checkout.
mkdir "$example_run_dir/npm"
for item in package.json package-lock.json packages migrations migration_config.json; do
  cp -R "examples/script-imports/npm/$item" "$example_run_dir/npm/$item"
done
npm --prefix "$example_run_dir/npm" ci --ignore-scripts --no-audit --no-fund --offline \
  --cache "$example_run_dir/npm-cache"

for example_name in local npm; do
  example_source="$repo_root/examples/script-imports/local"
  if [[ "$example_name" == npm ]]; then
    example_source="$example_run_dir/npm"
  fi
  export MIGRATE_EXAMPLE_DB="$example_run_dir/$example_name.db"
  "$example_run_dir/migrate" --working-dir "$example_source" create
  "$example_run_dir/migrate" --working-dir "$example_source" up --dry-run
  "$example_run_dir/migrate" --working-dir "$example_source" up
  "$example_run_dir/migrate" --working-dir "$example_source" status
  if [[ "$example_name" == local ]]; then
    "$example_run_dir/migrate" --working-dir "$example_source" down 2
    "$example_run_dir/migrate" --working-dir "$example_source" status
    "$example_run_dir/migrate" --working-dir "$example_source" up
  fi
  "$example_run_dir/migrate" --working-dir "$example_source" down --all
  "$example_run_dir/migrate" --working-dir "$example_source" status
  printf '%s example passed\n' "$example_name"
done

embedded_output=$(go run -tags=full ./examples/script-imports/embedded)
expected_output=$'after up: alice\nafter down: all migrations pending'
if [[ "$embedded_output" != "$expected_output" ]]; then
  printf '%s\n%s\n' 'Unexpected embedded example output' "$embedded_output" >&2
  exit 1
fi
printf '%s\n' "$embedded_output"
go test -tags=full . -run '^ExampleWithModules$'
printf 'All script import examples passed\n'
