#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

if ! command -v prek >/dev/null 2>&1; then
  echo "Error: 'prek' is not installed. See https://prek.j178.dev/quickstart/" >&2
  exit 1
fi

chmod +x .githooks/pre-commit

git config core.hooksPath .githooks

echo "Git hooks path set to .githooks"
echo "Running initial hook checks..."
prek run --config prek.toml --all-files

echo "Done. pre-commit hook is active."
