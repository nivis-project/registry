## 1. Contract emitter

- [ ] 1.1 `tools/generate`: walk extraction output; emit `registry/docs/providers/{ns}/{name}/{version}/index.json`
- [ ] 1.2 Match registry-ui contract types (resources/datasources/functions lists + per-item docs)
- [ ] 1.3 Embed compat record (tier, protocol, archs, e2e) into each provider `index.json`
- [ ] 1.4 Test: index lists every item; each item has a document; badge present

## 2. Nix rendering

- [ ] 2.1 Parse `nivis gen` constructor header (computed outputs, nested-block shapes) + signature
- [ ] 2.2 Render per-item document body as the Nivis Nix reference (no HCL block)
- [ ] 2.3 Test: `aws_s3_bucket`-style doc shows constructor + outputs, contains no HCL `resource` block
- [ ] 2.4 Test: regeneration reflects schema changes across versions
