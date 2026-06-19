# frontend/ — forked registry-ui SPA

This directory will hold a **fork of [`opentofu/registry-ui`](https://github.com/opentofu/registry-ui)**,
the React + Vite SPA that powers search.opentofu.org. It is **data-agnostic**: it consumes a JSON
contract, so it works against our static contract with no live backend (v1) and the later API
unchanged.

## What to keep vs adapt (see `frontend-fork` OpenSpec change)

**Keep:** React 19 + Vite + react-router + TanStack Query + Tailwind; routing; search UI; the contract
TypeScript types (generated from `backend/internal/server/openapi.yml`).

**Adapt:** the **resource-page rendering component** — render the **Nix-constructor** document bodies
our generator emits (signature, typed attributes, computed outputs), NOT scraped HCL markdown. Surface
the **compat badge** (tier, protocol, architectures, e2e status) on provider pages.

**Wire:** point the data layer at the Nivis static contract — local files in dev, S3 + CloudFront in
production.

## Before vendoring (MUST)

- Verify and record the upstream **license** (confirm redistribution is permitted) — note it in this
  file and in the repo `README.md`.
- Confirm `pnpm install` + dev server run inside `nix develop`.

## Not in v1 (milestone 06)

Search backend (API Gateway + Lambda), R2/Neon/Worker hosting. v1 is static only.
