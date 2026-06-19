---
# registry-otf0
title: make a list of the top 10 most populair providers for european services. E.g. hetzner, mistral, etc
status: completed
type: task
priority: normal
created_at: 2026-06-19T14:26:17Z
updated_at: 2026-06-19T16:21:10Z
---

 These should be added to the starting list

## Resolution (2026-06-19)

Added an `EuropeanAllowlist` to `tools/seed/seed.go` (always pinned, reason `europe`) and regenerated
`seed.json` (now 58 providers, `MaxSeed` 50→58):

- scaleway/scaleway (FR), ovh/ovh (FR), UpCloudLtd/upcloud (FI), ionos-cloud/ionoscloud (DE),
  stackitcloud/stackit (DE), cloudscale-ch/cloudscale (CH), aiven/aiven (FI), go-gandi/gandi (FR).

All verified extractable on the OpenTofu registry. hetznercloud/hcloud (DE) and exoscale/exoscale (CH)
are European too but already rank in the popularity list, so they need no explicit pin. (Mistral has
no Terraform provider — it's an LLM API, not infra — so it was not added.)

A dedicated **"European service" badge** is tracked separately in its own bean.
