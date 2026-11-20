---
# registry-rw7l
title: The generator writes null for an empty item list
status: todo
type: bug
tags:
    - discovered
created_at: 2026-11-20T12:41:13Z
updated_at: 2026-11-20T12:41:13Z
---

Found while prerendering (registry-cnl1): two providers in the contract, e.g. hashicorp/http 3.6.3, have docs.resources (and other lists) as null instead of []. contract.ts types them as arrays, so the provider page crashed on .length. The frontend now reads them with ?? [], but tools/generate should emit [] for an empty list (a Go nil slice marshals to null), and verify-contract.sh could refuse null.
