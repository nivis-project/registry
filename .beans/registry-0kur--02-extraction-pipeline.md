---
# registry-0kur
title: 02 Extraction pipeline
status: completed
type: milestone
priority: normal
created_at: 2026-06-19T11:14:11Z
updated_at: 2026-06-19T12:03:26Z
---

Resolve+verify provider binaries, run nivis gen, capture schema, compute compat tiers. OpenSpec change: extraction-pipeline.

## Summary of Changes

- `tools/seed`: pages the public popularity API (filter= ignored), de-dups, sorts by (tier,
  -downloads), unions anchors + utility allowlist + proxmox, caps at 50, writes a byte-stable
  `seed.json`. `cmd/seed` generated the pinned 50-provider manifest.
- `tools/extract`: mirrors the nivis verify-before-execute model (resolve via OpenTofu registry →
  download → SHA256-verify → unpack), then runs `nivis gen` and captures `.nix` constructors +
  normalized `metadata.json` (version/protocols/platforms). Resilient `Batch` skips+logs single
  failures with a per-provider timeout. `cmd/extract` is the driver.
- `tools/compat`: computes the 4-axis record (schema-extractable, protocols, arches = Nivis ∩
  published, e2e) — never claims `verified` without an allowlist entry.
- Proof: random (10), null (1), tls (4), time (4) extract cleanly with compat records.
- FINDING: `nivis gen` v0.4.0 Configure()s before fetching the schema, so credential-requiring
  providers (proxmox, azurerm, google) ERROR. Fix is in flight upstream; re-pin the nivis input
  after its patch release. Proof set uses 3 credential-free real providers; proxmox/azure are a
  documented, soon-fixed gap (compat `schema_extractable:false`).
