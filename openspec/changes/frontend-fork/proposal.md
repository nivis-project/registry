## Why

Building a registry frontend from scratch would reinvent OpenTofu's proven, data-agnostic SPA.
Instead we fork `opentofu/registry-ui`, keep its routing/search/contract-types, and adapt only the
resource-page rendering to display Nix constructors. This maximizes reuse and keeps us on the same
architecture the eventual API phase will use.

## What Changes

- Vendor/fork `opentofu/registry-ui` under `frontend/` (verify and record its license first).
- Keep React + Vite + react-router + TanStack Query + Tailwind, routing, and search UI.
- Point the SPA's data layer at our **static contract** (S3/CloudFront in production; local files in dev).
- **Adapt the resource-page rendering component** to render the Nix-constructor document bodies our
  generator emits, instead of scraped HCL markdown. Surface the compat badge on provider pages.

## Capabilities

### New Capabilities
- `registry-ui-fork`: A forked registry-ui SPA wired to the Nivis static contract, with resource pages rendering Nix constructors and a visible compat badge.

### Modified Capabilities

## Impact

- New `frontend/` (forked SPA). Build output is static, S3/CloudFront-ready.
- License of the upstream UI MUST be checked and recorded before vendoring.
- Depends on `contract-generator`.
