# `random_string`

Nivis Nix constructor for the `random_string` resource (provider `random`). This reference is derived from the provider binary's own schema via `nivis gen` — it always matches the real provider.

## Constructor

```nix
{
  name,
  length,
  keepers ? null,
  lower ? null,
  min_lower ? null,
  min_numeric ? null,
  min_special ? null,
  min_upper ? null,
  number ? null,
  numeric ? null,
  override_special ? null,
  special ? null,
  upper ? null,
  overrides ? {}
}
```

## Required arguments

- `length`

## Optional arguments

- `keepers` (defaults to `null`)
- `lower` (defaults to `null`)
- `min_lower` (defaults to `null`)
- `min_numeric` (defaults to `null`)
- `min_special` (defaults to `null`)
- `min_upper` (defaults to `null`)
- `number` (defaults to `null`)
- `numeric` (defaults to `null`)
- `override_special` (defaults to `null`)
- `special` (defaults to `null`)
- `upper` (defaults to `null`)

## Computed outputs

Read these from the resource's `refAttr`:

- `id`
- `result`

## Example

```nix
random_string {
  name = "string";
  length = …;  # required
}
```
