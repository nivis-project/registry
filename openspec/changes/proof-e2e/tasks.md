## 1. Hermetic e2e

- [ ] 1.1 Build a nivis fake provider (`cmd/provider-*`) as a test fixture (no network/creds)
- [ ] 1.2 Run it through extract → contract → render; assert a rendered page exists
- [ ] 1.3 Wire this path into CI so it runs on every change (networking disabled)

## 2. Real-provider e2e

- [ ] 2.1 Run `hashicorp/random`, `hashicorp/null`, `Telmate/proxmox` through the full pipeline
- [ ] 2.2 Assert contract `index.json` + per-item docs + compat badge for each
- [ ] 2.3 Assert rendered resource pages show Nix constructors (no HCL block)

## 3. PAUSE gate

- [ ] 3.1 Confirm e2e green and site builds locally
- [ ] 3.2 Commit as `Pim Snel <post@pimsnel.com>` (no Claude trailer); push to `main`
- [ ] 3.3 STOP and report; do NOT scale to 50 or deploy; leave milestone `06` as `todo`
