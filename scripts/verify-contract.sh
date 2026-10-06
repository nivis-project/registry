#!/usr/bin/env bash
# verify-contract.sh — refuse a contract that is absent or incomplete.
#
# This gate is why the contract is bundled into the build artifact rather than
# proxied from a second origin: a bad contract must fail the build and leave the
# previous deploy serving, instead of publishing a site that 404s its own data
# at runtime.
#
#   scripts/verify-contract.sh [contract-dir]     (default frontend/public/registry)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEST="${1:-${CONTRACT_DEST:-$ROOT/frontend/public/registry}}"

die() { printf 'verify-contract: %s\n' "$*" >&2; exit 1; }

[ -d "$DEST" ] || die "no contract directory at $DEST"
[ -f "$DEST/catalog.json" ] || die "no catalog.json under $DEST; the contract is absent"

python3 - "$DEST" <<'PY'
import json, pathlib, sys

dest = pathlib.Path(sys.argv[1])

def load(path):
    try:
        return json.loads(path.read_text())
    except json.JSONDecodeError as e:
        sys.exit(f"verify-contract: {path.name} is not valid JSON: {e}")

catalog = load(dest / "catalog.json")
if not catalog:
    sys.exit("verify-contract: catalog.json lists no providers; refusing an empty site")

missing = [
    f"{p['namespace']}/{p['name']}@{p['version']}"
    for p in catalog
    if not (dest / "docs" / "providers" / p["namespace"] / p["name"] / p["version"] / "index.json").is_file()
]
if missing:
    sys.exit(
        f"verify-contract: {len(missing)} catalogued provider(s) have no index document, "
        f"e.g. {', '.join(missing[:5])}"
    )

# modules.json is optional: a contract generated before the module pipeline has
# none. When present, every row must resolve on the same terms.
mod_count = 0
modules_file = dest / "modules.json"
if modules_file.is_file():
    modules = load(modules_file)
    mod_missing = [
        f"{m['owner']}/{m['name']}@{m['version']}"
        for m in modules
        if not (dest / "docs" / "modules" / m["owner"] / m["name"] / m["version"] / "index.json").is_file()
    ]
    if mod_missing:
        sys.exit(f"verify-contract: catalogued module(s) have no index document: {', '.join(mod_missing)}")
    mod_count = len(modules)

print(f"verify-contract: ok, {len(catalog)} provider(s), {mod_count} module(s), all documents present")
PY
