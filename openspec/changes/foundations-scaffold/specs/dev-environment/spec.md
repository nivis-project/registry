## ADDED Requirements

### Requirement: Reproducible Nix flake without flake-utils

The project SHALL provide a `flake.nix` that enumerates supported systems with plain Nix
(`builtins.listToAttrs` over an explicit `systems` list, the pattern used by the nivis flake)
and MUST NOT depend on `flake-utils`.

#### Scenario: Flake evaluates on all pinned systems

- **WHEN** `nix flake show` is run
- **THEN** outputs are produced for `x86_64-linux`, `aarch64-linux`, `x86_64-darwin`, and `aarch64-darwin`
- **AND** the flake inputs do not include `flake-utils`

#### Scenario: Dev shell provides the toolchain

- **WHEN** `nix develop` is entered
- **THEN** `go`, `pnpm`, `jj`, `beans`, and `openspec` are on `PATH`
- **AND** the `nivis` CLI is available via the `nivis` flake input

### Requirement: nivis available hermetically

The flake SHALL declare `nivis` (`github:wearetechnative/nivis`) as an input so `nivis gen`
can be invoked without an ambient install.

#### Scenario: nivis gen is runnable from the dev shell

- **WHEN** a developer runs `nivis gen --help` inside `nix develop`
- **THEN** the command resolves from the flake input and prints usage
