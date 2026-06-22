## 1. Width-aware signature

- [x] 1.1 `tools/generate/render.go` (`signature`): one line when within the wrap width; otherwise
  one argument per line (two-space indent, no trailing comma on the last, closing brace on its own line)
- [x] 1.2 Required args render without `? null`; optional args keep `? null`; `name` first, `overrides ? {}` last
- [x] 1.3 Tests: a wide constructor (e.g. `random_password`) wraps; a short one stays on one line

## 2. Spec + regen

- [x] 2.1 Update the `nix-rendering` spec with the width-aware wrapping requirement + scenarios
- [x] 2.2 `go test ./...` + `nix flake check` green
- [x] 2.3 Regenerate the contract so existing wide signatures reflow (no contract-shape change)
