# registry-ui-fork Specification

## Purpose
TBD - created by archiving change frontend-fork. Update Purpose after archive.
## Requirements
### Requirement: Forked SPA wired to the static contract

The forked registry-ui SPA SHALL read provider/resource data from the Nivis static contract
(local files in dev, S3/CloudFront in production) and SHALL build to static assets with no live
backend required for v1.

#### Scenario: SPA renders from static files

- **WHEN** the SPA is built and served against the generated static contract
- **THEN** provider and resource pages render from the static `index.json` + per-item documents
- **AND** no live API server is required

### Requirement: Resource pages render Nix constructors

The resource-page rendering SHALL display the Nivis Nix constructor (signature, attributes,
computed outputs) from the contract, not scraped HCL markdown.

#### Scenario: A resource page shows Nix

- **WHEN** a user opens a resource page
- **THEN** the page shows the Nix constructor and outputs
- **AND** does not present an HCL `resource "…" {}` block as the reference

### Requirement: Compat badge visible

Provider pages SHALL surface the compatibility badge (tier, protocol, architectures, e2e status)
from the contract.

#### Scenario: Badge is shown on a provider page

- **WHEN** a user opens a provider page
- **THEN** the compat tier and e2e status are visibly rendered
- **AND** an unverified provider is shown as `compatible by design`, not `verified`

### Requirement: Upstream license recorded

Before vendoring, the upstream registry-ui license SHALL be verified and recorded in the repo.

#### Scenario: License is documented

- **WHEN** the fork is added under `frontend/`
- **THEN** the upstream license is recorded and compatible with redistribution

