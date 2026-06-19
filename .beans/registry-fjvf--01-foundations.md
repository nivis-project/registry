---
# registry-fjvf
title: 01 Foundations
status: completed
type: milestone
priority: normal
created_at: 2026-06-19T11:14:11Z
updated_at: 2026-06-19T11:42:21Z
---

Reproducible Nix-flake dev env, jj VC, beans+OpenSpec, CI/test harness. OpenSpec change: foundations-scaffold.

## Summary of Changes

- Verified jj (colocated with git), remote `origin = git@github.com:nivis-project/registry.git`,
  author `Pim Snel <post@pimsnel.com>`. `.gitignore` covers Nix/Go/Node/cache outputs.
- `flake.nix`: plain-nix `forAllSystems`/`listToAttrs` (no flake-utils), four systems, `nivis`
  input exposed in the devShell; devShell carries go/nodejs/pnpm/jj + nivis CLI.
- `packages.default` builds `tools/` via `buildGoModule` (vendorHash = null; no external deps).
- `nix flake check` now runs `go test ./...` hermetically (`checks.go-tests`) plus a gofmt gate
  (`checks.gofmt`). `.github/workflows/ci.yml` runs `nix flake check` + `go test ./...`.
- Confirmed beans board (`01`–`06` + epics) and `openspec validate --all` pass.
- Verified the linchpin dependency: built the nivis CLI (v0.4.0) and ran `nivis gen` on a fake
  provider; captured the exact `.nix` constructor output format for the renderer to consume.
