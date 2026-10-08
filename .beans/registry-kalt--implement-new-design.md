---
# registry-kalt
title: implement new design
status: in-progress
type: feature
priority: normal
created_at: 2026-10-06T20:40:49Z
updated_at: 2026-10-08T18:03:14Z
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


## Progress: step 2 of 4 shipped (2026-10-08)

`catalog-enrichment`, archived `2026-10-08-catalog-enrichment`, commit `eb7bc670`.
23/23 tasks, 59 frontend tests (was 48), Go overall 75.8%.

Owner avatars plus the seed fields, combined into one change because both land in
`catalog.json` and `tools/generate`.

- `tools/avatar` fetches each owner's logo when the contract is built, normalises it to
  PNG and writes it to `registry/avatars/<owner>.png`. 36 owners, 212 KB. Served from our
  own origin: hot-linking would put a request to GitHub on every card of every page view.
- Five of the 36 come back as JPEG despite the `.png` in the URL, so normalisation is not
  optional. Standard library only, no new Go dependency.
- A failed fetch is skipped, never fatal, and leaves no field, so the SPA cannot render a
  broken image.
- The `--plate` token is light in BOTH themes. Seven logos are dark ink on transparency
  and would vanish on a dark page, `wearetechnative` and `nivis-project` among them.
- Avatars appear on cards, list rows, and beside the title on provider and module detail
  pages at 36px.
- No preferential treatment: same size and plate for every owner.

The override seam is in place and empty (`avatar.Overrides`, consulted before fetching),
so a hand-picked logo later is a data change, not a rework.

Remaining: `frontend-search-catalog`, `search-index`, `blueprints`.
