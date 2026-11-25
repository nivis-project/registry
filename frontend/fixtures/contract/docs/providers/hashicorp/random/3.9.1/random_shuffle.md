# `random_shuffle`

Nivis Nix constructor for the `random_shuffle` resource (provider `random`). This reference is derived from the provider binary's own schema via `nivis gen` — it always matches the real provider.

## Constructor

```nix
{
  name,
  input,
  keepers ? null,
  result_count ? null,
  seed ? null,
  overrides ? {}
}
```

## Required arguments

- `input`

## Optional arguments

- `keepers` (defaults to `null`)
- `result_count` (defaults to `null`)
- `seed` (defaults to `null`)

## Computed outputs

Read these from the resource's `refAttr`:

- `id`
- `result`

## Example

```nix
random_shuffle {
  name = "shuffle";
  input = …;  # required
}
```
