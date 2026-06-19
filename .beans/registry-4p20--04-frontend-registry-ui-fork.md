---
# registry-4p20
title: 04 Frontend (registry-ui fork)
status: completed
type: milestone
priority: normal
created_at: 2026-06-19T11:14:11Z
updated_at: 2026-06-19T12:24:22Z
---

Fork registry-ui; adapt resource-page rendering to Nix constructors; wire to static contract. OpenSpec change: frontend-fork.

## Summary of Changes

- Verified upstream registry-ui license = MPL-2.0 (fork/redistribution permitted); recorded
  provenance in frontend/README.md. This is a focused fork: same stack + contract model, clean-room
  (does not copy upstream source).
- Stack: Vite + React 19 + react-router (HashRouter for static hosting) + TanStack Query + Tailwind,
  built with pnpm in the nix dev shell. esbuild build approved via pnpm-workspace.yaml.
- src/lib/contract.ts mirrors registry-ui's ProviderVersion/ProviderDocs/ProviderDocItem + the Nivis
  compat record. src/lib/data.ts reads the static contract (registry/docs/... + catalog.json) — same
  paths a later API would serve, so the data layer never changes.
- ADAPTED rendering: src/pages/ResourcePage.tsx renders the Nix-constructor doc body (signature,
  args, computed outputs via refAttr) with NO HCL resource block; src/components/CompatBadge.tsx
  surfaces the 4-axis badge honestly. Minimal dependency-free Markdown parser (src/lib/markdown.ts).
- tools/generate now also emits registry/catalog.json for the browsable index (no live search in v1).
- Tests: vitest markdown parser + React ResourcePage render (asserts Nix shown, no HCL). pnpm build
  produces a self-contained static dist/ with the contract bundled. Added a frontend CI job.
