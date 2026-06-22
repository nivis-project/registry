## ADDED Requirements

### Requirement: Verified binary resolution

The extraction tool SHALL obtain a provider binary by the nivis model — resolve the version via
the OpenTofu registry, download, and verify the SHA256 checksum BEFORE execution — reusing nivis's
`internal/registry` behavior rather than reimplementing verification.

#### Scenario: Checksum mismatch aborts extraction

- **WHEN** a downloaded provider archive fails SHA256 verification
- **THEN** the extraction for that provider aborts with an error
- **AND** the provider binary is never executed

### Requirement: Schema capture via nivis gen

The extraction tool SHALL run `nivis gen --provider <binary> --identity <id> --out <dir>` and
capture both the generated `.nix` constructors and a normalized schema JSON per provider version.

#### Scenario: Extraction produces constructors and schema

- **WHEN** extraction runs for `hashicorp/random`
- **THEN** `.nix` constructor files exist for each resource type
- **AND** a normalized schema JSON is written for the provider version

### Requirement: Resilient batch extraction

A failure extracting any single provider SHALL be skipped and logged, and MUST NOT hang or abort
the overall run.

#### Scenario: One bad provider does not stop the batch

- **WHEN** one provider in a batch fails to extract
- **THEN** the failure is logged with the provider id and reason
- **AND** the remaining providers are still extracted
