# `random_id`

Nivis Nix constructor for the `random_id` resource (provider `random`). This reference is derived from the provider binary's own schema via `nivis gen` — it always matches the real provider.

## Constructor

```nix
{ name, byte_length, keepers ? null, prefix ? null, overrides ? {} }
```

## Required arguments

- `byte_length`

## Optional arguments

- `keepers` (defaults to `null`)
- `prefix` (defaults to `null`)

## Computed outputs

Read these from the resource's `refAttr`:

- `b64_std`
- `b64_url`
- `dec`
- `hex`
- `id`

## Example

```nix
random_id {
  name = "id";
  byte_length = …;  # required
}
```
