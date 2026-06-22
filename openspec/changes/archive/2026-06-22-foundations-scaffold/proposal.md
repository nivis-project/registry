## Why

The Nivis registry PoC needs a reproducible foundation before any pipeline or frontend work
can start: a Nix-flake dev environment, version control (jj), the beans board, and a hermetic
test harness. Without this, every later change re-litigates environment and tooling decisions.
This change establishes the scaffold the rest of the milestones build on.

## What Changes

- Initialize the repo with **jj** (colocated with git), remote `git@github.com:nivis-project/registry.git`,
  author **Pim Snel <post@pimsnel.com>**, **no Claude co-author trailer**.
- Add a **`flake.nix`** using the **plain-nix `forAllSystems`/`listToAttrs` pattern copied from nivis**
  (NO flake-utils), the four systems `x86_64-linux aarch64-linux x86_64-darwin aarch64-darwin`, a
  devShell (go, nodejs/pnpm, jj, beans, openspec, the `nivis` flake input), and package/app stubs.
- Wire **`nivis` as a flake input** so `nivis gen` is available hermetically.
- Establish the **CI/test harness**: `nix flake check` green; Go test scaffolding under `tools/`.
- Confirm beans + OpenSpec are initialized and the board/changes are present.

## Capabilities

### New Capabilities
- `dev-environment`: A reproducible Nix-flake dev shell and flake outputs (no flake-utils), pinned to four systems, exposing the toolchain and the nivis input.
- `project-scaffold`: jj VC with the correct author/remote, beans board, OpenSpec changes, and the CI/test harness that keeps `nix flake check` green.

### Modified Capabilities

## Impact

- New files: `flake.nix`, `flake.lock`, `nix/`, `.beans/`, `openspec/`, CI workflow.
- Tooling: jj, nix, beans, openspec, go, pnpm.
- No provider binaries are executed in this change.
