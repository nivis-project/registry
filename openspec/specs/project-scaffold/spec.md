# project-scaffold Specification

## Purpose
TBD - created by archiving change foundations-scaffold. Update Purpose after archive.
## Requirements
### Requirement: Version control with correct attribution

The repository SHALL use jj (colocated with git) with remote `origin` set to
`git@github.com:nivis-project/registry.git`, and all commits MUST be authored as
`Pim Snel <post@pimsnel.com>` with NO Claude co-author trailer or self-promotion.

#### Scenario: Commits are attributed to Pim Snel

- **WHEN** the bootstrap commit is created
- **THEN** the author is `Pim Snel <post@pimsnel.com>`
- **AND** the commit message contains no `Co-Authored-By: Claude` line and no tool self-promotion

### Requirement: Planning boards initialized

The project SHALL have a beans board (milestones `01`–`06` with child epics) and an OpenSpec
setup whose changes pass `openspec validate`.

#### Scenario: Boards are present and valid

- **WHEN** `beans list --json` and `openspec validate --all` are run
- **THEN** beans reports milestones `01` through `06` with child epics
- **AND** every OpenSpec change validates without errors

### Requirement: Green test harness before extraction

A CI/test harness SHALL keep `nix flake check` and `go test ./...` green, and this MUST be
true before any provider-extraction work begins.

#### Scenario: Harness is green on a clean checkout

- **WHEN** `nix flake check` and `go test ./...` run on a fresh clone
- **THEN** both succeed with no failures

