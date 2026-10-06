# Changelog

All notable changes to the Nivis Registry are recorded here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
the project uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- The registry has the Nivis visual identity: one indigo hue, self-hosted type,
  a generated mark in place of the emoji, and light and dark themes that follow
  your system by default and remember an override.

### Added

- Pages now have a state for loading, for a failed fetch with a retry, for an
  empty catalogue or filter, and for an unknown address.
- Modules are browsable. `/modules` lists them and each has a page showing what
  it creates, what it outputs and what it needs configured, with every resource
  linking to the provider documenting it.
- Nivis modules are catalogued alongside providers. Each module's resources,
  data sources and outputs are derived by evaluating the module itself, so the
  listing is machine-checked rather than copied from its documentation.
- Every resource a module creates links to the provider page documenting it.
- Module pages state what was derived and what was not: whether the structure
  evaluated, whether the configuration surface was inferred from source, and how
  many resource types the registry could resolve.
- A front page at the site root, explaining what the registry is and what the
  two compatibility tiers mean, instead of dropping a visitor straight into the
  provider list.
- A header navigation bar on every page, marking the section you are in.
- `nivis-project/hcloudimage` is in the catalogue. Providers that are published
  on the OpenTofu registry but never rank in the popularity API can now be
  catalogued by pinning them.
- `seed-pins.json` at the repo root lists every provider the seed always
  includes, with the reason for each. Adding one no longer means editing Go.

### Fixed

- The pipeline scripts now catalogue modules. They generated the contract
  without ever running the module extractor, so a full run produced an empty
  module list while reporting success.

### Changed

- The provider catalogue moved from the site root to `/providers`, so it is a
  section with its own address rather than the whole site.
- Pinning a provider grows the seed instead of evicting the lowest-ranked
  popular one. The cap is derived from the pin count plus a fixed fill of 38.
- `Telmate/proxmox` carries the reason `anchor` rather than `anchor:proxmox`;
  the detail moved to the pin's `note`.
