# `random_bytes`

Nivis Nix constructor for the `random_bytes` resource (provider `random`). This reference is derived from the provider binary's own schema via `nivis gen` — it always matches the real provider.

## Constructor

```nix
{ name, length, keepers ? null, overrides ? {} }
```

## Required arguments

- `length`

## Optional arguments

- `keepers` (defaults to `null`)

## Computed outputs

Read these from the resource's `refAttr`:

- `base64`
- `hex`

## Example

```nix
random_bytes {
  name = "bytes";
  length = …;  # required
}
```
