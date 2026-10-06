#!/usr/bin/env bash
# fetch-contract.sh — pull the generated contract into the frontend before the
# site build, then REFUSE to continue if it is missing or incomplete.
#
# The verification is the point. The contract is bundled rather than proxied
# from a second origin precisely so that a bad contract fails the build and
# leaves the previous deploy serving, instead of publishing a site that 404s its
# own data at runtime. Without this gate, bundling keeps the silent failure it
# was chosen to remove.
#
#   CONTRACT_BUCKET=s3://bucket/prefix scripts/fetch-contract.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

DEST="${CONTRACT_DEST:-frontend/public/registry}"

die() { printf '\nfetch-contract: %s\n' "$*" >&2; exit 1; }

[ -n "${CONTRACT_BUCKET:-}" ] || die "CONTRACT_BUCKET is not set; the build cannot obtain a contract"

echo "==> syncing contract from $CONTRACT_BUCKET"
mkdir -p "$DEST"
aws s3 sync "$CONTRACT_BUCKET" "$DEST" --delete --only-show-errors \
  || die "s3 sync failed"

# --- the gate ---------------------------------------------------------------
# Kept in its own script so it can be exercised without AWS.

echo "==> verifying the contract is complete"
"$ROOT/scripts/verify-contract.sh" "$DEST"

echo "==> contract ready in $DEST"
