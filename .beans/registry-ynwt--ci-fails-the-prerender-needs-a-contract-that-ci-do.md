---
# registry-ynwt
title: 'CI fails: the prerender needs a contract that CI does not have'
status: completed
type: bug
priority: normal
tags:
    - discovered
created_at: 2026-11-25T21:11:18Z
updated_at: 2026-11-25T21:14:23Z
---

Since crawlable-pages (registry-cnl1, shipped 2026-11-20) the CI frontend job fails: pnpm build now prerenders every route and reads frontend/dist/registry/catalog.json, but CI has no contract (it lives in a private S3 bucket that only the Amplify build reads). The live site was never affected, because Amplify fetches the contract before building. Fix: a small fixture contract in the repo that CI puts in place before the build, so CI also runs the prerender and check-prerender. Change: ci-fixture-contract.
