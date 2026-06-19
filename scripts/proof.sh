#!/usr/bin/env bash
# proof.sh — the end-to-end proof for the Nivis registry PoC.
#
# Runs the whole pipeline on the proof providers, generates the registry-ui
# contract into the frontend, and builds the static site. Intended to run inside
# `nix develop` (nivis, go, node, pnpm on PATH).
#
#   nix develop -c scripts/proof.sh
#
# The proof set now includes the credential-requiring providers (Telmate/proxmox,
# azurerm, google), which nivis >= 0.4.2 extracts: `nivis gen` fetches the schema
# WITHOUT calling ConfigureProvider (the nivis fix from bean nixform2-jcpm). The
# registry's `nivis` input is pinned to v0.4.2, so these need no credentials.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

NIVIS_BIN="${NIVIS_BIN:-$(command -v nivis)}"
EXTRACT_OUT="${EXTRACT_OUT:-extract-out}"
CONTRACT_DIR="frontend/public"

# The proof set: credential-free utilities + the hyperscalers + proxmox. The
# latter three exercise the nivis>=0.4.2 schema-without-configure path.
PROVIDERS=(random null tls Telmate/proxmox azurerm google)

echo "==> nivis: $NIVIS_BIN ($("$NIVIS_BIN" --version 2>/dev/null || echo '?'))"

echo "==> 1/3 extract: ${PROVIDERS[*]} (verify-before-execute -> nivis gen)"
( cd tools && go run ./cmd/extract -nivis "$NIVIS_BIN" -out "../$EXTRACT_OUT" "${PROVIDERS[@]}" )

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
