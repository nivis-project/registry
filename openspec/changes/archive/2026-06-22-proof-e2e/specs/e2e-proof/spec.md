## ADDED Requirements

### Requirement: Hermetic end-to-end proof

A hermetic e2e test SHALL build a nivis fake provider and run it through the full pipeline
(extract → contract → render) with no network access and no credentials, producing a rendered page.

#### Scenario: Offline pipeline produces a page

- **WHEN** the hermetic e2e runs in CI with networking disabled
- **THEN** a fake provider is extracted, a contract is emitted, and a page is rendered
- **AND** the test passes deterministically

### Requirement: Real-provider end-to-end proof

An e2e test SHALL run `hashicorp/random`, `hashicorp/null`, and `Telmate/proxmox` through the full
pipeline and assert contract files, rendered pages, and compat badges for each.

#### Scenario: Three real providers render

- **WHEN** the real-provider e2e runs
- **THEN** each provider has contract `index.json`, per-item documents, and a compat badge
- **AND** each rendered resource page shows Nix constructors

### Requirement: Hard PAUSE gate before scaling

When e2e is green and the site builds locally, the run SHALL commit (author `Pim Snel`) and push to
`main`, then STOP and report. It MUST NOT extract beyond the proof set or deploy to S3/CloudFront,
and milestone `06` MUST remain `todo`.

#### Scenario: Run stops at the gate

- **WHEN** the proof e2e is green and the site builds
- **THEN** the run commits and pushes to `main`
- **AND** it does not scale to 50 providers
- **AND** it does not deploy to S3/CloudFront
- **AND** it reports completion of the proof and pauses for human review
