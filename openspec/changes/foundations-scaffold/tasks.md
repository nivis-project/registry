## 1. Version control

- [ ] 1.1 Confirm `jj` repo (colocated with git), remote `origin` = `git@github.com:nivis-project/registry.git`
- [ ] 1.2 Confirm repo author is `Pim Snel <post@pimsnel.com>`; ensure NO Claude co-author trailer in any commit
- [ ] 1.3 Add `.gitignore` (Nix `result*`, Go build output, `node_modules`, `.direnv`)

## 2. Nix flake

- [ ] 2.1 Write `flake.nix` copying nivis's plain-nix `forAllSystems = f: builtins.listToAttrs (map …)` (NO flake-utils)
- [ ] 2.2 Pin systems: `x86_64-linux aarch64-linux x86_64-darwin aarch64-darwin`
- [ ] 2.3 Add `nivis` as a flake input (`github:wearetechnative/nivis`) and expose it in the devShell
- [ ] 2.4 devShell: go, nodejs + pnpm, jj, beans, openspec
- [ ] 2.5 Package/app stubs for the backend generator (`tools/generate`)
- [ ] 2.6 `nix flake check` passes; `nix develop` enters a working shell

## 3. Tooling boards

- [ ] 3.1 Confirm beans board present with milestones `01`–`06` and epics (see `beans list`)
- [ ] 3.2 Confirm OpenSpec changes present and `openspec validate` passes
- [ ] 3.3 `AUTONOMOUS-BUILD.md` + `REGISTRY-DESIGN.md` present at repo root

## 4. CI / test harness

- [ ] 4.1 Add Go module(s) under `tools/` with a passing trivial test
- [ ] 4.2 Add a CI workflow that runs `nix flake check` + `go test ./...`
- [ ] 4.3 Verify the harness is green before any extraction work begins
