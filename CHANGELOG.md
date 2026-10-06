# Changelog

All notable changes to the Nivis Registry are recorded here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
the project uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `nivis-project/hcloudimage` is in the catalogue. Providers that are published
  on the OpenTofu registry but never rank in the popularity API can now be
  catalogued by pinning them.
- `seed-pins.json` at the repo root lists every provider the seed always
  includes, with the reason for each. Adding one no longer means editing Go.

### Changed

- Pinning a provider grows the seed instead of evicting the lowest-ranked
  popular one. The cap is derived from the pin count plus a fixed fill of 38.
- `Telmate/proxmox` carries the reason `anchor` rather than `anchor:proxmox`;
  the detail moved to the pin's `note`.
