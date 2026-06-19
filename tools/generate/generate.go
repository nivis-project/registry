// Package generate emits the static JSON contract consumed by the forked
// registry-ui SPA, populated with schema-derived Nix constructors.
//
// Contract shape mirrors opentofu/registry-ui (backend/internal/server/openapi.yml):
//
//	registry/docs/providers/{ns}/{name}/{version}/index.json   # lists resources/datasources/functions + compat badge
//	registry/docs/providers/{ns}/{name}/{version}/<item>...     # per-item document
//
// The Nivis difference: the per-item document body is the RENDERED `nivis gen`
// constructor (signature + types + computed outputs), NOT scraped HCL markdown.
// Each provider index.json also carries the compat record from tools/compat.
//
// See openspec/changes/contract-generator/specs/{docs-contract,nix-rendering}/spec.md.
package generate

// ContractPath returns the index.json path for a provider version, relative to
// the contract root.
func ContractPath(namespace, name, version string) string {
	return "registry/docs/providers/" + namespace + "/" + name + "/" + version + "/index.json"
}

// TODO(contract-generator epic):
//   - Walk extraction output (tools/extract) per provider version.
//   - Emit index.json (item lists + embedded compat record).
//   - RenderConstructor(nixConstructor) -> per-item document body (no HCL block).
