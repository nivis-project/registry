## ADDED Requirements

### Requirement: Full-catalog extraction over the pinned seed

The pipeline SHALL extract every provider in `seed.json` and generate the contract for each, reusing
the existing resilient batch (a single provider's failure is skipped and logged, never aborting the
run) and without deploying anywhere.

#### Scenario: All seed providers are attempted

- **WHEN** the scale run executes against `seed.json`
- **THEN** every provider address in the seed is attempted
- **AND** a provider that fails extraction is recorded with its reason and does not abort the run
- **AND** no artifacts are deployed to S3/CloudFront

### Requirement: Auditable coverage report

The scale run SHALL produce a coverage report enumerating, for each seed provider, whether it was
extracted (and its constructor count) or skipped (and why).

#### Scenario: Coverage is reported

- **WHEN** the scale run completes
- **THEN** a summary reports the count of extracted vs. skipped providers
- **AND** each skipped provider is listed with its skip reason

### Requirement: Contract and site build at scale

The generated contract for the full catalog SHALL build into the static site, and the generated Nix
constructors SHALL remain valid.

#### Scenario: Site builds against the full contract

- **WHEN** the full contract is generated and the frontend is built
- **THEN** the static site build succeeds with every extracted provider present in the catalog
- **AND** a sampled set of generated constructors evaluates as valid Nix
