// Data layer — points at the Nivis static contract. In dev and in the static
// v1 deploy the contract files live under <base>/registry/docs/..., served as
// plain files (no live backend). The later API (API Gateway + Lambda) serves the
// SAME paths, so this layer does not change when hosting evolves.

import type {
  ModuleCatalogEntry,
  ModuleRef,
  ModuleVersion,
  ProviderRef,
  ProviderVersion,
} from "./contract";

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
  // tier is the COMPAT tier ("compatible by design"), not the publisher's
  // standing upstream. The two are different axes with similar vocabulary.
  tier: string;
  avatar?: string;
  // Why this provider is catalogued, verbatim from the seed: anchor, utility,
  // europe, curated or popular.
  reason?: string;
  // Upstream standing (official, partner). Absent when upstream reports none.
  publisher?: string;
}

export async function fetchCatalog(): Promise<CatalogEntry[]> {
  const res = await fetch("registry/catalog.json");
  if (!res.ok) throw new Error(`GET catalog -> ${res.status}`);
  return (await res.json()) as CatalogEntry[];
}

// --- modules ---------------------------------------------------------------

const moduleBase = "registry/docs/modules";

// avatarURL resolves a contract-relative avatar reference ("avatars/ovh.png")
// against the contract root. The contract stores it relative to itself so it
// does not bake in where the contract is mounted; every other path goes through
// this layer for the same reason, and an <img src> straight from the contract
// would resolve against the DOCUMENT instead and 404.
export function avatarURL(ref?: string): string | undefined {
  return ref ? `registry/${ref}` : undefined;
}

function moduleIndexURL(ref: ModuleRef): string {
  return `${moduleBase}/${ref.owner}/${ref.name}/${ref.version}/index.json`;
}

// fetchModuleCatalog lists the modules present in the static contract.
export async function fetchModuleCatalog(): Promise<ModuleCatalogEntry[]> {
  const res = await fetch("registry/modules.json");
  if (!res.ok) throw new Error(`GET modules catalog -> ${res.status}`);
  return (await res.json()) as ModuleCatalogEntry[];
}

// fetchModuleVersion loads one module version's derived structure.
export async function fetchModuleVersion(
  ref: ModuleRef,
): Promise<ModuleVersion> {
  const body = await getText(moduleIndexURL(ref));
  return JSON.parse(body) as ModuleVersion;
}
