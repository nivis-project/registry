import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { canonicalFor, fileFor, ORIGIN, publishedRoutes } from "./routes.mjs";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const dist = join(root, "dist");
const { render } = await import(pathToFileURL(join(root, "dist-server", "entry-server.js")).href);

const templateFile = join(root, "dist-server", "template.html");
if (!existsSync(templateFile)) writeFileSync(templateFile, readFileSync(join(dist, "index.html")));
const template = readFileSync(templateFile, "utf8");
const TITLE = "<title>Nivis Registry</title>";
const ROOT = '<div id="root"></div>';
if (!template.includes(TITLE) || !template.includes(ROOT)) {
  throw new Error("prerender: dist/index.html lacks the title or the empty root to fill");
}

const cache = new Map();
const fetcher = async (url) => {
  if (!url.startsWith("/registry/")) throw new Error(`prerender: unexpected fetch ${url}`);
  if (!cache.has(url)) cache.set(url, readFileSync(join(dist, url), "utf8"));
  return cache.get(url);
};

function page(rendered) {
  return template
    .replace(TITLE, rendered.headTags)
    .replace(ROOT, `<div id="root">${rendered.html}</div>`);
}

function write(file, html) {
  mkdirSync(dirname(file), { recursive: true });
  writeFileSync(file, html);
}

const routes = publishedRoutes(dist);
const started = Date.now();
const sitemap = [];
for (const route of routes) {
  const rendered = await render(route, fetcher);
  if (rendered.head.canonical !== canonicalFor(route)) {
    throw new Error(`prerender: ${route} has canonical ${rendered.head.canonical}`);
  }
  write(fileFor(dist, route), page(rendered));
  sitemap.push(rendered.head.canonical);
}

write(join(dist, "404.html"), page(await render("/404", fetcher)));

write(
  join(dist, "sitemap.xml"),
  `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n` +
    sitemap.map((u) => `  <url><loc>${u}</loc></url>`).join("\n") +
    `\n</urlset>\n`,
);
write(join(dist, "robots.txt"), `User-agent: *\nAllow: /\n\nSitemap: ${ORIGIN}/sitemap.xml\n`);

console.log(`prerender: ${routes.length} pages, 404.html, sitemap.xml, robots.txt in ${((Date.now() - started) / 1000).toFixed(1)}s`);
