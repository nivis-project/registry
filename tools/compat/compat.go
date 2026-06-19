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

import "github.com/nivis-project/registry/tools/extract"

// NivisSupportedArches is the set Nivis intends to support, as "os/arch".
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

// E2EStatus is the honest end-to-end axis.
type E2EStatus string

const (
	E2EStatusNone     E2EStatus = "none"
	E2EStatusVerified E2EStatus = "verified"
)

// Record is the per-provider compatibility record embedded in the contract.
type Record struct {
	Address           string    `json:"address"`
	Tier              Tier      `json:"tier"`               // headline badge
	SchemaExtractable bool      `json:"schema_extractable"` // axis 1
	Protocols         []string  `json:"protocols"`          // axis 2
	Architectures     []string  `json:"architectures"`      // axis 3 (intersection)
	E2E               E2EStatus `json:"e2e"`                // axis 4
}

// TierFor returns the headline tier for a provider address.
func TierFor(address string, schemaExtractable bool) Tier {
	if E2EVerified[address] {
		return TierVerified
	}
	return TierByDesign
}

// E2EFor returns the honest e2e status: "verified" only for allowlisted
// providers, "none" otherwise. A provider is NEVER verified without an entry.
func E2EFor(address string) E2EStatus {
	if E2EVerified[address] {
		return E2EStatusVerified
	}
	return E2EStatusNone
}

// IntersectArches returns Nivis-supported ∩ published, preserving Nivis order.
func IntersectArches(published []string) []string {
	set := make(map[string]bool, len(published))
	for _, p := range published {
		set[p] = true
	}
	out := []string{}
	for _, a := range NivisSupportedArches {
		if set[a] {
			out = append(out, a)
		}
	}
	return out
}

// Compute builds the compat record from a provider's extraction metadata and
// whether a schema was successfully extracted. The architecture set is the
// intersection of Nivis-supported and the provider's published platforms; the
// tier and e2e axes never claim "verified" without an allowlist entry.
func Compute(meta extract.Metadata, schemaExtractable bool) Record {
	published := make([]string, 0, len(meta.Platforms))
	for _, p := range meta.Platforms {
		published = append(published, p.String())
	}
	protocols := meta.Protocols
	if protocols == nil {
		protocols = []string{}
	}
	return Record{
		Address:           meta.Address,
		Tier:              TierFor(meta.Address, schemaExtractable),
		SchemaExtractable: schemaExtractable,
		Protocols:         protocols,
		Architectures:     IntersectArches(published),
		E2E:               E2EFor(meta.Address),
	}
}
