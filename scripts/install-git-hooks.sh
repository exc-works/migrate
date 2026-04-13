#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)

if ! command -v git >/dev/null 2>&1; then
  echo "git command not found"
  exit 1
fi

if [[ ! -f "${repo_root}/.githooks/pre-commit" ]]; then
  echo "missing hook file: ${repo_root}/.githooks/pre-commit"
  exit 1
fi

chmod +x "${repo_root}/.githooks/pre-commit"
git -C "${repo_root}" config core.hooksPath .githooks

echo "Installed git hooks: core.hooksPath=.githooks"
echo "pre-commit scans for private keys, absolute paths, and hardcoded credentials."
echo "Also enable protected-branch push protection on the remote repository."
