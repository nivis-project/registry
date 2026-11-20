import { readFileSync } from "node:fs";
import { join } from "node:path";

export function publishedRoutes(dist) {
  const read = (p) => JSON.parse(readFileSync(join(dist, p), "utf8"));
  const routes = ["/", "/providers", "/modules"];
  for (const p of read("registry/catalog.json")) {
    const base = `/providers/${p.namespace}/${p.name}/${p.version}`;
    routes.push(base);
    const index = read(`registry/docs/providers/${p.namespace}/${p.name}/${p.version}/index.json`);
    const docs = index.docs ?? {};
    for (const kind of ["resources", "datasources", "functions"]) {
      for (const item of docs[kind] ?? []) routes.push(`${base}/resources/${item.name}`);
    }
  }
  for (const m of read("registry/modules.json")) {
    routes.push(`/modules/${m.owner}/${m.name}/${m.version}`);
  }
  return routes;
}

export function fileFor(dist, route) {
  return route === "/" ? join(dist, "index.html") : join(dist, route, "index.html");
}

export const ORIGIN = "https://registry.nivis.tf";

export function canonicalFor(route) {
  return route === "/" ? `${ORIGIN}/` : `${ORIGIN}${route}/`;
}
