## Why

The PoC's purpose is to prove the whole chain works — extraction → contract → rendering — end to
end, with confidence and without scope creep. This change defines the e2e proof and the hard PAUSE
gate: the autonomous run stops here for human review before scaling to 50 providers or deploying.

## What Changes

- **Hermetic e2e**: build a nivis fake provider (`cmd/provider-*`, no network/creds), run it through
  the full pipeline, and assert a rendered page is produced — fully offline and deterministic.
- **Real-provider e2e**: run `hashicorp/random`, `hashicorp/null`, and `Telmate/proxmox` through the
  pipeline and assert contract + rendered pages + compat badges.
- **PAUSE gate**: when e2e is green and the site builds locally, commit (as Pim Snel) and push to
  `main`, then STOP and report. The run MUST NOT scale to 50 providers or deploy to S3/CloudFront;
  milestone `06` stays `todo`.

## Capabilities

### New Capabilities
- `e2e-proof`: Hermetic + real-provider end-to-end tests proving extraction→contract→rendering, gated by a hard PAUSE before scaling/deploying.

### Modified Capabilities

## Impact

- New e2e test suite; CI runs the hermetic path on every change.
- Defines the explicit stop point of the first autonomous run.
- Depends on `frontend-fork`.
