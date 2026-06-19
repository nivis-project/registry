# CLAUDE.md — Nivis Registry

**Start here:** read **`AUTONOMOUS-BUILD.md`** (the master brief) and **`REGISTRY-DESIGN.md`**
(architecture). They contain the locked decisions, verified ground truth, run order, and the hard
**PAUSE gate**. Do not relitigate decisions recorded there.

## Working rules

- **Tracker = beans** (NOT TodoWrite). Set the relevant bean `in-progress` before working it; check off
  `tasks.md` items as you go; mark `completed` only when no unchecked items remain; add a
  `## Summary of Changes` on completion. Run `beans list --json`. Milestones are titled `01`…`06`.
- **Spec before code = OpenSpec.** Work the existing changes (`openspec list`); create a new change for
  any work not already covered. Validate with `openspec validate --all`.
- **VC = jj**, colocated with git. Remote `origin` = `git@github.com:nivis-project/registry.git`.
  **Commit as `Pim Snel <post@pimsnel.com>`. No `Co-Authored-By: Claude` trailer. No self-promotion.**
- **Nix**: `flake.nix` is plain-nix (no flake-utils), four systems, `nivis` as an input. Keep
  `nix flake check` and `go test ./...` (in `tools/`) green.
- **The PAUSE gate is real**: prove on ~3 real providers + 1 hermetic fake, push to `main`, then STOP.
  Do not scale to 50 or deploy — that is milestone `06`.

## Map

| Milestone | OpenSpec change | What |
|-----------|-----------------|------|
| 01 Foundations | `foundations-scaffold` | flake, VC, boards, CI harness |
| 02 Extraction pipeline | `extraction-pipeline` | seed → extract (verify + `nivis gen`) → compat |
| 03 Contract and generator | `contract-generator` | emit JSON contract + render Nix constructors |
| 04 Frontend | `frontend-fork` | fork registry-ui; adapt rendering to Nix |
| 05 Proof and e2e | `proof-e2e` | hermetic + 3 real; **PAUSE gate** |
| 06 Roadmap (later) | — | scale/deploy, RSS sync, transpiler, search API |
