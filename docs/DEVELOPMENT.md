# Development Guide — Nivis Registry

A hands-on guide to working on the registry: the toolchain, the pipeline, how to
run it locally, and how to extend it. For *why* the system is shaped this way,
read [`REGISTRY-DESIGN.md`](../REGISTRY-DESIGN.md); for the design's
project-management context read [`AUTONOMOUS-BUILD.md`](../AUTONOMOUS-BUILD.md)
and [`CLAUDE.md`](../CLAUDE.md). This guide is the *what* and *how*.

## TL;DR

```sh
nix develop                       # go, node, pnpm, jj, nivis on PATH
nix develop -c scripts/proof.sh   # extract → contract → build the static site
nix develop -c bash -c 'cd frontend && pnpm dev'   # browse at http://127.0.0.1:5173/
```

`scripts/proof.sh` is the whole pipeline end to end. Read it first — it is the
shortest accurate description of how the pieces connect.

## The big picture

The registry is a **Go backend generator** that emits a **static JSON contract**,
plus a **React SPA** that renders it. The only Nivis-specific work is the
generator; the SPA is data-agnostic (it consumes a JSON contract, not Nivis or
OpenTofu concepts).

```
                         tools/  (Go backend generator)
  ┌───────┐   ┌────────────┐   ┌──────────┐   ┌────────────┐
  │ seed  │──▶│  extract   │──▶│  compat  │──▶│  generate  │──▶ static contract
  │       │   │ verify +   │   │ 4-axis   │   │ parse .nix │    registry/docs/...
  │seed.  │   │ nivis gen  │   │ badge    │   │ render Nix │    + catalog.json
  │ json  │   └────────────┘   └──────────┘   └────────────┘         │
  └───────┘                                                          │ reads (HTTP/files)
                         frontend/  (React + Vite SPA)               ▼
  ┌─────────────────────────────────────────────────────────────────────────┐
  │  IndexPage → ProviderPage → ResourcePage (renders the Nix constructor)    │
  │  data-agnostic: consumes the JSON contract; resource page ADAPTED to Nix  │
  └─────────────────────────────────────────────────────────────────────────┘
```

Repo layout:

| Path | What |
|------|------|
| `tools/` | the Go backend generator (one module, `go 1.23`, no external deps) |
| `tools/seed/`, `tools/extract/`, `tools/compat/`, `tools/generate/` | the pipeline libraries |
| `tools/cmd/{seed,extract,generate}/` | thin CLI drivers over those libraries |
| `tools/e2e/` | hermetic + real-provider end-to-end tests |
| `frontend/` | the SPA (focused fork of `opentofu/registry-ui`, MPL-2.0) |
| `scripts/proof.sh` | the end-to-end proof: extract → generate → build site |
| `flake.nix` / `flake.lock` | the dev env + the pinned `nivis` input |
| `openspec/` | the spec-before-code change proposals + tasks |
| `.beans/` | milestones (`01`…`06`) and epics |

## The toolchain (Nix dev shell)

Everything runs inside `nix develop`, which puts `go`, `node`, `pnpm`, `jj`, and
the pinned `nivis` CLI on `PATH` (see `flake.nix`). You do not install anything
globally.

```sh
nix develop                  # enter the shell
nix flake check              # build tools/ + run `go test ./...` + gofmt, hermetically
nix build .#nivis-cli        # build the exact pinned nivis CLI (used by CI)
nix eval --raw .#nivisRev    # the pinned nivis git rev
```

`flake.nix` is **plain Nix** (no flake-utils), four systems, with `nivis` as a
flake input. `nix flake check` is the harness gate — keep it green.

### The `nivis` pin

`nivis gen` is the heart of extraction, so the registry pins a specific `nivis`
commit in `flake.lock`. To move to a new nivis release:

```sh
nix flake update nivis       # bump the input to nivis HEAD/tag
nix eval --raw .#nivisRev    # confirm the new rev
```

Two nivis releases shaped this project (worth knowing the history):

