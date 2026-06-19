## 1. Hermetic e2e

- [x] 1.1 Build a nivis fake provider (`cmd/provider-beta`) as a test fixture (no network/creds)
- [x] 1.2 Run it through extract → contract → render; assert a rendered page exists (`tools/e2e`
  `TestHermeticPipeline`: Nix block present, no HCL block, declares the required `from` arg)
- [x] 1.3 Wire this path into CI so it runs on every change (the `e2e-hermetic` job builds the nivis
  CLI + matching source from the pinned input; the pipeline itself uses no provider network/creds)

## 2. Real-provider e2e

- [x] 2.1 Run `hashicorp/random`, `hashicorp/null`, and `hashicorp/tls` through the full pipeline
  (`tools/e2e` `TestRealProviderPipeline`, gated by `NIVIS_REGISTRY_NET_E2E=1`). NOTE: `Telmate/proxmox`
  is excluded — nivis gen v0.4.0 configures before fetching schema and proxmox/azurerm/google reject an
  all-null configure (fix in flight upstream; re-pin the nivis input after its patch release).
- [x] 2.2 Assert contract `index.json` + per-item docs + compat badge for each (proven; the proof
  script `scripts/proof.sh` emits the contract for all three)
- [x] 2.3 Assert rendered resource pages show Nix constructors (no HCL block)

## 3. PAUSE gate

- [x] 3.1 Confirm e2e green and site builds locally (`nix flake check`, `go test ./...`, hermetic +
  real e2e, frontend `pnpm test`/`build`, and `scripts/proof.sh` all green)
- [ ] 3.2 Commit as `Pim Snel <post@pimsnel.com>` (no Claude trailer); push to `main`
- [ ] 3.3 STOP and report; do NOT scale to 50 or deploy; leave milestone `06` as `todo`
