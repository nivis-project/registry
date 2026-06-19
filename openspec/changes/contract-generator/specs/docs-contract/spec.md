## ADDED Requirements

### Requirement: registry-ui-compatible contract tree

The generator SHALL emit a static contract tree at
`registry/docs/providers/{namespace}/{name}/{version}/index.json` that lists the provider's
resources, datasources, and functions in the shape consumed by the forked registry-ui SPA, with a
per-item document for each entry.

#### Scenario: Provider index lists all items

- **WHEN** the generator runs for an extracted provider version
- **THEN** `index.json` enumerates every resource, datasource, and function
- **AND** each listed item has a corresponding per-item document file

### Requirement: Compat badge embedded in the contract

Each provider `index.json` SHALL include the compatibility record produced by the compat tool.

#### Scenario: Badge data travels with the contract

- **WHEN** a provider's contract is generated
- **THEN** its `index.json` contains the compat tier, protocol(s), architectures, and e2e status
