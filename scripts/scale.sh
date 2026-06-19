#!/usr/bin/env bash
# scale.sh — run the full pipeline over ALL providers in seed.json.
#
# Extracts every seed provider (resiliently — a single failure is skipped and
# logged, never aborting), generates the full static contract, writes an
# auditable coverage report, and (unless SKIP_BUILD=1) builds the static site.
# Intended to run inside `nix develop` (nivis, go, node, pnpm on PATH).
#
#   nix develop -c scripts/scale.sh
#
# This is milestone-06 work: it scales extraction + contract generation ONLY.
# It does NOT deploy to S3/CloudFront (a separate, human-approved step) and does
# NOT change the seed.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

NIVIS_BIN="${NIVIS_BIN:-$(command -v nivis)}"
SEED="${SEED:-seed.json}"
EXTRACT_OUT="${EXTRACT_OUT:-extract-out}"
CONTRACT_DIR="frontend/public"
REPORT="${REPORT:-coverage.json}"

echo "==> nivis: $NIVIS_BIN ($("$NIVIS_BIN" --version 2>/dev/null || echo '?'))"
echo "==> seed:  $SEED ($(python3 -c "import json;print(json.load(open('$SEED'))['count'])" 2>/dev/null || echo '?') providers)"

echo "==> 1/3 extract: every provider in $SEED (verify-before-execute -> nivis gen)"
( cd tools && go run ./cmd/extract \
    -nivis "$NIVIS_BIN" -seed "../$SEED" -out "../$EXTRACT_OUT" -report "../$REPORT" )

echo "==> 2/3 generate: emit the full registry-ui contract + Nix-rendered docs"
( cd tools && go run ./cmd/generate -extract "../$EXTRACT_OUT" -contract "../$CONTRACT_DIR" )

if [ "${SKIP_BUILD:-0}" != "1" ]; then
  echo "==> 3/3 build the static site"
  ( cd frontend && pnpm install --frozen-lockfile && pnpm build )
else
  echo "==> 3/3 build skipped (SKIP_BUILD=1)"
fi

echo
echo "==> COVERAGE"
python3 - "$REPORT" <<'PY'
import json, sys
r = json.load(open(sys.argv[1]))
print(f"  extracted: {r['ok']}/{r['total']}   skipped: {r['skipped']}")
sk = [p for p in r["providers"] if not p["extracted"]]
if sk:
    print("  skipped providers:")
    for p in sk:
        print(f"    - {p['address']}: {p.get('reason','?')[:140]}")
PY

echo
echo "==> ARTIFACTS"
echo "  contract: $CONTRACT_DIR/registry/  ($(find "$CONTRACT_DIR/registry" -type f 2>/dev/null | wc -l) files)"
echo "  report:   $REPORT"
[ "${SKIP_BUILD:-0}" != "1" ] && echo "  site:     frontend/dist/  (static, contract bundled)"
echo "DONE — scale run complete (no deploy)."
