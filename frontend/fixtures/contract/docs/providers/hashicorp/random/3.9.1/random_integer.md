# `random_integer`

Nivis Nix constructor for the `random_integer` resource (provider `random`). This reference is derived from the provider binary's own schema via `nivis gen` — it always matches the real provider.

## Constructor

```nix
{ name, max, min, keepers ? null, seed ? null, overrides ? {} }
```

## Required arguments

- `max`
- `min`

## Optional arguments

- `keepers` (defaults to `null`)
- `seed` (defaults to `null`)

## Computed outputs

Read these from the resource's `refAttr`:

- `id`
- `result`

## Example

```nix
random_integer {
  name = "integer";
  max = …;  # required
  min = …;  # required
}
```
