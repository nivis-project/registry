## Why

The PoC proved the pipeline end-to-end on a 6-provider proof set and PAUSEd at the gate. The nivis
fixes (0.4.2 schema-without-configure, 0.4.3 reserved-name collision) removed the blockers that kept
the hyperscalers and credential-requiring providers out. This change is the first post-gate step of
milestone `06`: run the full pipeline over **all 50 pinned seed providers** and emit the complete
static contract, proving the system holds at the real catalog size.

**Scope guard:** this change scales **extraction + contract generation only**. It does **NOT** deploy
to S3/CloudFront (that remains a separate, human-approved step in the `Scale to 50 + deploy` epic),
and it does NOT change the seed (the 50 are already pinned in `seed.json`).

## What Changes

- A repeatable **full-catalog run**: extract every provider in `seed.json` (via `tools/extract`'s
  resilient `Batch` + the `-seed` flag), then generate the full contract (`tools/generate`
  `GenerateAll`). No tool code changes are required — this exercises the existing pipeline at scale.
- A **scale script** (`scripts/scale.sh`) that runs the full-catalog extract → generate and reports
  a coverage summary (how many of the 50 succeeded, which were skipped and why).
- A persisted **coverage report** so the outcome is auditable: which providers extracted, their
  constructor counts, and skip reasons for any that failed.
- Confirm the **site still builds** against the full contract (`pnpm build`) and that generated
  constructors remain valid Nix at scale.

## Capabilities

### New Capabilities
- `full-catalog-extraction`: Running the extraction + contract pipeline over the entire pinned seed
  (50 providers) resiliently, with an auditable coverage report, without deploying.

### Modified Capabilities

## Impact

- New files: `scripts/scale.sh`; a generated coverage report (build artifact, gitignored).
- Provider binaries execute in-sandbox via `nivis gen` (verify-before-execute, as today). A single
  provider's failure is skipped + logged — never aborts the run.
- No deploy. No seed change. No frontend code change (it already renders the full contract).
