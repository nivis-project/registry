## 1. Fork

- [x] 1.1 Verify and record `opentofu/registry-ui` license; confirm redistribution is permitted
- [x] 1.2 Vendor the SPA under `frontend/` (React + Vite + react-router + TanStack Query + Tailwind, pnpm)
- [x] 1.3 Confirm `pnpm install` + dev server run inside the Nix dev shell

## 2. Data wiring

- [x] 2.1 Point the data layer at the Nivis static contract (local files in dev)
- [x] 2.2 Build to static assets; confirm pages render with no live backend

## 3. Rendering adaptation

- [x] 3.1 Adapt the resource-page component to render Nix-constructor document bodies
- [x] 3.2 Surface the compat badge on provider pages (`compatible by design` vs `verified`)
- [x] 3.3 Verify a resource page contains no HCL `resource` block as the reference
