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

source release-naming.env

make tidy
make fmt
make cross-build

mkdir -p dist
for f in "dist/${BINARY_NAME}-"*; do
  tmpdir=$(mktemp -d)
  cp "$f" "$tmpdir/$BINARY_NAME"
  tar -czf "${f}-${VERSION}.tar.gz" -C "$tmpdir" "$BINARY_NAME"
  rm -rf "$tmpdir"
done

( cd dist && shasum -a 256 ./*.tar.gz > checksums.txt )

echo "Release artifacts created in dist/:"
ls -1 dist
