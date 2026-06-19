// Data layer — points at the Nivis static contract. In dev and in the static
// v1 deploy the contract files live under <base>/registry/docs/..., served as
// plain files (no live backend). The later API (API Gateway + Lambda) serves the
// SAME paths, so this layer does not change when hosting evolves.

import type { ProviderRef, ProviderVersion } from "./contract";

// contractBase is where the contract tree is served from. Relative so the SPA
// works under any mount point (CloudFront subpath, file://-ish previews, etc.).
const contractBase = "registry/docs/providers";

function providerIndexURL(ref: ProviderRef): string {
  return `${contractBase}/${ref.namespace}/${ref.name}/${ref.version}/index.json`;
}

function itemDocURL(ref: ProviderRef, item: string): string {
  return `${contractBase}/${ref.namespace}/${ref.name}/${ref.version}/${item}.md`;
}

async function getText(url: string): Promise<string> {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`GET ${url} -> ${res.status}`);
  return res.text();
}

// fetchProviderVersion loads a provider version's index.json (item lists + compat).
export async function fetchProviderVersion(
  ref: ProviderRef,
): Promise<ProviderVersion> {
  const body = await getText(providerIndexURL(ref));
  return JSON.parse(body) as ProviderVersion;
}

// fetchItemDoc loads the Nix-rendered document body (Markdown) for one item.
export async function fetchItemDoc(
  ref: ProviderRef,
  item: string,
): Promise<string> {
  return getText(itemDocURL(ref, item));
}

// catalog.json lists the providers present in the static contract so the index
// page can link them without a live search backend (v1). It is generated
// alongside the contract.
export interface CatalogEntry extends ProviderRef {
  tier: string;
}

export async function fetchCatalog(): Promise<CatalogEntry[]> {
  const res = await fetch("registry/catalog.json");
  if (!res.ok) throw new Error(`GET catalog -> ${res.status}`);
  return (await res.json()) as CatalogEntry[];
}
