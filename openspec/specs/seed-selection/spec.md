# seed-selection Specification

## Purpose
TBD - created by archiving change extraction-pipeline. Update Purpose after archive.
## Requirements
### Requirement: Derived, pinned seed manifest

The seed tool SHALL derive the provider list from the public popularity API
(`https://registry.terraform.io/v1/providers`) by paging and sorting on `tier` and `downloads`
(the `filter=` query parameter is inert and MUST NOT be relied upon), union a fixed utility
allowlist and `Telmate/proxmox`, and write a pinned `seed.json`.

#### Scenario: Seed resolves the expected anchors

- **WHEN** the seed tool runs against the popularity API
- **THEN** `seed.json` includes the hyperscalers (`hashicorp/aws`, `hashicorp/azurerm`, `hashicorp/google`)
- **AND** includes `Telmate/proxmox`
- **AND** includes the utility allowlist (`random`, `null`, `local`, `tls`, `time`, `http`)

#### Scenario: Seed is reproducible and pinned

- **WHEN** the seed tool runs twice with the same upstream data
- **THEN** the resulting `seed.json` is byte-identical and ordered deterministically

