---
# registry-kalt
title: implement new design
status: in-progress
type: feature
priority: normal
created_at: 2026-10-06T20:40:49Z
updated_at: 2026-10-06T21:11:22Z
---

Implement as much as you can of this new design. Always seperate content from style. Use yaml or markdown for content. Use placeholder for non existing features like the search engine or the blueprints menu item.


## Progress: step 1 of 4 shipped (2026-10-06)

`frontend-design-system`, archived `2026-10-06-frontend-design-system`, commit `aeae9198`.
39/39 tasks. 48 frontend tests green (was 20).

- `src/styles/tokens.css` is the only place a colour exists, and
  `src/styles/tokens.test.ts` fails the build if a second place appears. All 132
  default-palette utilities are gone.
- Hind and IBM Plex Mono vendored as Latin-subset woff2 (95 KB, six faces), with their
  OFL notices. No runtime request to a font service.
- `Mark.tsx` draws the mark from `r(θ) = a + b·cos(kθ)`, computed once at module level,
  and the favicon uses the same geometry so the tab cannot drift from the header.
- Light and dark follow the system, can be overridden, persist, and are applied by an
  inline script before first paint.
- All six pages restyled, plus the compatibility panel, `MarkdownDoc`, and a Nix
  tokenizer for the code blocks.
- The states the mockups omit: skeletons, a retryable error, empty catalogue and filter,
  and a 404 route. Per-route document titles.

Deliberately not taken from the mockups, each recorded in the proposal:

- The headline. "Every OpenTofu provider, documented for Nix" describes 59 providers out
  of several thousand.
- The provider card's description and tier chip. `catalog.json` carries neither, and the
  upstream registry's description for `hashicorp/aws` is the string
  `terraform-provider-aws`. `catalog-enrichment` adds the fields.
- The header search field, which would open nothing until the overlay exists.

Measured: contrast passes 4.5:1 on all 13 token pairs in both themes. Bundle js
267.4 -> 285.8 KB, css 11.4 -> 16.1 KB, fonts +95 KB. Verified serving from a deep
subpath with relative assets.

Remaining: `catalog-enrichment`, `frontend-search-catalog`, `search-index`, `blueprints`.
Blueprints is blocked on the maintainer: where they are published, the clone command, and
which parts of the data model are derived versus authored.
