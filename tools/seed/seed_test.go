package seed

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// fixturePins mirrors the shape of seed-pins.json: anchors, utilities, europe,
// then curated, each carrying its own reason.
func fixturePins() []Pin {
	return []Pin{
		{Address: "hashicorp/aws", Reason: "anchor"},
		{Address: "hashicorp/azurerm", Reason: "anchor"},
		{Address: "hashicorp/google", Reason: "anchor"},
		{Address: "Telmate/proxmox", Reason: "anchor", Note: "most-popular proxmox provider"},
		{Address: "hashicorp/random", Reason: "utility"},
		{Address: "hashicorp/null", Reason: "utility"},
		{Address: "hashicorp/external", Reason: "utility"},
		{Address: "scaleway/scaleway", Reason: "europe", Note: "Scaleway (FR)"},
		{Address: "nivis-project/hcloudimage", Reason: "curated"},
	}
}

// repoPins loads the committed seed-pins.json so the real file is covered by the
// same invariants as the fixture.
func repoPins(t *testing.T) []Pin {
	t.Helper()
	pins, err := LoadPins(filepath.Join("..", "..", DefaultPinsPath))
	if err != nil {
		t.Fatalf("LoadPins(repo): %v", err)
	}
	return pins
}

// fixtureRows is a small deterministic stand-in for the popularity API.
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

// fillerRows builds n extra deterministic community rows on top of the fixture.
func fillerRows(n int) []Provider {
	rows := fixtureRows()
	for i := 0; i < n; i++ {
		rows = append(rows, Provider{
			Namespace: "filler",
			Name:      "p" + string(rune('a'+i%26)) + string(rune('a'+i/26)),
			Tier:      "community",
			Downloads: int64(n - i),
		})
	}
	return rows
}

func addrSet(entries []SeedEntry) map[string]SeedEntry {
	m := make(map[string]SeedEntry, len(entries))
	for _, e := range entries {
		m[e.Address] = e
	}
	return m
}

// TestRepoPinsValid guards the committed pin file: it parses, has no duplicates,
// and still carries the providers that must always be catalogued.
func TestRepoPinsValid(t *testing.T) {
	pins := repoPins(t)
	if len(pins) == 0 {
		t.Fatal("seed-pins.json must not be empty")
	}
	byAddr := map[string]Pin{}
	for _, p := range pins {
		if _, dup := byAddr[p.Address]; dup {
			t.Errorf("duplicate pin %q", p.Address)
		}
		byAddr[p.Address] = p
	}
	for _, want := range []string{
		"hashicorp/aws", "hashicorp/azurerm", "hashicorp/google", "Telmate/proxmox",
		"hashicorp/random", "hashicorp/null", "hashicorp/local", "hashicorp/tls",
		"hashicorp/time", "hashicorp/http", "scaleway/scaleway", "ovh/ovh",
		"nivis-project/hcloudimage",
	} {
		if _, ok := byAddr[want]; !ok {
			t.Errorf("seed-pins.json missing %q", want)
		}
	}
	if got := byAddr["nivis-project/hcloudimage"].Reason; got != "curated" {
		t.Errorf("hcloudimage reason = %q, want curated", got)
	}
}

// TestLoadPinsRejectsBadEntries: the pin file is hand-edited, so a bad address
// or a missing reason must fail loudly rather than yield a silent empty reason.
func TestLoadPinsRejectsBadEntries(t *testing.T) {
	for name, body := range map[string]string{
		"bad address":    `{"pinned":[{"address":"nope","reason":"anchor"}]}`,
		"empty reason":   `{"pinned":[{"address":"a/b","reason":""}]}`,
		"malformed json": `{"pinned":[`,
	} {
		path := filepath.Join(t.TempDir(), "pins.json")
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadPins(path); err == nil {
			t.Errorf("%s: LoadPins should have failed", name)
		}
	}
	if _, err := LoadPins(filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Error("LoadPins on a missing file should fail")
	}
}

// TestPinsNeverEvictPopular is the invariant the derived cap exists for.
func TestPinsNeverEvictPopular(t *testing.T) {
	rows := fillerRows(PopularFill * 3)
	base := Select(fixturePins(), rows)
	grown := Select(append(fixturePins(), Pin{Address: "extra/pin", Reason: "curated"}), rows)

	if len(grown) != len(base)+1 {
		t.Errorf("adding a pin changed the manifest by %d entries, want 1", len(grown)-len(base))
	}
	after := addrSet(grown)
	for _, e := range base {
		if e.Reason != "popular" {
			continue
		}
		if _, ok := after[e.Address]; !ok {
			t.Errorf("popular provider %q was evicted by adding a pin", e.Address)
		}
	}
}

