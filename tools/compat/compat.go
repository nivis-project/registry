// Package compat computes the per-provider compatibility record (the "badge").
//
// Four INDEPENDENT axes (be honest — never conflate "should work" with "proven"):
//  1. SchemaExtractable: did the pipeline get a schema from the binary?
//  2. Protocols: from upstream metadata ("5.0" / "6.0"). Nivis handles both; 6.0 untested.
//  3. Architectures: Nivis-supported ∩ provider-published `targets` (an intersection).
//  4. E2E: "none" | "verified", from a small hand-maintained allowlist.
//
// Default badge = "compatible by design" (axes 1–3). "verified" is ONLY for
// providers in the e2e allowlist.
//
// See openspec/changes/extraction-pipeline/specs/compat-tiers/spec.md.
package compat

// NivisSupportedArches is the set Nivis intends to support.
var NivisSupportedArches = []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64"}

// E2EVerified is the hand-maintained allowlist of providers proven end-to-end.
// Keep this honest and small; everything else is "compatible by design".
var E2EVerified = map[string]bool{
	// "Telmate/proxmox": true,  // add once a real e2e proves it
}

// Tier is the headline label shown on a provider page.
type Tier string

const (
	TierByDesign Tier = "compatible by design"
	TierVerified Tier = "e2e verified"
)

// TierFor returns the headline tier for a provider address.
func TierFor(address string, schemaExtractable bool) Tier {
	if E2EVerified[address] {
		return TierVerified
	}
	return TierByDesign
}

// IntersectArches returns Nivis-supported ∩ published, preserving Nivis order.
func IntersectArches(published []string) []string {
	set := make(map[string]bool, len(published))
	for _, p := range published {
		set[p] = true
	}
	var out []string
	for _, a := range NivisSupportedArches {
		if set[a] {
			out = append(out, a)
		}
	}
	return out
}

// TODO(extraction-pipeline / compat-tiers epic):
//   - Record struct { Tier; Protocols []string; Arches []string; E2E string }
//   - Compute(address, metadata, schema) Record
