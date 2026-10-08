// Contract types — mirrored from opentofu/registry-ui's
// backend/internal/server/openapi.yml (ProviderVersion / ProviderDocs /
// ProviderDocItem), extended with the Nivis `compat` record. Keeping these in
// sync with the generator (tools/generate) is what makes the SPA data-agnostic:
// it consumes a JSON contract, not Nivis or OpenTofu concepts.

export interface DocItem {
  name: string;
  title: string;
  description?: string;
  subcategory?: string;
}

export interface ProviderDocs {
  resources: DocItem[];
  datasources: DocItem[];
  functions: DocItem[];
  guides: DocItem[];
}

export type CompatTier = "compatible by design" | "e2e verified";
export type E2EStatus = "none" | "verified";

// CompatRecord is the Nivis-specific badge embedded in each provider index.
export interface CompatRecord {
  address: string;
  tier: CompatTier;
  schema_extractable: boolean;
  protocols: string[];
  architectures: string[];
  e2e: E2EStatus;
}

// ProviderVersion is the index.json shape (registry-ui ProviderVersion + compat).
export interface ProviderVersion {
  id: string; // version number
  published?: string;
  docs: ProviderDocs;
  compat: CompatRecord;
  avatar?: string;
}

// ProviderRef identifies a provider version within the contract tree.
export interface ProviderRef {
  namespace: string;
  name: string;
  version: string;
}

// --- modules ---------------------------------------------------------------
// Mirrored from tools/generate/modules.go. A nivis module is a flake output,
// not a binary, so its structure is DERIVED by evaluating it rather than read
// from a schema call. The record says how much of the page is machine-checked.

// ModuleCoord is one resource or data source a module declares. `docs` is the
// provider version documenting its type, absent when the registry does not
// catalogue that provider.
export interface ModuleCoord {
  provider: string;
  type: string;
  name: string;
  docs?: string;
}

// ModuleCfgKey is one configuration path a module reads. Nested paths arrive
// in full ("ses.from"), and a path under an optional ancestor is itself
// optional.
export interface ModuleCfgKey {
  path: string;
  required: boolean;
}

export type CfgSource = "scanned" | "declared";

// ModuleRecord is the module counterpart of CompatRecord: what was derived and
// what was not.
export interface ModuleRecord {
  structure_extractable: boolean;
  cfg_source: CfgSource;
  providers_resolved: number;
  providers_total: number;
  failure_reason?: string;
}

export interface ModuleVersion {
  id: string; // version
  owner: string;
  name: string;
  tag?: string;
  rev?: string;
  description?: string;
  resources: ModuleCoord[];
  data_sources: ModuleCoord[];
  outputs: string[];
  composition?: string[];
  cfg: ModuleCfgKey[];
  record: ModuleRecord;
  avatar?: string;
  readme?: string;
}

export interface ModuleCatalogEntry {
  owner: string;
  name: string;
  version: string;
  description?: string;
  derived: boolean;
  // Contract-relative reference to the owner's brand avatar. Absent when none
  // could be obtained, so a consumer never renders a broken image.
  avatar?: string;
}

// ModuleRef identifies a module version within the contract tree.
export interface ModuleRef {
  owner: string;
  name: string;
  version: string;
}
