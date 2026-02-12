#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

VERSION="${1:-}"
if [[ -z "$VERSION" ]]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 v0.1.0"
  exit 1
fi

make tidy
make fmt
make cross-build

mkdir -p dist
for f in dist/roam-cli-*; do
  tar -czf "${f}-${VERSION}.tar.gz" -C dist "$(basename "$f")"
done

( cd dist && shasum -a 256 ./*.tar.gz > checksums.txt )

echo "Release artifacts created in dist/:"
ls -1 dist
