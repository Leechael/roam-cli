#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  scripts/init-prek.sh [--force] <target-repo-dir>
USAGE
}

force="false"
if [[ "${1:-}" == "--force" ]]; then
  force="true"
  shift
fi

target="${1:-}"
if [[ -z "$target" ]]; then
  usage
  exit 1
fi

if [[ ! -d "$target" ]]; then
  echo "target repo dir does not exist: $target" >&2
  exit 1
fi

dest="$target/prek.toml"
if [[ -f "$dest" && "$force" != "true" ]]; then
  echo "prek.toml already exists: $dest" >&2
  echo "Use --force to overwrite." >&2
  exit 1
fi

cat > "$dest" <<'EOF'
minimum_prek_version = "0.0.0"

[[repos]]
repo = "https://github.com/pre-commit/pre-commit-hooks"
rev = "v4.6.0"
hooks = [
  { id = "check-merge-conflict" },
  { id = "check-yaml" },
  { id = "end-of-file-fixer" },
  { id = "trailing-whitespace" },
]

[[repos]]
repo = "local"
hooks = [
  { id = "gofmt-check", name = "gofmt check", entry = "bash -c 'files=$(gofmt -l ./cmd ./internal ./tests); if [ -n \"$files\" ]; then echo \"Unformatted files:\"; echo \"$files\"; exit 1; fi'", language = "system", pass_filenames = false },
  { id = "go-vet", name = "go vet", entry = "go vet ./...", language = "system", pass_filenames = false },
  { id = "go-unit-test", name = "go unit test", entry = "go test ./... -count=1", language = "system", pass_filenames = false },
]
EOF

echo "Created $dest"
echo "Next: cd $target && prek validate-config && prek install --install-hooks && prek run --all-files"
