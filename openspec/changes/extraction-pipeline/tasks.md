## 1. Seed selection

- [ ] 1.1 `tools/seed`: page `registry.terraform.io/v1/providers`, sort by tier+downloads (do not trust `filter=`)
- [ ] 1.2 Union a fixed utility allowlist + `Telmate/proxmox`; cap at 50; write deterministic `seed.json`
- [ ] 1.3 Test: seed includes hyperscalers, proxmox, utilities; output is byte-stable

## 2. Provider extraction

- [ ] 2.1 `tools/extract`: resolve+download+SHA256-verify a binary (reuse nivis `internal/registry` model)
- [ ] 2.2 Invoke `nivis gen --provider <binary> --identity <id> --out <dir>`; capture `.nix` + normalized schema JSON
- [ ] 2.3 Resilient batch: skip+log a single failure, never hang
- [ ] 2.4 Test against the hermetic fake provider (built from nivis `cmd/provider-*`) — no network/creds

## 3. Compat tiers

- [ ] 3.1 `tools/compat`: compute schema-extractable, protocol, archs (Nivis ∩ published), e2e tier
- [ ] 3.2 Hand-maintained e2e allowlist; default tier `none` / `compatible by design`
- [ ] 3.3 Test: badge never claims `verified` without an allowlist entry; archs are an intersection

## 4. Proof set

- [ ] 4.1 Run the full pipeline on `hashicorp/random`, `hashicorp/null`, `Telmate/proxmox`
- [ ] 4.2 Confirm extraction output + compat records for all three
