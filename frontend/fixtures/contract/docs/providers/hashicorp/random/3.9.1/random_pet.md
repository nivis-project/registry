# `random_pet`

Nivis Nix constructor for the `random_pet` resource (provider `random`). This reference is derived from the provider binary's own schema via `nivis gen` — it always matches the real provider.

## Constructor

```nix
{
  name,
  keepers ? null,
  length ? null,
  prefix ? null,
  separator ? null,
  overrides ? {}
}
```

## Required arguments

_None._

## Optional arguments

- `keepers` (defaults to `null`)
- `length` (defaults to `null`)
- `prefix` (defaults to `null`)
- `separator` (defaults to `null`)

## Computed outputs

Read these from the resource's `refAttr`:

- `id`

## Example

```nix
random_pet {
  name = "pet";
}
```
