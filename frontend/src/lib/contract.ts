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
}

// ProviderRef identifies a provider version within the contract tree.
export interface ProviderRef {
  namespace: string;
  name: string;
  version: string;
}
