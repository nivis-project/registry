## 1. Contract emitter

- [x] 1.1 `tools/generate`: walk extraction output; emit `registry/docs/providers/{ns}/{name}/{version}/index.json`
- [x] 1.2 Match registry-ui contract types (resources/datasources/functions lists + per-item docs)
- [x] 1.3 Embed compat record (tier, protocol, archs, e2e) into each provider `index.json`
- [x] 1.4 Test: index lists every item; each item has a document; badge present

## 2. Nix rendering

- [x] 2.1 Parse `nivis gen` constructor header (computed outputs, nested-block shapes) + signature
- [x] 2.2 Render per-item document body as the Nivis Nix reference (no HCL block)
- [x] 2.3 Test: `aws_s3_bucket`-style doc shows constructor + outputs, contains no HCL `resource` block
- [x] 2.4 Test: regeneration reflects schema changes across versions