- **0.4.2** — `nivis gen` stopped calling `ConfigureProvider` before fetching the
  schema, which unblocked credential-requiring providers (azurerm/google/proxmox).
- **0.4.3** — fixed a reserved-name collision: a provider attribute literally
  named `name` (or `overrides`/`nivis`) is now aliased to `cfg_<name>` so the
  generated constructor is valid Nix.

The registry is pinned to **0.4.3**. The commit history (`jj log` / `git log`)
records each pin bump and what it unblocked.

## The pipeline, stage by stage

Each stage is a small Go library under `tools/` with a CLI driver under
`tools/cmd/`. The libraries are pure and unit-tested; the drivers wire flags and
logging.

### 1. `seed` — pin the provider list

`tools/seed` derives a deterministic `seed.json` from the public Terraform
registry popularity API.

- **Source:** `https://registry.terraform.io/v1/providers?limit=N&offset=M`.
  ⚠️ The `filter=` query param is **inert** — page and sort client-side.
- **Selection** (`seed.Select`): explicit anchors (`aws`, `azurerm`, `google`,
  `Telmate/proxmox`) first, then a fixed utility allowlist (`random`, `null`,
  `tls`, …), then the most-popular remaining by `(tier, -downloads, address)`,
  capped at `MaxSeed = 50`. Output is **byte-stable** across runs.
- **Run:** `cd tools && go run ./cmd/seed -out ../seed.json`

`seed.json` is checked in. It carries `{address, tier, downloads, reason}` per
provider, where `reason` is `anchor` / `utility` / `popular` so the manifest is
self-documenting.

> Note: `seed.json` pins the *catalog* (what the registry knows about). The proof
> set actually extracted today is the smaller list in `scripts/proof.sh`. Scaling
> extraction to all 50 is milestone `06`.

### 2. `extract` — verified binary → `nivis gen`

