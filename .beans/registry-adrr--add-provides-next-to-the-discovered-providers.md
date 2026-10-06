---
# registry-adrr
title: add providers next to the discovered providers
status: completed
type: task
priority: normal
created_at: 2026-10-06T13:51:13Z
updated_at: 2026-10-06T16:03:40Z
---

e.g. https://github.com/nivis-project/terraform-provider-nivis-tunnel and https://github.com/nivis-project/terraform-provider-hcloudimage


## OpenSpec

`seed-pins-data-file` (store: registry). Moves the three hardcoded pin lists in
`tools/seed/seed.go` to `seed-pins.json`, adds `nivis-project/hcloudimage` with
reason `curated`, and derives the seed cap so a pin can never evict a
popularity-ranked provider.

## Findings (explored 2026-10-06)

The two example providers are not the same case:

| Provider | OpenTofu registry | Status |
|------------------------------|-------------------|-----------------------------|
| `nivis-project/hcloudimage`  | 200, v0.1.0       | pinnable today              |
| `nivis-project/nivis-tunnel` | 404               | not released yet            |

Entry condition for a curated pin is that the address resolves on
`registry.opentofu.org`. The extraction pipeline needs no change: `extract.Batch`
already resolves against OpenTofu and skips a 404 with a reason instead of
failing the run.

`terraform-provider-nivis-tunnel` has no tags, no releases, and no
`.github/workflows`. To register it, copy from `terraform-provider-hcloudimage`,
which already has the full setup:

- `.goreleaser.yml` (with the `signs:` block for the detached SHA256SUMS signature)
- `.github/workflows/release.yml`
- `terraform-registry-manifest.json`
- `GPG_FINGERPRINT` repo secret
- then tag `v0.1.0`, then PR to `opentofu/registry`

Check whether the existing `nivis-project` GPG key (`FC80F1F128669C1B`) can be
reused rather than minting a second one.

## Parked: hosting the registry protocol ourselves

Nivis already accepts `host/ns/name` and goes straight to
`https://{host}/v1/providers/...` with no service-discovery step
(`nivis/internal/registry/resolve.go`). So serving the protocol ourselves is
static JSON plus a GitHub release indexer, and we would never host binaries.

Decided against for now. Nivis verifies SHA256 only and has no GPG support, so
"submit here instead of OpenTofu" would currently mean submitting to the weaker
supply chain. Our differentiation is the Nix docs, and those work fine on top of
OpenTofu. Revisit when one of these is true:

1. A provider OpenTofu will not or cannot host (private, customer-internal).
2. We want independence from the OpenTofu registry as a strategic position.
3. Measured submission friction at OpenTofu is turning authors away.


## Summary of Changes

Shipped as OpenSpec change `seed-pins-data-file` (archived
`2026-10-06-seed-pins-data-file`), commit `0a644709`.

- `seed-pins.json` at the repo root replaces `MustInclude`, `UtilityAllowlist`
  and `EuropeanAllowlist` in `tools/seed/seed.go`. 21 entries, each carrying its
  own `address` / `reason` / `note`, file order is selection order.
- `nivis-project/hcloudimage` added with reason `curated`. It extracts: 1
  constructor at 0.1.0, protocol 5.0, all four Nivis-supported architectures.
- The `if a == "Telmate/proxmox"` reason override is gone. Proxmox is a plain
  `anchor` and its annotation lives in the pin's `note`.
- `MaxSeed` became `PopularFill = 38` with the cap derived as
  `len(pins) + PopularFill`, so pinning grows the manifest instead of evicting
  the lowest-ranked popular provider. `seed.json` is 58 -> 59 entries with
  nothing dropped.
- `flake.nix` now composes `tools/` plus `seed-pins.json` for the `go-tests`
  check and sets `modRoot = "tools"`. Without it the new tests could not reach
  the repo-root pin file from the Nix sandbox.

Still open, tracked in the notes above: `terraform-provider-nivis-tunnel` has no
release yet, so it is not pinnable. Registry protocol hosting stays parked.
