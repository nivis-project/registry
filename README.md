# Nivis Registry

A public, browsable catalog of OpenTofu-compatible providers with **Nix-native** documentation,
for the [Nivis](https://wearetechnative.github.io/nivis/) project.

Every OpenTofu provider is Nivis-compatible *by design* — Nivis runs unmodified provider binaries over
the plugin protocol. This registry makes them discoverable and publishes **schema-derived Nix
documentation** (rendered from `nivis gen`), plus an honest per-provider **compatibility badge**.

## Status

Bootstrapping the PoC (alpha base). The build is driven autonomously per **`AUTONOMOUS-BUILD.md`** —
start there. Architecture is in **`REGISTRY-DESIGN.md`**.

## Layout

- `tools/` — the Go backend generator: `seed` → `extract` → `compat` → `generate`.
- `frontend/` — forked `registry-ui` SPA (resource rendering adapted to Nix constructors).
- `flake.nix` — reproducible dev env (plain Nix, no flake-utils; four systems; `nivis` as an input).
- `.beans/` — milestones (`01`…`06`) and epics. `openspec/` — change proposals, specs, tasks.

## Develop

```sh
nix develop          # go, pnpm, jj, beans, openspec, nivis on PATH
nix flake check      # must stay green
beans list --json    # the board
openspec list        # the changes
```

## License

Project code: Apache-2.0 (matching nivis). The vendored `frontend/` retains its upstream license
(recorded when vendored).
