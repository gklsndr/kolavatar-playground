#!/usr/bin/env bash
# Stages copies of the two sibling repos into the playground so the Docker
# build context is self-contained:
#
#   ts/vendor-kolavatar-ts/src/   ← copy of kolavatar-ts/src (Vite alias target)
#   vendor-kolavatar-go/          ← copy of kolavatar-go    (Go replace target)
#
# Both paths are .gitignored. Run before `fly deploy`. Override the source
# locations with $KOLAVATAR_TS / $KOLAVATAR_GO if the siblings aren't at
# the default paths.
#
# Usage:
#   bash scripts/prepare-deploy.sh
#   fly deploy

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
KOLAVATAR_TS="${KOLAVATAR_TS:-$ROOT/../kolavatar-ts}"
KOLAVATAR_GO="${KOLAVATAR_GO:-$ROOT/../kolavatar-go}"

for path in "$KOLAVATAR_TS" "$KOLAVATAR_GO"; do
  if [ ! -d "$path" ]; then
    echo "error: required source not found at $path" >&2
    echo "       set KOLAVATAR_TS / KOLAVATAR_GO env vars to override." >&2
    exit 1
  fi
done

# kolavatar-ts: only the src is consumed by the Vite alias. node_modules,
# dist, tests, and package metadata aren't needed for the bundle.
TS_DEST="$ROOT/ts/vendor-kolavatar-ts"
rm -rf "$TS_DEST"
mkdir -p "$TS_DEST"
cp -r "$KOLAVATAR_TS/src" "$TS_DEST/src"
[ -f "$KOLAVATAR_TS/LICENSE" ] && cp "$KOLAVATAR_TS/LICENSE" "$TS_DEST/LICENSE"

# kolavatar-go: the whole module (the Go replace directive points at the
# whole tree, not just one subdir). Exclude .git and node-style cruft.
GO_DEST="$ROOT/vendor-kolavatar-go"
rm -rf "$GO_DEST"
mkdir -p "$GO_DEST"
( cd "$KOLAVATAR_GO" && find . \
    -path './.git' -prune -o \
    -path './node_modules' -prune -o \
    -path './bin' -prune -o \
    -type f -print ) | while read -r rel; do
  dest="$GO_DEST/${rel#./}"
  mkdir -p "$(dirname "$dest")"
  cp "$KOLAVATAR_GO/${rel#./}" "$dest"
done

ts_count=$(find "$TS_DEST/src" -name '*.ts' | wc -l | tr -d ' ')
go_count=$(find "$GO_DEST" -name '*.go' | wc -l | tr -d ' ')
echo "vendored kolavatar-ts/src → ts/vendor-kolavatar-ts/src ($ts_count files)"
echo "vendored kolavatar-go     → vendor-kolavatar-go        ($go_count files)"
