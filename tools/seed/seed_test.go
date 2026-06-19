package seed

import (
	"strings"
	"testing"
)

// TestAnchorsPresent guards the invariants the seed must always satisfy.
func TestAnchorsPresent(t *testing.T) {
	if len(MustInclude) == 0 {
		t.Fatal("MustInclude must not be empty")
	}
	foundProxmox := false
	for _, p := range MustInclude {
		if p == "Telmate/proxmox" {
			foundProxmox = true
		}
	}
	if !foundProxmox {
		t.Error("seed must always include Telmate/proxmox (the most-popular proxmox provider)")
	}
	if MaxSeed != 50 {
		t.Errorf("MaxSeed = %d, want 50", MaxSeed)
	}
}

func TestUtilityAllowlistNonEmpty(t *testing.T) {
	if len(UtilityAllowlist) == 0 {
		t.Fatal("UtilityAllowlist must not be empty")
	}
}

// fixtureRows is a small deterministic stand-in for the popularity API: a couple
// hyperscalers, a couple utilities, proxmox, plus filler community providers.
func fixtureRows() []Provider {
	return []Provider{
		{Namespace: "hashicorp", Name: "aws", Tier: "official", Downloads: 6_500_000_000},
		{Namespace: "hashicorp", Name: "google", Tier: "official", Downloads: 2_000_000_000},
		{Namespace: "hashicorp", Name: "azurerm", Tier: "official", Downloads: 1_800_000_000},
		{Namespace: "hashicorp", Name: "random", Tier: "official", Downloads: 900_000_000},
		{Namespace: "hashicorp", Name: "null", Tier: "official", Downloads: 800_000_000},
		{Namespace: "Telmate", Name: "proxmox", Tier: "community", Downloads: 16_000_000},
		{Namespace: "datadog", Name: "datadog", Tier: "partner", Downloads: 500_000_000},
		{Namespace: "cloudflare", Name: "cloudflare", Tier: "partner", Downloads: 400_000_000},
		{Namespace: "someone", Name: "obscure", Tier: "community", Downloads: 1_000},
	}
}

func addrSet(entries []SeedEntry) map[string]SeedEntry {
	m := make(map[string]SeedEntry, len(entries))
	for _, e := range entries {
		m[e.Address] = e
	}
	return m
}

// TestSelectIncludesAnchorsAndUtilities matches the seed-selection spec:
// hyperscalers, proxmox, and the utility allowlist must all be present.
func TestSelectIncludesAnchorsAndUtilities(t *testing.T) {
	got := addrSet(Select(fixtureRows()))
	for _, want := range []string{
		"hashicorp/aws", "hashicorp/azurerm", "hashicorp/google", "Telmate/proxmox",
		"hashicorp/random", "hashicorp/null", "hashicorp/local", "hashicorp/tls",
		"hashicorp/time", "hashicorp/http",
	} {
		if _, ok := got[want]; !ok {
			t.Errorf("seed missing required provider %q", want)
		}
	}
}

// TestSelectAnchorsEvenWhenAbsentUpstream ensures anchors/utilities are pinned
// even if the popularity fetch did not return them (e.g. local/external are not
// in the fixture rows at all).
func TestSelectAnchorsEvenWhenAbsentUpstream(t *testing.T) {
	got := addrSet(Select(fixtureRows()))
	if _, ok := got["hashicorp/external"]; !ok {
		t.Error("utility hashicorp/external must be pinned even when absent from fetched rows")
	}
}

// TestSelectDeterministicOrder asserts the order is anchors → utilities →
// popular, so seed.json is byte-stable. Anchors lead in declared order.
func TestSelectDeterministicOrder(t *testing.T) {
	entries := Select(fixtureRows())
	wantPrefix := []string{
		"hashicorp/aws", "hashicorp/azurerm", "hashicorp/google", "Telmate/proxmox",
	}
	for i, w := range wantPrefix {
		if entries[i].Address != w {
			t.Errorf("entries[%d] = %q, want anchor %q", i, entries[i].Address, w)
		}
	}
	// Running twice yields the identical sequence.
	again := Select(fixtureRows())
	if len(entries) != len(again) {
		t.Fatalf("length not stable: %d vs %d", len(entries), len(again))
	}
	for i := range entries {
		if entries[i].Address != again[i].Address {
			t.Errorf("order not stable at %d: %q vs %q", i, entries[i].Address, again[i].Address)
		}
	}
}

// TestSelectReasonsAnnotated checks the manifest is self-documenting.
func TestSelectReasonsAnnotated(t *testing.T) {
	got := addrSet(Select(fixtureRows()))
	if got["Telmate/proxmox"].Reason != "anchor:proxmox" {
		t.Errorf("proxmox reason = %q, want anchor:proxmox", got["Telmate/proxmox"].Reason)
	}
	if got["hashicorp/tls"].Reason != "utility" {
		t.Errorf("tls reason = %q, want utility", got["hashicorp/tls"].Reason)
	}
	if got["datadog/datadog"].Reason != "popular" {
		t.Errorf("datadog reason = %q, want popular", got["datadog/datadog"].Reason)
	}
}

// TestSelectCapsAtMaxSeed ensures the manifest never exceeds the cap.
func TestSelectCapsAtMaxSeed(t *testing.T) {
	// Build more than MaxSeed community rows.
	rows := fixtureRows()
	for i := 0; i < MaxSeed*2; i++ {
		rows = append(rows, Provider{
			Namespace: "filler",
			Name:      "p" + string(rune('a'+i%26)) + string(rune('a'+i/26)),
			Tier:      "community",
			Downloads: int64(MaxSeed*2 - i),
		})
	}
	entries := Select(rows)
	if len(entries) > MaxSeed {
		t.Errorf("Select returned %d entries, want <= %d", len(entries), MaxSeed)
	}
}

func TestParseAddress(t *testing.T) {
	ns, name, err := ParseAddress("Telmate/proxmox")
	if err != nil || ns != "Telmate" || name != "proxmox" {
		t.Errorf("ParseAddress(Telmate/proxmox) = %q,%q,%v", ns, name, err)
	}
	if _, _, err := ParseAddress("nope"); err == nil {
		t.Error("ParseAddress(nope) should error")
	}
	if _, _, err := ParseAddress("/x"); err == nil {
		t.Error("ParseAddress(/x) should error")
	}
}

func TestParseAddressRejectsEmpty(t *testing.T) {
	if _, _, err := ParseAddress(""); err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("ParseAddress(empty) err = %v, want invalid", err)
	}
}
