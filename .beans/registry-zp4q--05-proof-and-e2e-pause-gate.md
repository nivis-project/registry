---
# registry-zp4q
title: 05 Proof and e2e (PAUSE gate)
status: completed
type: milestone
priority: normal
created_at: 2026-06-19T11:14:11Z
updated_at: 2026-06-19T12:29:09Z
---

Hermetic fake-provider + 3 real-provider e2e; commit+push main; STOP for review. OpenSpec change: proof-e2e.

## Summary of Changes

- Hermetic e2e (`tools/e2e` `TestHermeticPipeline`): builds the nivis fake provider (beta) and runs
  it through extract → contract → render OFFLINE; asserts a Nix page exists, no HCL block, declares
  the required `from` argument. Wired into CI as the `e2e-hermetic` job (nivis CLI + matching source
  from the pinned input; the pipeline uses no provider network/creds).
- Real-provider e2e (`TestRealProviderPipeline`): random, null, tls through the full pipeline.
  Telmate/proxmox excluded — nivis gen v0.4.0 can't configure credential-requiring providers; fix in
  flight upstream (re-pin the nivis input after its patch release).
- `scripts/proof.sh`: end-to-end proof (extract 3 real providers → contract → static site build),
  proven green locally. Served dist/ over HTTP — every SPA data URL resolves; docs are Nix, no HCL.
- Gate: all checks green (`nix flake check`, `go test ./...`, both e2e, frontend test+build).
  Committed as Pim Snel (no Claude trailer) and pushed to `main`. STOPPED — did NOT scale to 50 or
  deploy; milestone 06 stays todo/draft.
