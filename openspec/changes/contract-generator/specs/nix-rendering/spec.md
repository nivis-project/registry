## ADDED Requirements

### Requirement: Render Nix constructors, not HCL

The generator SHALL render each `nivis gen` constructor — its doc-comment header (computed outputs,
nested-block shapes) and typed signature — into the per-item document body, presenting the Nivis Nix
form. It MUST NOT require or emit HCL syntax for the reference body.

#### Scenario: Resource document shows the Nix constructor

- **WHEN** the per-item document for `aws_s3_bucket` is generated
- **THEN** it shows the Nix constructor signature with required/optional attributes and types
- **AND** it lists computed outputs accessible via `refAttr`
- **AND** it contains no HCL `resource "…" {}` block as the reference

### Requirement: Faithful to generated output

The rendered reference SHALL be derived from the actual `nivis gen` output for that provider
version, so it always matches the provider binary's real schema.

#### Scenario: Rendering matches extraction

- **WHEN** a provider's schema changes between versions
- **THEN** the regenerated documents reflect the new attributes and outputs for that version
