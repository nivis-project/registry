---
# registry-fhl1
title: Deploy the site at registry.nivis.tf (Amplify + contract origin)
status: in-progress
type: task
priority: normal
created_at: 2026-09-08T10:57:09Z
updated_at: 2026-10-06T18:18:11Z
parent: registry-pao9
---

OpenSpec change: `registry-site-deploy` (openspec/changes/registry-site-deploy/).

Amplify serves the SPA (push-to-deploy on main); a scheduled GitHub Actions workflow publishes the
50-provider contract to S3+CloudFront; an Amplify 200-rewrite `/registry/<*>` joins them so the SPA
stays same-origin and unchanged.

Cross-repo: the Amplify app is a new nivis domain `stack/030_registry_site/` in ../infra; the
publication workflow is in this repo. Tasks are labelled [registry] / [infra] / [ops].

Blocked on three prerequisites (tasks 1.1-1.4): target AWS account, where nivis-aws-amplify-site
lives, and the Amplify GitHub App install on the nivis-project org.


## Progress 2026-10-06: the registry half is in, the account half is not

Commit `cc8060fb`. The OpenSpec change was rewritten against applied reality first:
the stack is `050_amplify_registry_nivis_tf` in the account repo (not `../infra`
`030_registry_site`), it inlines the module rather than composing it, and it sets no
rules and no build spec.

Landed here (12 of 30 tasks):

- `amplify.yml`: builds `frontend/`, publishes `frontend/dist`. The app had no build
  spec and this repo had no root `package.json`, so Amplify had nothing to detect.
- `scripts/verify-contract.sh`: refuses an absent, empty, malformed or incomplete
  contract. Exercised against seven cases; it rejects all six bad ones.
- `scripts/fetch-contract.sh`: syncs the bucket into `frontend/public/registry` and
  runs the gate.
- `.github/workflows/contract-publish.yml`: scheduled extraction, documents uploaded
  before catalogues, Amplify build triggered on success only. No-ops while
  `CONTRACT_BUCKET` or `AWS_ROLE_ARN` is unset.
- `REGISTRY-DESIGN.md` hosting section, including the ~150 MB trigger for switching
  to a second origin.
- `docs/runbook-deploy.md`.

Blocked, and why: 13 `[ops]` tasks are AWS console or live-site verification, and 4
`[account]` tasks are in the account repo and apply real infrastructure. Neither is
mine to do.

The next actionable thing is task 1.3: does the Amplify build service role, which
carries `AdministratorAccess-Amplify` rather than plain `AdministratorAccess`, reach
an S3 bucket in the account? That answer decides whether `050` needs any change at
all.
