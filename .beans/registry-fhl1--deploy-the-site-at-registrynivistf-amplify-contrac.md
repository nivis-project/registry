---
# registry-fhl1
title: Deploy the site at registry.nivis.tf (Amplify + contract origin)
status: todo
type: task
priority: normal
created_at: 2026-09-08T10:57:09Z
updated_at: 2026-09-08T10:57:09Z
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