// TestPopularFillIsStable: the popularity-ranked count does not move with the
// pin count.
func TestPopularFillIsStable(t *testing.T) {
	rows := fillerRows(PopularFill * 3)
	for _, pins := range [][]Pin{
		fixturePins(),
		append(fixturePins(), Pin{Address: "extra/one", Reason: "curated"}),
		{{Address: "hashicorp/aws", Reason: "anchor"}},
	} {
		got := 0
		for _, e := range Select(pins, rows) {
			if e.Reason == "popular" {
				got++
			}
		}
		if got != PopularFill {
			t.Errorf("with %d pins: %d popular entries, want %d", len(pins), got, PopularFill)
		}
	}
}

// TestSelectReasonComesFromTheFile: selection must not special-case any address.
func TestSelectReasonComesFromTheFile(t *testing.T) {
	pins := []Pin{
		{Address: "Telmate/proxmox", Reason: "whatever-the-file-says"},
		{Address: "hashicorp/aws", Reason: "anchor"},
	}
	got := addrSet(Select(pins, fixtureRows()))
	if got["Telmate/proxmox"].Reason != "whatever-the-file-says" {
		t.Errorf("proxmox reason = %q, want the file's value", got["Telmate/proxmox"].Reason)
	}
}

// TestSelectPinsAbsentUpstream: a pin with no popularity row is still present,
// with an empty tier and zero downloads.
func TestSelectPinsAbsentUpstream(t *testing.T) {
	got := addrSet(Select(fixturePins(), fixtureRows()))
	e, ok := got["nivis-project/hcloudimage"]
	if !ok {
		t.Fatal("curated pin must be present even when absent from fetched rows")
	}
	if e.Tier != "" || e.Downloads != 0 {
		t.Errorf("absent pin = tier %q downloads %d, want empty/0", e.Tier, e.Downloads)
	}
	if _, ok := got["hashicorp/external"]; !ok {
		t.Error("utility hashicorp/external must be pinned even when absent from fetched rows")
	}
}

// TestSelectIncludesAnchorsAndUtilities matches the seed-selection spec.
func TestSelectIncludesAnchorsAndUtilities(t *testing.T) {
	got := addrSet(Select(repoPins(t), fixtureRows()))
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

// TestEuropeanAllowlistPinned: every provider pinned as `europe` reaches the
// manifest with that reason intact.
func TestEuropeanAllowlistPinned(t *testing.T) {
	pins := repoPins(t)
	got := addrSet(Select(pins, fixtureRows()))
	europe := 0
	for _, p := range pins {
		if p.Reason != "europe" {
			continue
		}
		europe++
		e, ok := got[p.Address]
		if !ok {
			t.Errorf("seed missing European provider %q", p.Address)
			continue
		}
		if e.Reason != "europe" {
			t.Errorf("%q reason = %q, want europe", p.Address, e.Reason)
		}
	}
	if europe == 0 {
		t.Error("seed-pins.json declares no European providers")
	}
}

// TestSelectDeterministicOrder asserts pins lead in file order and that the
// sequence is stable, so seed.json is byte-stable.
func TestSelectDeterministicOrder(t *testing.T) {
	pins := repoPins(t)
	entries := Select(pins, fixtureRows())
	for i, p := range pins {
		if entries[i].Address != p.Address {
			t.Errorf("entries[%d] = %q, want pin %q", i, entries[i].Address, p.Address)
		}
	}
	again := Select(pins, fixtureRows())
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
	got := addrSet(Select(repoPins(t), fixtureRows()))
	if got["Telmate/proxmox"].Reason != "anchor" {
		t.Errorf("proxmox reason = %q, want anchor", got["Telmate/proxmox"].Reason)
	}
	if got["hashicorp/tls"].Reason != "utility" {
		t.Errorf("tls reason = %q, want utility", got["hashicorp/tls"].Reason)
	}
	if got["datadog/datadog"].Reason != "popular" {
		t.Errorf("datadog reason = %q, want popular", got["datadog/datadog"].Reason)
	}
}

// TestSelectCapsAtMaxSeed ensures the manifest never exceeds the derived cap.
func TestSelectCapsAtMaxSeed(t *testing.T) {
	pins := repoPins(t)
	entries := Select(pins, fillerRows(MaxSeedFor(pins)*2))
	if len(entries) > MaxSeedFor(pins) {
		t.Errorf("Select returned %d entries, want <= %d", len(entries), MaxSeedFor(pins))
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

// fakeListServer serves the popularity API in pages of two, so Fetch's paging
// and de-duplication are both exercised.
func fakeListServer(t *testing.T, pages [][]Provider) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		idx := offset / 2
		var rows []Provider
		if idx < len(pages) {
			rows = pages[idx]
		}
		next := offset + 2
		if idx >= len(pages)-1 {
			next = offset // no forward progress: Fetch must stop
		}
		resp := map[string]interface{}{
			"meta":      map[string]interface{}{"next_offset": next},
			"providers": rows,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestFetchPagesAndDeduplicates(t *testing.T) {
	srv := fakeListServer(t, [][]Provider{
		{
			{Namespace: "a", Name: "one", Tier: "official", Downloads: 10},
			{Namespace: "b", Name: "two", Tier: "partner", Downloads: 20},
		},
		{
			// "a/one" repeats with a HIGHER count: the higher row must win.
			{Namespace: "a", Name: "one", Tier: "official", Downloads: 99},
			{Namespace: "c", Name: "three", Tier: "community", Downloads: 5},
		},
	})

	rows, err := Fetch(context.Background(), srv.Client(), srv.URL, 100)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("Fetch returned %d distinct providers, want 3", len(rows))
	}
	byAddr := map[string]Provider{}
	for _, p := range rows {
		byAddr[p.Address()] = p
	}
	if got := byAddr["a/one"].Downloads; got != 99 {
		t.Errorf("a/one downloads = %d, want the higher row (99)", got)
	}
}

// TestFetchStopsWhenSatisfied: Fetch must not keep paging once it has `want`.
func TestFetchStopsWhenSatisfied(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"meta": map[string]interface{}{"next_offset": offset + 2},
			"providers": []Provider{
				{Namespace: "n", Name: "a" + strconv.Itoa(offset), Downloads: 1},
				{Namespace: "n", Name: "b" + strconv.Itoa(offset), Downloads: 1},
			},
		})
	}))
	t.Cleanup(srv.Close)

	rows, err := Fetch(context.Background(), srv.Client(), srv.URL, 2)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(rows) < 2 {
		t.Errorf("Fetch returned %d rows, want at least the 2 requested", len(rows))
	}
	if hits != 1 {
		t.Errorf("Fetch made %d requests, want 1 once `want` was satisfied", hits)
	}
}

func TestFetchSurfacesServerErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	t.Cleanup(srv.Close)

	if _, err := Fetch(context.Background(), srv.Client(), srv.URL, 10); err == nil {
		t.Error("Fetch should fail on a malformed response")
	}
}

// TestWriteIsByteStable is the spec's reproducibility scenario.
func TestWriteIsByteStable(t *testing.T) {
	entries := Select(fixturePins(), fixtureRows())
	dir := t.TempDir()
	a := filepath.Join(dir, "a.json")
	b := filepath.Join(dir, "b.json")

	if err := Write(a, DefaultListURL, entries); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := Write(b, DefaultListURL, Select(fixturePins(), fixtureRows())); err != nil {
		t.Fatalf("Write: %v", err)
	}

	ba, err := os.ReadFile(a)
	if err != nil {
		t.Fatal(err)
	}
	bb, err := os.ReadFile(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(ba) != string(bb) {
		t.Error("two Writes of the same selection produced different bytes")
	}
	if !strings.HasSuffix(string(ba), "\n") {
		t.Error("manifest must end with a trailing newline")
	}

	var m Manifest
	if err := json.Unmarshal(ba, &m); err != nil {
		t.Fatalf("manifest does not round-trip: %v", err)
	}
	if m.Count != len(entries) || len(m.Providers) != len(entries) {
		t.Errorf("manifest count = %d / %d entries, want %d", m.Count, len(m.Providers), len(entries))
	}
	if m.GeneratedFrom != DefaultListURL {
		t.Errorf("generated_from = %q, want %q", m.GeneratedFrom, DefaultListURL)
	}
}

func TestWriteRejectsUnwritablePath(t *testing.T) {
	err := Write(filepath.Join(t.TempDir(), "no-such-dir", "seed.json"), DefaultListURL, nil)
	if err == nil {
		t.Error("Write to a missing directory should fail")
	}
}

// TestAddressList: the extraction batch consumes this, so order must survive.
func TestAddressList(t *testing.T) {
	entries := Select(fixturePins(), fixtureRows())
	addrs := AddressList(entries)
	if len(addrs) != len(entries) {
		t.Fatalf("AddressList returned %d addresses, want %d", len(addrs), len(entries))
	}
	for i := range entries {
		if addrs[i] != entries[i].Address {
			t.Errorf("addrs[%d] = %q, want %q", i, addrs[i], entries[i].Address)
		}
	}
	if len(AddressList(nil)) != 0 {
		t.Error("AddressList(nil) should be empty")
	}
}