`tools/extract` turns a provider address into generated Nix constructors, the
**verify-before-execute** way (mirrors nivis's own registry client):

1. **Resolve** the latest version against the OpenTofu registry
   (`registry.opentofu.org`), capturing `protocols` + `platforms` metadata.
2. **Download** the platform archive + its `SHA256SUMS`.
3. **Verify** the archive checksum — `extract.VerifyArchive`. A mismatch aborts;
   **the binary is never unpacked or executed** on failure.
4. **Unpack** the `terraform-provider-*` executable into the cache.
5. **Run** `nivis gen --provider <binary> --identity <name> --out <dir>` —
   `extract.RunGen`.

The driver `Batch` is **resilient**: one provider's failure is skipped + logged
(with a per-provider timeout) and never aborts the run.

- **Run:** `cd tools && go run ./cmd/extract -nivis "$(command -v nivis)" -out ../extract-out random null tls Telmate/proxmox`
  - bare `<name>` expands to `hashicorp/<name>`; or pass `-seed ../seed.json`.
- **Output tree** (`extract-out/<ns>/<name>/<version>/`):

  ```
  hashicorp/random/3.9.0/metadata.json           # registry metadata (version, protocols, platforms)
  hashicorp/random/3.9.0/compat.json             # the 4-axis compat record (written by the driver)
  hashicorp/random/3.9.0/random/random_id.nix    # one nivis gen constructor per resource type
  hashicorp/random/3.9.0/random/random_password.nix
  ...
  ```

`nivis gen` emits **resource** constructors only (no datasources in the current
nivis). Each `.nix` file is a typed constructor: a `{ nivis }:` module returning a
lambda `{ name, <required>, <optional> ? null, overrides ? {} }:` plus a header
comment listing the computed outputs.

### 3. `compat` — the honest badge

`tools/compat` computes a per-provider `Record` from four **independent** axes
(`compat.Compute`):

1. **`schema_extractable`** — did the pipeline get a schema? (machine-verified)
2. **`protocols`** — `5.0` / `6.0`, from the registry metadata.
3. **`architectures`** — `NivisSupportedArches ∩ published platforms` (a real
   intersection: a provider that ships only `linux/amd64` can't serve arm64).
4. **`e2e`** — `none` | `verified`, from the hand-maintained `E2EVerified`
   allowlist (currently empty — nothing is `verified` until a real e2e proves it).

The headline `tier` is `compatible by design` unless a provider is in the e2e
allowlist (then `e2e verified`). **The badge never claims `verified` without an
allowlist entry** — this invariant is unit-tested.

### 4. `generate` — contract + Nix-rendered docs

`tools/generate` walks the extraction tree and emits the static contract the SPA
consumes. Three files do the work:

- **`nixparse.go`** (`ParseConstructor`) — parses a `nivis gen` `.nix` file into a
  `Constructor` model: provider, type, required args (the `throw`-guarded ones),
  optional args, and computed outputs (from the header comment). Tolerant: a field
  it can't find is left empty rather than failing.
- **`render.go`** (`RenderDoc`) — renders a `Constructor` as a Markdown doc body
  in **Nix** form: the constructor signature, required/optional arguments,
  computed outputs (read via `refAttr`), and a Nix usage example. It emits **no
  HCL `resource {}` block** — `ContainsHCLResourceBlock` is used in tests to prove
  this.
- **`generate.go`** (`GenerateProvider` / `GenerateAll`) — for each provider
  version, writes `index.json` (the registry-ui `ProviderVersion` shape:
  `docs.{resources,datasources,functions,guides}` + the embedded `compat` record)
  and one `<type>.md` per resource. `GenerateAll` also writes `catalog.json` (the
  browsable provider list, since v1 has no live search backend).

- **Run:** `cd tools && go run ./cmd/generate -extract ../extract-out -contract ../frontend/public`
- **Output tree** (`registry/docs/providers/<ns>/<name>/<version>/`):

  ```
  registry/catalog.json                                          # browsable list
  registry/docs/providers/hashicorp/random/3.9.0/index.json      # ProviderVersion + compat
  registry/docs/providers/hashicorp/random/3.9.0/random_id.md    # rendered Nix doc per item
  ...
  ```

The contract types mirror `opentofu/registry-ui`'s
`backend/internal/server/openapi.yml`, so the SPA's contract types are reusable;
the only addition is the Nivis `compat` field.

## The frontend (`frontend/`)

A **focused fork** of `opentofu/registry-ui` (MPL-2.0): same stack and contract
model, with the resource-page rendering adapted to Nix. It is a clean-room
implementation on the registry-ui pattern (it does not copy upstream source). See
`frontend/README.md` for the license provenance.

Stack: Vite + React 19 + react-router (`HashRouter`, so it works as plain static
files) + TanStack Query + Tailwind, built with pnpm.

Source map:

| File | Role |
|------|------|
| `src/lib/contract.ts` | the contract TS types (mirror registry-ui + `compat`) |
| `src/lib/data.ts` | the data layer — fetches `registry/docs/...` + `catalog.json` |
| `src/lib/markdown.ts` | a tiny dependency-free Markdown block parser |
| `src/pages/IndexPage.tsx` | provider list (from `catalog.json`) |
| `src/pages/ProviderPage.tsx` | a provider's resources + the compat badge |
| `src/pages/ResourcePage.tsx` | **the adapted part** — renders the Nix-constructor doc |
| `src/components/CompatBadge.tsx` | the four-axis badge |
| `src/components/MarkdownDoc.tsx` | renders parsed Markdown blocks to React |

The data layer reads relative paths under `registry/docs/providers/...` — the
**same paths a future API (API Gateway + Lambda) would serve** — so the SPA does
not change when hosting evolves (static now; dynamic later, milestone `06`).

Commands (inside `nix develop`, from `frontend/`):

```sh
pnpm install          # esbuild build is approved in pnpm-workspace.yaml
pnpm dev              # dev server (http://127.0.0.1:5173/)
pnpm test             # vitest: markdown parser + ResourcePage render (no-HCL)
pnpm build            # tsc -b && vite build → dist/ (self-contained static site)
```

`pnpm build` produces a self-contained `dist/` with the contract bundled — ready
for S3 + CloudFront. The contract under `frontend/public/registry/` and `dist/`
is a **build artifact** (gitignored), regenerated by `tools/generate`.

## Running it locally, end to end

The one command that ties it all together:

```sh
nix develop -c scripts/proof.sh
```

It extracts the proof set (`random null tls Telmate/proxmox azurerm google`),
generates the contract into `frontend/public`, and builds the static site. Then:

```sh
nix develop -c bash -c 'cd frontend && pnpm dev'   # browse http://127.0.0.1:5173/
```

To run only part of it, run the stage drivers directly (see each stage above).

## Testing

- **`go test ./...`** (from `tools/`) — unit tests for every stage. The e2e tests
  **skip** unless their prerequisites are present, so this stays green anywhere.
- **`nix flake check`** — builds `tools/` and runs the suite hermetically, plus a
  gofmt gate. The harness gate.
- **Hermetic e2e** (`tools/e2e` `TestHermeticPipeline`) — builds a nivis fake
  provider and runs it through extract → contract → render **offline**; asserts a
  Nix page with no HCL. Needs `NIVIS_BIN` + `NIVIS_SRC`:

  ```sh
  NIVIS_BIN="$(command -v nivis)" NIVIS_SRC=/path/to/nivis/checkout \
    go test ./e2e/ -run TestHermeticPipeline
  ```

- **Real-provider e2e** (`TestRealProviderPipeline`) — runs the proof providers
  through the full pipeline and `nix eval`s a sample of the generated
  constructors (so the "duplicate formal argument" class can't regress). Needs
  network + `nix` on PATH:

  ```sh
  NIVIS_BIN="$(command -v nivis)" NIVIS_REGISTRY_NET_E2E=1 \
    go test ./e2e/ -run TestRealProviderPipeline
  ```

CI (`.github/workflows/ci.yml`) runs: `nix flake check`, `go test ./...`, the
hermetic e2e (building the pinned nivis CLI + matching source), and the frontend
build/test.

## How to … (common tasks)

### Add a provider to the proof set
Edit the `PROVIDERS=(…)` array in `scripts/proof.sh` (and the list in
`tools/e2e/e2e_test.go` `TestRealProviderPipeline` if you want it e2e-covered),
then re-run the proof. The seed catalog (`seed.json`) is separate — regenerate it
with `go run ./cmd/seed` if you want to change the 50-provider catalog.

### Mark a provider e2e-verified
Add it to `E2EVerified` in `tools/compat/compat.go` **only when a real
end-to-end apply proves it**. The badge invariant (no `verified` without an
allowlist entry) is enforced by tests — keep it honest.

### Change how a resource page renders
The Nix rendering lives in `tools/generate/render.go` (`RenderDoc`) — the doc body
is generated there, not in the frontend. The frontend just displays the Markdown.
If you change the doc shape, update the tests in `tools/generate/generate_test.go`
(they assert the Nix form and the absence of HCL).

### Bump nivis
`nix flake update nivis`, then re-run `scripts/proof.sh` and the e2e to confirm
the new `nivis gen` output still parses and the constructors still `nix eval`.

## Conventions

- **Tracker = beans** (not ad-hoc TODOs): milestones `01`…`06` + epics. `beans list --json`.
- **Spec before code = OpenSpec**: a change proposal + tasks per unit of work.
  `openspec list`, `openspec validate --all`.
- **VC = jj**, colocated with git; remote `origin` = the `nivis-project/registry`
  repo. Commit as `Pim Snel <post@pimsnel.com>`, no `Co-Authored-By` trailer.
- **Go**: keep `gofmt` clean (the flake checks it) and `go test ./...` green.

## What's next (milestone 06, not built)

Scale extraction to all 50 seed providers, deploy to S3 + CloudFront, an upstream
RSS-delta sync, the search API (API Gateway + Lambda), and the HCL→Nix transpiler.
None of these change the contract or the SPA — they are additive.
