#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
hook_src="${repo_root}/.githooks/pre-commit"

if [[ ! -f "${hook_src}" ]]; then
  echo "hook file missing: ${hook_src}"
  exit 1
fi

tmp_root=$(mktemp -d)
trap 'rm -rf "${tmp_root}"' EXIT

work_repo="${tmp_root}/repo"
git init "${work_repo}" >/dev/null

cd "${work_repo}"
git config user.name "hook-test"
git config user.email "hook-test@example.com"
mkdir -p .githooks
cp "${hook_src}" .githooks/pre-commit
chmod +x .githooks/pre-commit
git config core.hooksPath .githooks

echo "init" > README.md
git add README.md
git commit -m "init" >/dev/null

# Should block Unix absolute path.
echo "path: /usr/local/bin" > abs-unix.txt
git add abs-unix.txt
if git commit -m "should-fail-unix-abs" >/dev/null 2>&1; then
  echo "expected unix absolute path commit to fail"
  exit 1
fi
git reset --hard -q HEAD

# Should block Windows absolute path.
echo "path: C:\\Users\\alice\\secret" > abs-win.txt
git add abs-win.txt
if git commit -m "should-fail-win-abs" >/dev/null 2>&1; then
  echo "expected windows absolute path commit to fail"
  exit 1
fi
git reset --hard -q HEAD

# Should allow URL lines (not filesystem absolute paths).
echo "docs: https://example.com/a/b" > url-ok.txt
git add url-ok.txt
if ! git commit -m "should-pass-url" >/dev/null 2>&1; then
  echo "expected URL-only commit to pass"
  exit 1
fi

echo "pre-commit hook smoke tests passed"
