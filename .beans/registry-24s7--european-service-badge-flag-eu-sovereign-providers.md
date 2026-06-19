---
# registry-24s7
title: 'European-service badge: flag EU-sovereign providers in the contract + UI'
status: todo
type: task
created_at: 2026-06-19T16:21:19Z
updated_at: 2026-06-19T16:21:19Z
parent: registry-vk2i
priority: normal
---

> Scheduled: do this just before the weekend (per Pim, 2026-06-19).

## Why

The seed now pins the most popular **European-service** providers (Scaleway, OVH,
UpCloud, IONOS, STACKIT, cloudscale.ch, Aiven, Gandi — see the now-closed
`registry-otf0` and `EuropeanAllowlist` in `tools/seed/seed.go`). European /
EU-sovereign hosting is a real selection criterion for many users, so the
registry should make it **discoverable at a glance**: a dedicated "European
service" badge on provider pages and a filter/marker in the index.

This is a presentation feature on top of data we already have — distinct from the
compat badge (which is about whether a provider *works* with Nivis). A provider
can be both "compatible by design" and "European".

## What it needs

The "is this a European service" fact is editorial (sovereignty/HQ), not derivable
from the provider binary — so it lives in a small, explicit, hand-maintained list,
like the compat e2e allowlist.

1. **Source of truth.** A hand-maintained set of European-service provider
   addresses (plus optionally country code), e.g. in `tools/compat` (next to
   `E2EVerified`) or a new tiny `tools/region` package:
   `var EuropeanServices = map[string]string{ "scaleway/scaleway": "FR", "ovh/ovh": "FR", "UpCloudLtd/upcloud": "FI", "ionos-cloud/ionoscloud": "DE", "stackitcloud/stackit": "DE", "cloudscale-ch/cloudscale": "CH", "aiven/aiven": "FI", "go-gandi/gandi": "FR", "hetznercloud/hcloud": "DE", "exoscale/exoscale": "CH" }`
   (note: hetzner + exoscale are European too but reached the seed via popularity,
   so they are NOT in `EuropeanAllowlist` — include them in THIS list).
2. **Contract.** Add an optional field to the provider `index.json` — either on
   the existing `compat` record or a sibling, e.g.
   `"region": { "european": true, "country": "FR" }`. Keep it optional so
   non-European providers simply omit it. Update `tools/generate` to populate it
   from the source-of-truth list, and the TS type in
   `frontend/src/lib/contract.ts`.
3. **UI.** A small badge (e.g. 🇪🇺 / "European service") on `ProviderPage` and a
   marker on the `IndexPage` list. Consider a simple "European only" filter on the
   index (client-side; v1 has no search backend).
4. **Honesty.** Like the compat badge, only show the badge for providers actually
   in the list. Don't infer from namespace.

## Acceptance criteria

- [ ] A hand-maintained European-service list (address → country) exists and is
      unit-tested (non-empty; entries are valid addresses).
- [ ] The provider `index.json` carries the European/region marker for listed
      providers and omits it for others; the generator populates it; the contract
      TS type is updated.
- [ ] `ProviderPage` shows a "European service" badge (and the index marks them);
      a provider not in the list shows nothing.
- [ ] `go test ./...`, `nix flake check`, and the frontend test/build stay green.

## Notes

- The 8 `europe`-reason providers are already in `seed.json`; once the badge ships
  a `scripts/scale.sh` run renders them with the marker. No seed change needed.
- Keep this presentation-only — it does not change extraction or compat logic.

