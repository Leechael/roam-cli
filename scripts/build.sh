#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

make tidy
make fmt
make build

echo "Built: $ROOT_DIR/roam-cli"
