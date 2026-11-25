# `random_uuid`

Nivis Nix constructor for the `random_uuid` resource (provider `random`). This reference is derived from the provider binary's own schema via `nivis gen` — it always matches the real provider.

## Constructor

```nix
{ name, keepers ? null, overrides ? {} }
```

## Required arguments

_None._

## Optional arguments

- `keepers` (defaults to `null`)

## Computed outputs

Read these from the resource's `refAttr`:

- `id`
- `result`

## Example

```nix
random_uuid {
  name = "uuid";
}
```
