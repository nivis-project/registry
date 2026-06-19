---
# registry-bvar
title: Scale extraction to all 50 seed providers (no deploy)
status: completed
type: task
priority: normal
created_at: 2026-06-19T15:58:20Z
updated_at: 2026-06-19T16:12:35Z
parent: registry-pao9
---


## Summary of Changes

- `scripts/scale.sh`: full-catalog run — extract every provider in seed.json (resilient batch) →
  generate the full contract → build the static site → print a coverage summary. No deploy.
- `tools/cmd/extract`: added `-report <path>` writing a JSON coverage report (per-provider
  extracted/constructor-count vs. skipped/reason + totals).
- Fixed two latent bugs surfaced by scaling:
  - **datasource-only providers** (hashicorp/http, hashicorp/external) no longer count as failures:
    `RunGen` returns an empty slice (not an error) when nivis gen emits no resources; the generator
    treats a successfully-extracted provider as schema_extractable even with 0 constructors.
  - **prerelease version selection**: `latestVersionIndex` now prefers the latest STABLE release over
    a higher `-rc`/`-beta` (falling back to a prerelease only when no stable exists). This fixed
    signalfx (9.30.2 vs broken 10.0.0-rc1), proxmox (2.9.14), and rancher2 (14.1.1).
- Result: **50/50 providers extracted, 0 skipped**, 10,600 contract files, 50-entry catalog, site
  builds. Nix-validity verified at scale: 0 broken across the per-provider sample + all 3954
  cfg_-aliased files. go test ./... + nix flake check green.
- NO deploy, NO seed change (per the scope guard).
