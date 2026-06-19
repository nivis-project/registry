---
# registry-w9e6
title: 03 Contract and generator
status: completed
type: milestone
priority: normal
created_at: 2026-06-19T11:14:11Z
updated_at: 2026-06-19T12:11:35Z
---

Emit registry-ui-compatible static JSON contract + render Nix constructors. OpenSpec change: contract-generator.

## Summary of Changes

- `tools/generate/nixparse.go`: parses nivis gen .nix constructors into a
  Constructor model (provider, type, required/optional args, computed outputs)
  by reading the header comment, the lambda arg set, and the throw-guards.
- `tools/generate/render.go`: renders a Constructor as a Markdown doc body in
  Nix form — signature, required/optional args, computed outputs (via refAttr),
  and a Nix usage example. Emits NO HCL `resource {}` block.
- `tools/generate/generate.go`: emits the registry-ui-compatible contract
  (ProviderVersion shape: docs.{resources,datasources,functions,guides}) at
  registry/docs/providers/{ns}/{name}/{version}/index.json, with the compat
  record embedded, plus per-item <type>.md documents. GenerateAll walks the
  extraction tree resiliently.
- `cmd/generate` driver; real run produced contracts for random/null/tls/time
  with rendered Nix docs and compat badges.
- Contract types mirror opentofu/registry-ui openapi.yml (verified against
  upstream) so the forked SPA's contract types are reusable as-is.
