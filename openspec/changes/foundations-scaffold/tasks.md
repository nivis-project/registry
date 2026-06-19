## 1. Version control

- [x] 1.1 Confirm `jj` repo (colocated with git), remote `origin` = `git@github.com:nivis-project/registry.git`
- [x] 1.2 Confirm repo author is `Pim Snel <post@pimsnel.com>`; ensure NO Claude co-author trailer in any commit
- [x] 1.3 Add `.gitignore` (Nix `result*`, Go build output, `node_modules`, `.direnv`)

## 2. Nix flake

- [x] 2.1 Write `flake.nix` copying nivis's plain-nix `forAllSystems = f: builtins.listToAttrs (map …)` (NO flake-utils)
- [x] 2.2 Pin systems: `x86_64-linux aarch64-linux x86_64-darwin aarch64-darwin`
- [x] 2.3 Add `nivis` as a flake input (`github:wearetechnative/nivis`) and expose it in the devShell
- [x] 2.4 devShell: go, nodejs + pnpm, jj, beans, openspec
- [x] 2.5 Package/app stubs for the backend generator (`tools/generate`)
- [x] 2.6 `nix flake check` passes; `nix develop` enters a working shell

## 3. Tooling boards

- [x] 3.1 Confirm beans board present with milestones `01`–`06` and epics (see `beans list`)
- [x] 3.2 Confirm OpenSpec changes present and `openspec validate` passes
- [x] 3.3 `AUTONOMOUS-BUILD.md` + `REGISTRY-DESIGN.md` present at repo root

## 4. CI / test harness

- [x] 4.1 Add Go module(s) under `tools/` with a passing trivial test
- [x] 4.2 Add a CI workflow that runs `nix flake check` + `go test ./...`
- [x] 4.3 Verify the harness is green before any extraction work begins
