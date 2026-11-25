---
# registry-a3ee
title: 'verify-contract: refuse null document lists'
status: completed
type: task
priority: normal
tags:
    - discovered
created_at: 2026-11-24T22:13:46Z
updated_at: 2026-11-25T21:08:45Z
---

Follow-up to registry-rw7l (empty-doc-lists). Once the contract has been regenerated with the fixed generator and republished by contract-publish.yml, scripts/verify-contract.sh should refuse a provider index.json whose docs.resources, datasources, functions or guides is null, so a regression fails the build instead of reaching the site. Not added together with the fix because the contract in S3 still held two null lists, and the gate would have failed the next site build.
