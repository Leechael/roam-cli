#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  scripts/init-release-naming.sh [--force] <target-repo-dir>

Examples:
  scripts/init-release-naming.sh /path/to/repo
  scripts/init-release-naming.sh --force /path/to/repo
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

dest="$target/release-naming.env"

if [[ -f "$dest" && "$force" != "true" ]]; then
  echo "release-naming.env already exists: $dest" >&2
  echo "Use --force to overwrite." >&2
  exit 1
fi

cat > "$dest" <<'EOF'
# Single source of truth for release naming used by docs/scripts/workflows.
CLI_NAME=roam-cli
BINARY_NAME=roam-cli
TAG_PREFIX=v
ARTIFACT_GLOB=roam-cli-*.tar.gz
EOF

echo "Created $dest"
