# AUTONOMOUS BUILD BRIEF — Nivis Registry PoC

> **Read this first.** This file is the self-contained brief for an autonomous Claude Code run.
> It encodes every decision, the verified ground truth, the run order, and a **hard PAUSE gate**.
> You should not need any outside chat context to execute it.

## Mission

Build the first working version (alpha base) of the **Nivis registry**: a public, browsable catalog
of OpenTofu-compatible providers with **Nix-native** documentation. Every OpenTofu provider is
Nivis-compatible *by design* (Nivis runs unmodified provider binaries over the plugin protocol). The
registry's unique value is **schema-derived Nix documentation**, rendered from nivis's existing
`nivis gen` codegen — better and more legally defensible than scraping HCL prose.

See `REGISTRY-DESIGN.md` for the architecture. Track work in **beans** and **OpenSpec** (below).

## Locked decisions (do not relitigate)

- **Build approach A**: real `nivis gen` extraction (download + execute real provider binaries, sandboxed).
- **Architecture = OpenTofu's split**: a Go backend generator emits a static JSON contract; a **forked
  `registry-ui` React SPA** consumes it. Reuse its routing/search/contract-types; **adapt only the
  resource-page rendering** to show Nix constructors instead of scraped HCL markdown.
- **Host**: AWS **S3 + CloudFront** for v1 static; **API Gateway + Lambda** later (additive, same contract).
- **Seed**: 50 providers — hyperscalers + official utilities + top community by downloads + **Telmate/proxmox**.
  Derived by a script over `registry.terraform.io/v1/providers`, then pinned to `seed.json`.
- **Compat tiers + badge** required (see design doc for the four axes).
- **Modules dropped** for now (TF-module compat is a later maybe).

## Guardrails (MUST)

- **VC**: jj, colocated with git. Remote `origin` = `git@github.com:nivis-project/registry.git`.
- **Author every commit as `Pim Snel <post@pimsnel.com>`. NO `Co-Authored-By: Claude` trailer. NO tool
  self-promotion anywhere in commit messages, docs, or site copy.**
- **Push to `main`** (this repo is new/empty; the proof commit establishes it).
- Provider binaries execute **in-sandbox only**. If a single provider's extraction fails, **skip it and
  log the reason — never hang or abort the whole run.**
- **Spec before code**: for any new work not already covered, create an OpenSpec change first.
- **Beans is the tracker** (not TodoWrite): set the relevant bean `in-progress` before working it, check
  off `tasks.md` items as you go, mark `completed` only when no unchecked items remain, and add a
  `## Summary of Changes` on completion.

## Verified ground truth (already researched — use directly)

- **`nivis gen`**: `nivis gen --provider <BINARY_PATH> [--identity <id>] [--out <dir>]` → emits
  `<out>/<id>/<type>.nix` typed constructors with doc-comment headers (computed outputs, nested-block
  shapes). Takes a **local binary path**, NOT a registry address. Available via the `nivis` flake input:
  `nix run github:wearetechnative/nivis#nivis -- gen …`. nivis is Apache-2.0.
- **nivis registry client** (`internal/registry` in the nivis repo): resolves `hashicorp/aws` →
  OpenTofu registry → download → **SHA256-verify** → cache. Reuse this *model* for `tools/extract`
  (call nivis to fetch, or mirror its verify-before-execute behavior).
- **nivis fake providers** (`cmd/provider-{alpha,beta,delta,gamma}`, `go build`, no net/creds): the
  hermetic e2e fixtures. Build one and run it through the whole pipeline offline.
- **nivis `flake.nix`**: enumerates `systems = ["x86_64-linux" "aarch64-linux" "x86_64-darwin"
  "aarch64-darwin"]` via `forAllSystems = f: builtins.listToAttrs (map …)` — **plain nix, NO
  flake-utils. Copy this pattern.**
- **OpenTofu `registry-ui`** (`opentofu/registry-ui`): React 19 + Vite + react-router + TanStack Query +
  Tailwind (pnpm). A **data-agnostic SPA over a JSON contract** (`backend/internal/server/openapi.yml`):
  `GET /registry/docs/providers/{ns}/{name}/{version}/index.json` + per-item documents. Forkable. Its
  R2/Neon/Worker hosting is OpenTofu-specific and NOT reused. **Check its license before vendoring.**
- **Popularity API**: `https://registry.terraform.io/v1/providers?limit=N&offset=M` is public; returns
  `tier` (official/partner/community) + `downloads`. The `filter=` param is **inert** — page and sort
  client-side. Most-popular proxmox = **Telmate/proxmox** (~16M dl; `bpg/proxmox` is the modern runner-up).

## Boards

**Beans milestones** (titles ordered `01`…`06`; epics are children — run `beans list --json`):

| # | Milestone | OpenSpec change |
|---|-----------|-----------------|
| 01 | Foundations | `foundations-scaffold` |
| 02 | Extraction pipeline | `extraction-pipeline` |
| 03 | Contract and generator | `contract-generator` |
| 04 | Frontend (registry-ui fork) | `frontend-fork` |
| 05 | Proof and e2e (PAUSE gate) | `proof-e2e` |
| 06 | Roadmap (later) — **deferred** | — |

Run `openspec list` for tasks; `openspec show <change>` for detail; `openspec validate <change>`.

## Run order (and the PAUSE gate)

1. **Foundations** (`foundations-scaffold`): finish the flake (plain-nix, 4 systems, `nivis` input,
   devShell), `.gitignore`, CI harness. `nix flake check` + `go test ./...` green. Commit.
2. **Extraction** (`extraction-pipeline`): `tools/seed` → `seed.json`; `tools/extract` (verify-before-
   execute → `nivis gen`); `tools/compat`. Prove on **3 small real providers**
   (`hashicorp/random`, `hashicorp/null`, `Telmate/proxmox`) — **not** giant AWS/Azure in the proof.
3. **Contract + generator** (`contract-generator`): `tools/generate` emits the registry-ui contract;
   render Nix constructors into per-item docs; embed the compat badge.
4. **Frontend** (`frontend-fork`): fork registry-ui under `frontend/`; adapt the resource-page
   rendering to the Nix-constructor docs; wire to the static contract; surface the badge.
5. **Proof** (`proof-e2e`): hermetic fake-provider e2e (offline, in CI) + the 3-real-provider e2e, all
   GREEN; site builds locally.
6. **★ PAUSE GATE ★**: commit (as Pim Snel) and **push to `main`**, then **STOP and report**.
   **Do NOT** scale to 50 providers. **Do NOT** deploy to S3/CloudFront. Milestone `06` stays `todo`/`draft`.
   Report: what was proven, where the artifacts are, and what milestone `06` holds for the next run.

## Definition of done for this run

- `nix flake check`, `go test ./...`, and the hermetic e2e are green.
- The 3 real providers have extraction output, contract files, rendered pages, and compat badges.
- The forked SPA builds and renders a Nix-constructor resource page against the static contract.
- One commit authored `Pim Snel <post@pimsnel.com>` (no Claude trailer), pushed to `main`.
- A report is printed and the run stops at the gate.
