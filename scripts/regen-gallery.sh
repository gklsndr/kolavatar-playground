#!/usr/bin/env bash
# Regenerates ts/public/gallery.json — the 100-descriptor catalogue the TS
# playground's Gallery panel reads from. Run after any SDK change that
# would affect descriptor output (a SpecVersion bump, a generator tweak,
# a new motif, etc.) so the gallery stays in sync with the live API.
#
# Usage:
#   bash scripts/regen-gallery.sh

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$ROOT/ts/public/gallery.json"

cd "$ROOT/go"
mkdir -p "$ROOT/ts/public"
go run -tags=kolavatardev ./cmd/kolavatar-gallery-json > "$OUT"

count=$(python3 -c "import json; print(len(json.load(open('$OUT'))))" 2>/dev/null \
        || grep -c '"descriptor"' "$OUT")
echo "regenerated $OUT ($count entries)"
