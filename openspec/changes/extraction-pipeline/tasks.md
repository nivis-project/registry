## 1. Seed selection

- [x] 1.1 `tools/seed`: page `registry.terraform.io/v1/providers`, sort by tier+downloads (do not trust `filter=`)
- [x] 1.2 Union a fixed utility allowlist + `Telmate/proxmox`; cap at 50; write deterministic `seed.json`
- [x] 1.3 Test: seed includes hyperscalers, proxmox, utilities; output is byte-stable

## 2. Provider extraction

- [x] 2.1 `tools/extract`: resolve+download+SHA256-verify a binary (reuse nivis `internal/registry` model)
- [x] 2.2 Invoke `nivis gen --provider <binary> --identity <id> --out <dir>`; capture `.nix` + normalized schema JSON
- [x] 2.3 Resilient batch: skip+log a single failure, never hang
- [x] 2.4 Test against the hermetic fake provider (built from nivis `cmd/provider-*`) — no network/creds

## 3. Compat tiers

- [x] 3.1 `tools/compat`: compute schema-extractable, protocol, archs (Nivis ∩ published), e2e tier
- [x] 3.2 Hand-maintained e2e allowlist; default tier `none` / `compatible by design`
- [x] 3.3 Test: badge never claims `verified` without an allowlist entry; archs are an intersection

## 4. Proof set

- [x] 4.1 Run the full pipeline on `hashicorp/random`, `hashicorp/null`, and a third credential-free
  real provider (`hashicorp/tls`). NOTE: `Telmate/proxmox` (and `azurerm`/`google`) cannot be
  schema-extracted by `nivis gen` v0.4.0 — it calls `ConfigureProvider` with an all-null config
  before fetching the schema, and credential-requiring providers ERROR there. The pipeline handles
  this resiliently (skip+log). A fix is in flight upstream in the nivis repo; once a nivis patch
  release lands, re-pinning the `nivis` flake input makes proxmox/azure extract with no code change.
  See memory `nivis-gen-configure-limitation`.
- [x] 4.2 Confirm extraction output (`.nix` constructors + `metadata.json`) + compat records for all
  three green providers; record proxmox/azure as a documented, soon-fixed gap (compat axis 1
  `schema_extractable:false`).
