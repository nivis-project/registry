---
# registry-jjrt
title: layout breaks with certain resources
status: completed
type: task
priority: normal
created_at: 2026-06-19T16:31:40Z
updated_at: 2026-06-22T14:12:00Z
---

e.g. http://localhost:5173/#/providers/scaleway/scaleway/2.76.0/resources/scaleway_autoscaling_instance_group

th screen gets too wide

## Resolution (2026-06-22)

Fixed via OpenSpec change `render-signature-wrapping`. The rendered Nix constructor signature is now
width-aware (`tools/generate/render.go` `signature`): wide signatures break to one argument per line
instead of a single over-wide line, so resources with many (often mandatory) arguments — e.g.
scaleway_autoscaling_instance_group — no longer overflow the screen. The `nix-rendering` spec records
the requirement so the formatting stays adapted to a max width going forward.
