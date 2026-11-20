import type { QueryClient } from "@tanstack/react-query";
import type { CatalogEntry } from "./data";
import type { ModuleCatalogEntry, ModuleVersion, ProviderVersion } from "./contract";
import { home, meta, modules as modulesCopy, providers as providersCopy, site, states } from "../content/site";
import {
  catalogQuery,
  moduleCatalogQuery,
  moduleQuery,
  providerQuery,
  type RouteMatch,
} from "./queries";

export interface PageHead {
  title: string;
  description: string;
  canonical: string;
  indexable: boolean;
}

export function pageTitle(title?: string): string {
  return title ? `${title} · ${site.name}` : site.name;
}

function canonical(pathname: string): string {
  const path = pathname === "/" ? "/" : `${pathname.replace(/\/+$/, "")}/`;
  return `${site.origin}${path}`;
}

export function headFor(route: RouteMatch, pathname: string, client: QueryClient): PageHead {
  const at = (title: string | undefined, description: string): PageHead => ({
    title: pageTitle(title),
    description,
    canonical: canonical(pathname),
    indexable: true,
  });
  switch (route.kind) {
    case "home":
      return at(undefined, home.lead);
    case "providers": {
      const list = client.getQueryData<CatalogEntry[]>(catalogQuery().queryKey) ?? [];
      return at(providersCopy.title, meta.providers(list.length));
    }
    case "provider": {
      const { namespace, name, version } = route.ref;
      const d = client.getQueryData<ProviderVersion>(providerQuery(route.ref).queryKey);
      return at(`${namespace}/${name}`, meta.provider(`${namespace}/${name}`, version, d?.docs.resources?.length ?? 0));
    }
    case "resource": {
      const { namespace, name, version } = route.ref;
      return at(route.item, meta.resource(route.item, `${namespace}/${name}`, version));
    }
    case "modules": {
      const list = client.getQueryData<ModuleCatalogEntry[]>(moduleCatalogQuery().queryKey) ?? [];
      return at(modulesCopy.title, meta.modules(list.length));
    }
    case "module": {
      const { owner, name } = route.ref;
      const d = client.getQueryData<ModuleVersion>(moduleQuery(route.ref).queryKey);
      return at(`${owner}/${name}`, meta.module(`${owner}/${name}`, d?.description));
    }
    case "unknown":
      return { ...at(states.notFoundTitle, home.lead), indexable: false };
  }
}

function escape(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

export function headTags(head: PageHead): string {
  const e = escape;
  const tags = [
    `<title>${e(head.title)}</title>`,
    `<meta name="description" content="${e(head.description)}">`,
    head.indexable
      ? `<link rel="canonical" href="${e(head.canonical)}">`
      : `<meta name="robots" content="noindex">`,
    `<meta property="og:type" content="website">`,
    `<meta property="og:site_name" content="${e(site.name)}">`,
    `<meta property="og:title" content="${e(head.title)}">`,
    `<meta property="og:description" content="${e(head.description)}">`,
    head.indexable ? `<meta property="og:url" content="${e(head.canonical)}">` : "",
    `<meta property="og:image" content="${e(site.socialImage)}">`,
    `<meta property="og:image:type" content="image/png">`,
    `<meta property="og:image:width" content="1200">`,
    `<meta property="og:image:height" content="630">`,
    `<meta name="twitter:card" content="summary_large_image">`,
  ];
  return tags.filter(Boolean).join("\n    ");
}
