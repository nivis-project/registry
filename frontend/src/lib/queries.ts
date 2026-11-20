import type { QueryClient } from "@tanstack/react-query";
import { matchPath } from "react-router-dom";
import type { ModuleRef, ProviderRef } from "./contract";
import {
  fetchCatalog,
  fetchItemDoc,
  fetchModuleCatalog,
  fetchModuleVersion,
  fetchProviderVersion,
} from "./data";

export const routes = {
  home: "/",
  providers: "/providers",
  provider: "/providers/:namespace/:name/:version",
  resource: "/providers/:namespace/:name/:version/resources/:item",
  modules: "/modules",
  module: "/modules/:owner/:name/:version",
} as const;

export const catalogQuery = () => ({
  queryKey: ["catalog"],
  queryFn: fetchCatalog,
});

export const moduleCatalogQuery = () => ({
  queryKey: ["module-catalog"],
  queryFn: fetchModuleCatalog,
});

export const providerQuery = (ref: ProviderRef) => ({
  queryKey: ["provider", ref.namespace, ref.name, ref.version],
  queryFn: () => fetchProviderVersion(ref),
});

export const docQuery = (ref: ProviderRef, item: string) => ({
  queryKey: ["doc", ref.namespace, ref.name, ref.version, item],
  queryFn: () => fetchItemDoc(ref, item),
});

export const moduleQuery = (ref: ModuleRef) => ({
  queryKey: ["module", ref.owner, ref.name, ref.version],
  queryFn: () => fetchModuleVersion(ref),
});

export type RouteMatch =
  | { kind: "home" }
  | { kind: "providers" }
  | { kind: "provider"; ref: ProviderRef }
  | { kind: "resource"; ref: ProviderRef; item: string }
  | { kind: "modules" }
  | { kind: "module"; ref: ModuleRef }
  | { kind: "unknown" };

export function matchRoute(pathname: string): RouteMatch {
  const m = (pattern: string) => matchPath({ path: pattern, end: true }, pathname);
  let r;
  if (m(routes.home)) return { kind: "home" };
  if (m(routes.providers)) return { kind: "providers" };
  if ((r = m(routes.resource))) {
    const { namespace = "", name = "", version = "", item = "" } = r.params;
    return { kind: "resource", ref: { namespace, name, version }, item };
  }
  if ((r = m(routes.provider))) {
    const { namespace = "", name = "", version = "" } = r.params;
    return { kind: "provider", ref: { namespace, name, version } };
  }
  if (m(routes.modules)) return { kind: "modules" };
  if ((r = m(routes.module))) {
    const { owner = "", name = "", version = "" } = r.params;
    return { kind: "module", ref: { owner, name, version } };
  }
  return { kind: "unknown" };
}

export async function loadRoute(client: QueryClient, pathname: string): Promise<RouteMatch> {
  const route = matchRoute(pathname);
  switch (route.kind) {
    case "providers":
      await client.fetchQuery(catalogQuery());
      break;
    case "provider":
      await client.fetchQuery(providerQuery(route.ref));
      break;
    case "resource":
      await Promise.all([
        client.fetchQuery(docQuery(route.ref, route.item)),
        client.fetchQuery(providerQuery(route.ref)),
      ]);
      break;
    case "modules":
      await client.fetchQuery(moduleCatalogQuery());
      break;
    case "module":
      await client.fetchQuery(moduleQuery(route.ref));
      break;
  }
  return route;
}

export function fragmentToPath(hash: string): string | null {
  if (!hash.startsWith("#/")) return null;
  return hash.slice(1);
}
