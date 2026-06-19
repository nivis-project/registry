## 1. Full-catalog run

- [x] 1.1 `scripts/scale.sh`: extract every provider in `seed.json` (`tools/extract -seed seed.json`)
  then generate the full contract (`tools/generate`), resiliently (skip+log single failures)
- [x] 1.2 Emit a coverage report: per-provider extracted (constructor count) vs. skipped (reason),
  plus a summary count
- [x] 1.3 Run it; confirm the hyperscalers (aws/azurerm/google) and proxmox are among the extracted

## 2. Verify at scale

- [x] 2.1 Confirm the full contract generates (`registry/docs/providers/...` + `catalog.json`)
- [x] 2.2 Build the static site against the full contract (`pnpm build`) — green
- [x] 2.3 Sample-eval generated constructors as valid Nix (no duplicate-formal regressions at scale)

## 3. Guard rails

- [x] 3.1 No deploy to S3/CloudFront (that is the separate, human-approved epic)
- [x] 3.2 No seed change (the 50 are already pinned)
- [x] 3.3 `go test ./...` + `nix flake check` stay green
