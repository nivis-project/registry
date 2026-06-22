## Why

The forked registry-ui SPA consumes a JSON contract (the shape defined by OpenTofu's
`openapi.yml`). To reuse the SPA unchanged in structure, our backend must emit that same contract —
but populated with **schema-derived Nix constructors** instead of scraped HCL markdown. This change
builds the generator that turns extraction output into the static contract the SPA reads.

## What Changes

- **Docs contract emitter** (`tools/generate/`): walk the extraction output and emit the static
  contract tree `registry/docs/providers/{ns}/{name}/{version}/index.json` (listing resources/
  datasources/functions) plus per-item documents, matching registry-ui's contract types.
- **Nix rendering**: render each `nivis gen` constructor (its doc-comment header — computed outputs,
  nested-block shapes — and typed signature) into the per-item document body, so a resource page
  shows the Nivis Nix form, not HCL.
- Embed the **compat badge** (from `compat-tiers`) into each provider's `index.json`.

## Capabilities

### New Capabilities
- `docs-contract`: Emit the registry-ui-compatible static JSON contract tree from extraction output.
- `nix-rendering`: Render schema-derived Nix constructors into the contract's per-item documents.

### Modified Capabilities

## Impact

- New Go tool `tools/generate`.
- Output is static files (S3/CloudFront-ready); no live server required for v1.
- Depends on `extraction-pipeline`.
