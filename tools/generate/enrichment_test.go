package generate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nivis-project/registry/tools/extract"
)

// writeExtractOut lays out an extraction tree the way tools/extract does.
func writeExtractOut(t *testing.T, root string, meta extract.Metadata) {
	t.Helper()
	dir := filepath.Join(root, meta.Namespace, meta.Name, meta.Version)
	if err := os.MkdirAll(filepath.Join(dir, meta.Name), 0o755); err != nil {
		t.Fatal(err)
	}
	b, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	nix := "# `thing`\n\n```nix\n{ name }\n```\n"
	if err := os.WriteFile(filepath.Join(dir, meta.Name, "thing.nix"), []byte(nix), 0o644); err != nil {
		t.Fatal(err)
	}
}

func sampleMeta() extract.Metadata {
	return extract.Metadata{
		Address: "ovh/ovh", Namespace: "ovh", Name: "ovh", Version: "1.0.0",
		Protocols: []string{"5.0"},
		Platforms: []extract.Platform{{OS: "linux", Arch: "amd64"}},
	}
}

func readCatalog(t *testing.T, contractRoot string) []CatalogEntry {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(contractRoot, "registry", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	var c []CatalogEntry
	if err := json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCatalogueCarriesReasonAvatarAndPublisher(t *testing.T) {
	extractRoot, contractRoot := t.TempDir(), t.TempDir()
	writeExtractOut(t, extractRoot, sampleMeta())

	enr := Enrichment{
		Providers: map[string]ProviderFacts{"ovh/ovh": {Reason: "europe", Publisher: "partner"}},
		Avatars:   map[string]string{"ovh": "avatars/ovh.png"},
	}
	if _, err := GenerateAllEnriched(extractRoot, contractRoot, enr, nil); err != nil {
		t.Fatal(err)
	}
	e := readCatalog(t, contractRoot)[0]
	if e.Reason != "europe" {
		t.Errorf("reason = %q, want the seed's value verbatim", e.Reason)
	}
	if e.Publisher != "partner" {
		t.Errorf("publisher = %q", e.Publisher)
	}
	if e.Avatar != "avatars/ovh.png" {
		t.Errorf("avatar = %q", e.Avatar)
	}
}

// TestCompatTierKeepsItsMeaning is the collision this change had to avoid:
// `tier` was already the compat tier and the SPA renders it. The upstream
// standing is a different axis and takes a different name.
func TestCompatTierKeepsItsMeaning(t *testing.T) {
	extractRoot, contractRoot := t.TempDir(), t.TempDir()
	writeExtractOut(t, extractRoot, sampleMeta())

	enr := Enrichment{
		Providers: map[string]ProviderFacts{"ovh/ovh": {Reason: "europe", Publisher: "partner"}},
	}
	if _, err := GenerateAllEnriched(extractRoot, contractRoot, enr, nil); err != nil {
		t.Fatal(err)
	}
	e := readCatalog(t, contractRoot)[0]
	if e.Tier != "compatible by design" {
		t.Errorf("tier = %q, want the compat tier unchanged", e.Tier)
	}
	if e.Tier == e.Publisher {
		t.Error("the two axes must not collapse into one value")
	}
}

func TestCatalogueOmitsWhatIsNotKnown(t *testing.T) {
	extractRoot, contractRoot := t.TempDir(), t.TempDir()
	writeExtractOut(t, extractRoot, sampleMeta())

	// Nothing supplied: no seed, no avatars.
	if _, err := GenerateAll(extractRoot, contractRoot, nil); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(contractRoot, "registry", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"avatar", "reason", "publisher"} {
		if strings.Contains(string(b), key) {
			t.Errorf("catalog.json should omit %q when unknown:\n%s", key, b)
		}
	}
	// What was always there is still there.
	e := readCatalog(t, contractRoot)[0]
	if e.Namespace != "ovh" || e.Tier == "" {
		t.Errorf("the pre-existing fields must survive: %+v", e)
	}
}

func TestPublisherOmittedWhenUpstreamReportsNone(t *testing.T) {
	extractRoot, contractRoot := t.TempDir(), t.TempDir()
	writeExtractOut(t, extractRoot, sampleMeta())

	// Pinned providers such as Telmate/proxmox have no upstream standing.
	enr := Enrichment{Providers: map[string]ProviderFacts{"ovh/ovh": {Reason: "anchor", Publisher: ""}}}
	if _, err := GenerateAllEnriched(extractRoot, contractRoot, enr, nil); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(contractRoot, "registry", "catalog.json"))
	if strings.Contains(string(b), "publisher") {
		t.Errorf("publisher must be absent, not empty:\n%s", b)
	}
	if readCatalog(t, contractRoot)[0].Reason != "anchor" {
		t.Error("reason should still be carried")
	}
}
