---
# registry-cnl1
title: 'Crawlable pages: prerendered HTML, metadata, sitemap'
status: completed
type: feature
priority: normal
created_at: 2026-11-20T12:31:17Z
updated_at: 2026-11-20T12:45:13Z
---

Phase 4 of the GitHub polish briefing. registry.nivis.tf serves an empty shell and routes with a HashRouter, so search engines and link previews see one empty page for ~12,000 routes. Switch to path URLs, prerender every route at build time with the same React components (renderToString with the data filled in), give each page a title, description, og tags with a PNG in the nivis.tf style and a canonical URL, publish sitemap.xml and robots.txt, redirect old #/ links, and align the header and footer with nivis.tf (Home, Docs, Star on GitHub). Approach agreed with Pim on 2026-11-20. Change: crawlable-pages.
