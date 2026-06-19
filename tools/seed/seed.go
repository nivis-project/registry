// Package seed derives and pins the provider seed manifest (seed.json) for the
// Nivis registry.
//
// Source of truth for popularity: the PUBLIC Terraform registry listing API,
//
//	https://registry.terraform.io/v1/providers?limit=N&offset=M
//
// which returns per-provider `tier` (official/partner/community) and `downloads`.
// NOTE: the `filter=` query parameter is INERT — it does not filter. Page through
// with limit/offset and sort client-side by (tier, downloads).
//
// The manifest is: top providers by downloads (favoring official/partner) UNION a
// fixed utility allowlist UNION Telmate/proxmox, capped at 50, ordered
// deterministically so seed.json is byte-stable across runs.
//
// See openspec/changes/extraction-pipeline/specs/seed-selection/spec.md.
package seed

// UtilityAllowlist is always included regardless of download rank: tiny, ubiquitous
// providers that appear in nearly every real Nivis config.
var UtilityAllowlist = []string{
	"hashicorp/random",
	"hashicorp/null",
	"hashicorp/local",
	"hashicorp/tls",
	"hashicorp/time",
	"hashicorp/http",
	"hashicorp/external",
	"hashicorp/cloudinit",
}

// MustInclude are explicit anchors the seed must always contain.
var MustInclude = []string{
	"hashicorp/aws",
	"hashicorp/azurerm",
	"hashicorp/google",
	"Telmate/proxmox", // most-popular proxmox provider (~16M downloads)
}

// MaxSeed caps the first registry at 50 providers.
const MaxSeed = 50

// TODO(extraction-pipeline / seed-derivation epic):
//   - Provider struct { Namespace, Name, Tier string; Downloads int }
//   - Fetch(client) ([]Provider, error): page registry.terraform.io/v1/providers
//   - Select([]Provider) []Provider: sort by tier+downloads, union allowlists, cap.
//   - Write(path string, []Provider) error: deterministic, byte-stable seed.json.
