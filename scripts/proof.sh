#!/usr/bin/env bash
# proof.sh — the end-to-end proof for the Nivis registry PoC.
#
# Runs the whole pipeline on the three credential-free real providers
# (random, null, tls), generates the registry-ui contract into the frontend,
# and builds the static site. Intended to run inside `nix develop` (nivis, go,
# node, pnpm on PATH).
#
#   nix develop -c scripts/proof.sh
#
# Telmate/proxmox is intentionally NOT in the proof set: nivis gen v0.4.0
# configures the provider before fetching its schema, and proxmox (like
# azurerm/google) rejects an all-null configure. A fix is in flight upstream;
# once a nivis patch release lands, add proxmox here and re-pin the input.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

NIVIS_BIN="${NIVIS_BIN:-$(command -v nivis)}"
EXTRACT_OUT="${EXTRACT_OUT:-extract-out}"
CONTRACT_DIR="frontend/public"

echo "==> nivis: $NIVIS_BIN ($("$NIVIS_BIN" --version 2>/dev/null || echo '?'))"

echo "==> 1/3 extract: random null tls (verify-before-execute -> nivis gen)"
( cd tools && go run ./cmd/extract -nivis "$NIVIS_BIN" -out "../$EXTRACT_OUT" random null tls )

echo "==> 2/3 generate: emit the registry-ui contract + Nix-rendered docs"
( cd tools && go run ./cmd/generate -extract "../$EXTRACT_OUT" -contract "../$CONTRACT_DIR" )

echo "==> 3/3 build the static site"
( cd frontend && pnpm install --frozen-lockfile && pnpm build )

echo
echo "==> PROOF ARTIFACTS"
echo "  contract:  $CONTRACT_DIR/registry/  ($(find "$CONTRACT_DIR/registry" -type f | wc -l) files)"
echo "  catalog:   $CONTRACT_DIR/registry/catalog.json"
echo "  site:      frontend/dist/  (static, contract bundled)"
echo "DONE — proof green."
