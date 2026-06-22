# compat-tiers Specification

## Purpose
TBD - created by archiving change extraction-pipeline. Update Purpose after archive.
## Requirements
### Requirement: Per-provider compatibility tier

The compat tool SHALL compute a per-provider compatibility record from four independent axes:
schema-extractable (yes/no), protocol version(s) from metadata, architectures (Nivis-supported ∩
published `targets`), and an e2e tier of `none` or `verified` sourced from a hand-maintained allowlist.

#### Scenario: Badge reflects honest e2e status

- **WHEN** a provider has been schema-extracted but has no entry in the e2e allowlist
- **THEN** its compat record reports `compatible by design` with e2e tier `none`
- **AND** it is NOT labeled `verified`

#### Scenario: Architecture set is an intersection

- **WHEN** a provider publishes only `linux/amd64` and `darwin/*` targets
- **THEN** the compat record's architecture list excludes `linux/arm64`
- **AND** lists only architectures that are both Nivis-supported and published

