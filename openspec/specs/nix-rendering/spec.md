# nix-rendering Specification

## Purpose
TBD - created by archiving change contract-generator. Update Purpose after archive.
## Requirements
### Requirement: Render Nix constructors, not HCL

The generator SHALL render each `nivis gen` constructor — its doc-comment header (computed outputs,
nested-block shapes) and typed signature — into the per-item document body, presenting the Nivis Nix
form. It MUST NOT require or emit HCL syntax for the reference body.

The rendered constructor **signature** SHALL be formatted for readability and adapt to a maximum
width: a signature within the width stays on a single line; a signature exceeding it SHALL break to
**one argument per line**, two-space indented, with the closing brace on its own line. Required
arguments are rendered without a default (`name`, `arg`); optional arguments keep their `? null`
default; the final argument carries no trailing comma.

#### Scenario: Resource document shows the Nix constructor

- **WHEN** the per-item document for `aws_s3_bucket` is generated
- **THEN** it shows the Nix constructor signature with required/optional attributes and types
- **AND** it lists computed outputs accessible via `refAttr`
- **AND** it contains no HCL `resource "…" {}` block as the reference

#### Scenario: Wide signature wraps to one argument per line

- **WHEN** a resource has enough arguments that the one-line signature would exceed the maximum width
- **THEN** the rendered signature places each argument on its own indented line
- **AND** required arguments appear without `? null` and optional arguments keep `? null`
- **AND** the last argument has no trailing comma and the closing brace is on its own line

#### Scenario: Short signature stays on one line

- **WHEN** a resource's one-line signature is within the maximum width
- **THEN** the rendered signature is a single line (e.g. `{ name, triggers ? null, overrides ? {} }`)

### Requirement: Faithful to generated output

The rendered reference SHALL be derived from the actual `nivis gen` output for that provider
version, so it always matches the provider binary's real schema.

#### Scenario: Rendering matches extraction

- **WHEN** a provider's schema changes between versions
- **THEN** the regenerated documents reflect the new attributes and outputs for that version

