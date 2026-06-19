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

- [x] 4.1 Run the full pipeline on `hashicorp/random`, `hashicorp/null`, `hashicorp/tls`, AND the
  credential-requiring providers `Telmate/proxmox`, `hashicorp/azurerm`, `hashicorp/google`. The
  latter three were blocked by `nivis gen` <= 0.4.1 (it called `ConfigureProvider` before fetching
  the schema, and credential-requiring providers ERROR there). FIXED upstream in **nivis 0.4.2**
  (`nivis gen` now uses a schema-only client that skips Configure — bean `nixform2-jcpm`); the
  `nivis` flake input is pinned to v0.4.2, so all six extract with no credentials.
- [x] 4.2 Confirm extraction output (`.nix` constructors + `metadata.json`) + compat records for all
  six providers (proxmox=7, azurerm=1130, google=1277 constructors).
