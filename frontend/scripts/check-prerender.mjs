import { existsSync, readFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { canonicalFor, fileFor, ORIGIN, publishedRoutes } from "./routes.mjs";

const dist = resolve(process.argv[2] ?? join(dirname(fileURLToPath(import.meta.url)), "..", "dist"));
const errors = [];
const fail = (msg) => {
  if (errors.length < 25) errors.push(msg);
  else if (errors.length === 25) errors.push("...more");
};

const routes = publishedRoutes(dist);
const providers = routes.filter((r) => /^\/providers\/[^/]+\/[^/]+\/[^/]+$/.test(r));
if (providers.length === 0) fail("the contract lists no providers; there is nothing to check");

const between = (html, a, b) => {
  const i = html.indexOf(a);
  if (i < 0) return null;
  const j = html.indexOf(b, i + a.length);
  return j < 0 ? null : html.slice(i + a.length, j);
};

for (const route of routes) {
  const file = fileFor(dist, route);
  if (!existsSync(file)) {
    fail(`${route}: no ${file}`);
    continue;
  }
  const html = readFileSync(file, "utf8");
  const body = between(html, '<div id="root">', "</body>") ?? "";
  if (body.trim().startsWith("</div>") || body.length < 200) fail(`${route}: the page has no rendered content`);
  const title = between(html, "<title>", "</title>");
  if (!title || !title.endsWith("Nivis Registry")) fail(`${route}: title is ${JSON.stringify(title)}`);
  if (!/<meta name="description" content="[^"]{40,}">/.test(html)) fail(`${route}: no meta description`);
  const canonical = canonicalFor(route);
  if (!html.includes(`<link rel="canonical" href="${canonical}">`)) fail(`${route}: canonical is not ${canonical}`);
  if (!html.includes(`<meta property="og:url" content="${canonical}">`)) fail(`${route}: og:url is not ${canonical}`);
  for (const tag of [
    `<meta property="og:image" content="${ORIGIN}/social.png">`,
    `<meta property="og:image:width" content="1200">`,
    `<meta property="og:image:height" content="630">`,
    `<meta property="og:title" content="`,
    `<meta property="og:description" content="`,
  ]) {
    if (!html.includes(tag)) fail(`${route}: missing ${tag}`);
  }
  const resource = route.match(/\/resources\/([^/]+)$/);
  if (resource && !body.includes(resource[1])) fail(`${route}: the content does not name ${resource[1]}`);
  const provider = route.match(/^\/providers\/([^/]+)\/([^/]+)\/[^/]+$/);
  if (provider && !body.includes(`${provider[1]}/${provider[2]}`)) fail(`${route}: the content does not name ${provider[1]}/${provider[2]}`);
}

const sitemapFile = join(dist, "sitemap.xml");
if (!existsSync(sitemapFile)) fail("no sitemap.xml");
else {
  const listed = [...readFileSync(sitemapFile, "utf8").matchAll(/<loc>([^<]+)<\/loc>/g)].map((m) => m[1]);
  const want = new Set(routes.map(canonicalFor));
  const got = new Set(listed);
  if (listed.length !== got.size) fail("sitemap.xml lists a URL twice");
  for (const u of want) if (!got.has(u)) fail(`sitemap.xml lacks ${u}`);
  for (const u of got) if (!want.has(u)) fail(`sitemap.xml lists ${u}, which is not a published page`);
}

const robotsFile = join(dist, "robots.txt");
if (!existsSync(robotsFile)) fail("no robots.txt");
else {
  const robots = readFileSync(robotsFile, "utf8");
  if (!robots.includes(`Sitemap: ${ORIGIN}/sitemap.xml`)) fail("robots.txt does not name the sitemap");
  if (/^Disallow:\s*\/\s*$/m.test(robots)) fail("robots.txt disallows the whole site");
}

const notFound = join(dist, "404.html");
if (!existsSync(notFound)) fail("no 404.html");
else if (!readFileSync(notFound, "utf8").includes('<meta name="robots" content="noindex">')) fail("404.html can be indexed");

const png = join(dist, "social.png");
if (!existsSync(png)) fail("no social.png");
else {
  const b = readFileSync(png);
  const signature = b.subarray(0, 8).equals(Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]));
  if (!signature) fail("social.png is not a PNG");
  else if (b.readUInt32BE(16) !== 1200 || b.readUInt32BE(20) !== 630) fail(`social.png is ${b.readUInt32BE(16)}x${b.readUInt32BE(20)}, not 1200x630`);
}

if (errors.length) {
  for (const e of errors) console.error(`check-prerender: FAIL ${e}`);
  process.exit(1);
}
console.log(`check-prerender: ok, ${routes.length} pages with content and head, sitemap.xml, robots.txt, 404.html, social.png`);
