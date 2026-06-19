## Why

The registry's unique value is schema-derived Nix documentation. That requires a pipeline that,
for each selected provider, resolves and verifies its binary, runs `nivis gen` to capture the
schema-derived constructors, and computes a compatibility tier. This change builds that pipeline
and proves it on a small, fast set of providers (never the giant AWS/Azure during the proof).

## What Changes

- **Seed selection** (`tools/seed/`): derive the provider list from the public popularity API
  (`registry.terraform.io/v1/providers`, paging + sorting by `downloads`/`tier` since `filter=` is
  inert), union a fixed utility allowlist and **Telmate/proxmox**, and pin the result to `seed.json`.
- **Provider extraction** (`tools/extract/`): reuse nivis's `internal/registry` resolution model
  (address → download → **SHA256 verify** → cache) to obtain a binary, then invoke
  `nivis gen --provider <binary> --identity <id> --out <dir>` and capture the generated `.nix`
  constructors plus a normalized schema JSON.
- **Compat tiers** (`tools/compat/`): compute a per-provider badge from
  (a) schema-extractable yes/no, (b) protocol 5.0/6.0 from metadata, (c) architectures =
  Nivis-supported ∩ published `targets`, (d) e2e tier `none|verified` from a tiny hand-maintained allowlist.
- Prove the pipeline on **3 small real providers** (`hashicorp/random`, `hashicorp/null`, `Telmate/proxmox`)
  and **1 hermetic fake provider** built from nivis's `cmd/provider-*`.

## Capabilities

### New Capabilities
- `seed-selection`: Derive and pin the provider seed manifest from the public popularity signal plus a utility allowlist and proxmox.
- `provider-extraction`: Resolve+verify a provider binary and run `nivis gen` to capture schema-derived Nix constructors and normalized schema JSON.
- `compat-tiers`: Compute a per-provider compatibility tier/badge from metadata + schema + an e2e allowlist.

### Modified Capabilities

## Impact

- New Go tools under `tools/seed`, `tools/extract`, `tools/compat`.
- Executes real provider binaries (sandboxed). A failing single extraction MUST be skipped and logged, never hang the run.
- Depends on `foundations-scaffold`.
