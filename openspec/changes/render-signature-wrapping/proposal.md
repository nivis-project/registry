## Why

The rendered Nix constructor signature is emitted on a single line. For providers
whose resources have many arguments — common when several are **mandatory**
(e.g. scaleway, azurerm, google resources) — that one-liner runs far past a
readable width and is hard to scan. The signature should adapt to a sensible
maximum width: stay on one line when short, but break to **one argument per
line** when wide.

## What Changes

- `tools/generate/render.go` (`signature`): render the constructor argument set
  width-aware. A signature within the wrap width stays on one line; a wider one
  breaks to one argument per line:

  ```nix
  {
    name,
    arg1,
    arg2 ? null,
    overrides ? {}
  }
  ```

  Required arguments appear without `? null`; optional arguments keep `? null`;
  the last argument has no trailing comma. This is a presentation change to the
  rendered document only — the parsed `Constructor` model and the contract shape
  are unchanged.
- Update the `nix-rendering` spec so the wrapping rule is the documented
  requirement going forward.

## Capabilities

### Modified Capabilities
- `nix-rendering`: the rendered constructor signature is formatted for
  readability — one line when short, one argument per line when it exceeds a
  maximum width.

## Impact

- Changed: `tools/generate/render.go`, its tests, and the `nix-rendering` spec.
- Regenerating the contract reflows wide signatures (e.g. `random_password`,
  scaleway/azurerm resources). No contract-shape or extraction change; no deploy.
