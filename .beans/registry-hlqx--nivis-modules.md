---
# registry-hlqx
title: nivis modules
status: in-progress
type: task
priority: normal
created_at: 2026-10-06T13:46:01Z
updated_at: 2026-10-06T16:50:50Z
---

The power of nix is having flakes and the module system. We have developed two modules already:

https://github.com/wearetechnative/nivis-aws-amplify-site
https://github.com/wearetechnative/nivis-aws-form-action

and we want to create a lot more.

The registry should have it's own listing of available modules and later a submission method.

terraform modules are not supported but we could get inspiration from regisrty.terraform and registry.opentofu how to present nivis modules.


## Progress: backend shipped (2026-10-06)

OpenSpec change `module-pipeline` (archived `2026-10-06-module-pipeline`), commit `9efe7e15`.
Bean stays open: the listing exists in the contract but there is no page yet.

Done:

- `module-pins.json` at the repo root, the curated module list.
- `tools/module`: three passes per module. Structure by evaluating the module with
  `cfg = throw` (Nix laziness means the throw is never forced, because `mkResource`
  derives identity from coordinates). Configuration surface by scanning the source.
  Provider links by exact lookup against the generated provider contract.
- `tools/version`: shared stable-over-prerelease ordering, now used by both provider
  versions and module tags.
- `tools/generate`: `registry/docs/modules/{owner}/{name}/{version}/index.json` plus
  `registry/modules.json`.
- Real run against both pinned modules, matching the spike exactly:
  amplify-site 0.1.0 (5 resources, 1 data source, 4 outputs, 6/6 linked),
  form-action 0.1.0 (9 resources, 0 data sources, 3 outputs, 9/9 linked, exposes
  `apiEndpointRef`).

Still open for this bean:

- The `/modules` page and the Modules nav entry. They ship together, so the nav never
  carries a dead link (requirement in the `site-navigation` spec).
- A submission method.

Worth knowing:

- Module flake refs use `git+https://...?rev=<peeled tag commit>`, NOT the GitHub API.
  The API path returns 403 under the `wearetechnative` org policy on fine-grained
  tokens older than 366 days, which will otherwise bite any local run.
- An annotated tag's plain ref names the tag object, not the commit. The peeled
  `^{}` ref is the one to use; the other fetches the wrong tree.
- Descriptions stay underived. That `hmacParam` must be an SSM SecureString is in the
  README and nowhere else, so the contract carries the README as a separate field.
