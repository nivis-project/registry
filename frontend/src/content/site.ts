// Page copy, kept out of the components so that changing a sentence is not a
// code change. Imported at build time, never fetched: HomePage has to render
// when the contract is missing or stale.
//
// The honesty wording lives here too. Those sentences are specified behavior
// (see the compat-tiers, module-extraction and registry-ui-fork specs), so they
// are gathered in one place where a reviewer can read them together rather than
// scattered across six components.

export const site = {
  name: "Nivis Registry",
  nameLead: "Nivis",
  nameTail: "Registry",
  homeUrl: "https://nivis.tf",
  repoUrl: "https://github.com/nivis-project/registry",
  footer: "Apache-2.0 · provider data from the OpenTofu registry",
} as const;

export const home = {
  // Deliberately not the mockup's "Every OpenTofu provider, documented for
  // Nix". The catalogue holds a pinned seed, not every provider, and this
  // project does not present a partial catalogue as a complete one.
  title: "OpenTofu providers, documented as Nix.",
  lead: "Nivis runs unmodified provider binaries, so every provider is compatible by design. The registry shows each resource as the Nix constructor you would write.",
  providers: {
    title: "Providers",
    body: "Browse the catalogue. Each provider lists its resources and data sources as typed Nix constructors, with a compatibility badge.",
    cta: "Browse providers",
  },
  modules: {
    title: "Modules",
    body: "Composable nivis modules: whole pieces of infrastructure as a single Nix import. What each one creates is derived by evaluating it.",
    cta: "Browse modules",
  },
  badge: {
    title: "What the badge means",
    byDesign: {
      term: "Compatible by design",
      body: "The pipeline downloaded the provider, verified it, and read its schema; its protocol version and published architectures are recorded. This is the default, and it is a machine-checked claim about the provider's interface. It does not mean anyone has run it.",
    },
    verified: {
      term: "E2E verified",
      body: "A separate, smaller, hand-maintained set: providers actually driven end to end against real infrastructure. A provider is never shown as verified without an explicit entry on that list.",
    },
  },
} as const;

export const providers = {
  title: "Providers",
  empty: "No providers are catalogued yet.",
  filterLabel: "Filter providers",
  filterPlaceholder: "Filter by address",
  noMatch: "No provider matches that filter.",
  generatedNote:
    "The reference for every provider is generated from the provider's own schema, not copied from its documentation.",
} as const;

export const providerPage = {
  useThis: "Use this provider",
  compatTitle: "Compatibility",
  tabs: { resources: "Resources", datasources: "Data sources", functions: "Functions" },
  filterLabel: "Filter items",
  filterPlaceholder: "Filter by name",
  noMatch: "No item matches that filter.",
} as const;

export const modules = {
  title: "Modules",
  lead: "Composable nivis modules: whole pieces of infrastructure as a single Nix import. What each one creates is derived by evaluating the module, not copied from its documentation.",
  empty: "No modules are catalogued yet.",
  notDerivedChip: "structure not derived",
} as const;

export const modulePage = {
  creates: "Creates",
  reads: "Reads",
  outputs: "Outputs",
  exposes: "Also exposes",
  exposesBody: "Further values this module returns, which another module can consume.",
  configuration: "Configuration",
  required: "Required",
  optional: "Optional",
  none: "None.",
  // Specified wording. Do not reword without changing the spec.
  uncatalogued: "provider not catalogued here",
  inferredCfg:
    "Inferred by reading the module's source. A nivis module does not declare its configuration, so this list is a best effort, not a declaration.",
  notDerivedTitle: "Structure not derived",
  notDerivedBody:
    "This module could not be evaluated, so the registry cannot say what it creates. This is not a claim that it creates nothing.",
  readmeTitle: "From the module's README",
  readmeNote: "Written by the module's authors, not derived from it.",
} as const;

export const states = {
  loading: "Loading…",
  retry: "Try again",
  errorTitle: "Could not load this page",
  notFoundTitle: "Not found",
  notFoundBody: "That page does not exist in this registry.",
  backHome: "Go to the front page",
} as const;
