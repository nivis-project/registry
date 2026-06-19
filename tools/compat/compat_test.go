package compat

import (
	"reflect"
	"testing"

	"github.com/nivis-project/registry/tools/extract"
)

// A provider not in the allowlist must never be "verified".
func TestUnverifiedIsByDesign(t *testing.T) {
	if got := TierFor("hashicorp/random", true); got != TierByDesign {
		t.Errorf("TierFor(unverified) = %q, want %q", got, TierByDesign)
	}
	if got := E2EFor("hashicorp/random"); got != E2EStatusNone {
		t.Errorf("E2EFor(unverified) = %q, want %q", got, E2EStatusNone)
	}
}

// Architectures are an intersection: linux/arm64 must drop when not published,
// and windows/amd64 must drop because Nivis does not support it.
func TestIntersectArches(t *testing.T) {
	published := []string{"linux/amd64", "darwin/amd64", "darwin/arm64", "windows/amd64"}
	got := IntersectArches(published)
	want := []string{"linux/amd64", "darwin/amd64", "darwin/arm64"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("IntersectArches = %v, want %v (linux/arm64 must drop, windows is not Nivis-supported)", got, want)
	}
}

// TestComputeByDesign is the spec scenario "Badge reflects honest e2e status":
// a schema-extracted provider with no allowlist entry is "compatible by design"
// with e2e "none", never "verified".
func TestComputeByDesign(t *testing.T) {
	meta := extract.Metadata{
		Address:   "hashicorp/random",
		Protocols: []string{"5.0"},
		Platforms: []extract.Platform{
			{OS: "linux", Arch: "amd64"},
			{OS: "linux", Arch: "arm64"},
			{OS: "darwin", Arch: "arm64"},
		},
	}
	rec := Compute(meta, true)
	if rec.Tier != TierByDesign {
		t.Errorf("Tier = %q, want %q", rec.Tier, TierByDesign)
	}
	if rec.E2E != E2EStatusNone {
		t.Errorf("E2E = %q, want %q", rec.E2E, E2EStatusNone)
	}
	if !rec.SchemaExtractable {
		t.Error("SchemaExtractable should be true")
	}
	if !reflect.DeepEqual(rec.Protocols, []string{"5.0"}) {
		t.Errorf("Protocols = %v, want [5.0]", rec.Protocols)
	}
	want := []string{"linux/amd64", "linux/arm64", "darwin/arm64"}
	if !reflect.DeepEqual(rec.Architectures, want) {
		t.Errorf("Architectures = %v, want %v", rec.Architectures, want)
	}
}

// TestComputeArchesIntersectionScenario is the spec scenario "Architecture set
// is an intersection": a provider publishing only linux/amd64 + darwin/* yields
// a record that excludes linux/arm64.
func TestComputeArchesIntersectionScenario(t *testing.T) {
	meta := extract.Metadata{
		Address: "acme/thing",
		Platforms: []extract.Platform{
			{OS: "linux", Arch: "amd64"},
			{OS: "darwin", Arch: "amd64"},
			{OS: "darwin", Arch: "arm64"},
		},
	}
	rec := Compute(meta, true)
	for _, a := range rec.Architectures {
		if a == "linux/arm64" {
			t.Error("linux/arm64 must NOT appear when not published")
		}
	}
	want := []string{"linux/amd64", "darwin/amd64", "darwin/arm64"}
	if !reflect.DeepEqual(rec.Architectures, want) {
		t.Errorf("Architectures = %v, want %v", rec.Architectures, want)
	}
}

// TestComputeVerifiedNeedsAllowlist proves "verified" requires an allowlist
// entry: temporarily allowlist a provider and confirm both tier and e2e flip.
func TestComputeVerifiedNeedsAllowlist(t *testing.T) {
	const addr = "acme/proven"
	E2EVerified[addr] = true
	t.Cleanup(func() { delete(E2EVerified, addr) })

	rec := Compute(extract.Metadata{Address: addr}, true)
	if rec.Tier != TierVerified {
		t.Errorf("Tier = %q, want %q for allowlisted provider", rec.Tier, TierVerified)
	}
	if rec.E2E != E2EStatusVerified {
		t.Errorf("E2E = %q, want %q for allowlisted provider", rec.E2E, E2EStatusVerified)
	}

	// A provider NOT on the allowlist stays by-design even with a schema.
	other := Compute(extract.Metadata{Address: "acme/unproven"}, true)
	if other.Tier == TierVerified || other.E2E == E2EStatusVerified {
		t.Error("non-allowlisted provider must never be verified")
	}
}

// TestComputeEmptyMetadataSafe ensures nil protocols/platforms serialize as
// empty slices, not null, so the contract is stable.
func TestComputeEmptyMetadataSafe(t *testing.T) {
	rec := Compute(extract.Metadata{Address: "x/y"}, false)
	if rec.Protocols == nil {
		t.Error("Protocols should be [] not nil")
	}
	if rec.Architectures == nil {
		t.Error("Architectures should be [] not nil")
	}
	if rec.SchemaExtractable {
		t.Error("SchemaExtractable should be false")
	}
}
