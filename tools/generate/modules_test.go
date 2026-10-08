package generate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nivis-project/registry/tools/module"
)

// writeModuleOut lays out a module-extraction root the way tools/module does.
func writeModuleOut(t *testing.T, root string, m module.Module) {
	t.Helper()
	dir := filepath.Join(root, m.Owner, m.Name, m.Version)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "module.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func sampleModule() module.Module {
	return module.Module{
		Owner: "acme", Name: "thing", Version: "0.1.0", Tag: "v0.1.0", Rev: "abc123",
		Description: "a nivis module",
		Resources: []module.Coord{
			{Provider: "aws", Type: "aws_iam_role", Name: "r", Docs: "hashicorp/aws/6.67.0"},
		},
		DataSources: []module.Coord{},
		Outputs:     []string{"role_arn"},
		Composition: []string{"apiEndpointRef"},
		Cfg:         []module.CfgKey{{Path: "name", Required: true}},
		Readme:      "# thing\nhuman prose",
		Record: module.Record{
			StructureExtractable: true, CfgSource: module.CfgSourceScanned,
			ProvidersResolved: 1, ProvidersTotal: 1,
		},
	}
}

func TestModuleContractPathMirrorsProviders(t *testing.T) {
	got := ModuleContractPath("acme", "thing", "0.1.0")
	want := filepath.Join("registry", "docs", "modules", "acme", "thing", "0.1.0")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGenerateModuleEmitsTheDerivedStructure(t *testing.T) {
	moduleRoot, contractRoot := t.TempDir(), t.TempDir()
	writeModuleOut(t, moduleRoot, sampleModule())

	path, err := GenerateModule(filepath.Join(moduleRoot, "acme", "thing", "0.1.0"), contractRoot)
	if err != nil {
		t.Fatalf("GenerateModule: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var mv ModuleVersion
	if err := json.Unmarshal(b, &mv); err != nil {
		t.Fatalf("index.json does not decode: %v", err)
	}
	if mv.ID != "0.1.0" || mv.Owner != "acme" || mv.Name != "thing" {
		t.Errorf("identity wrong: %+v", mv)
	}
	if len(mv.Resources) != 1 || mv.Resources[0].Docs != "hashicorp/aws/6.67.0" {
		t.Errorf("provider link did not travel with the contract: %+v", mv.Resources)
	}
	if !mv.Record.StructureExtractable || mv.Record.CfgSource != module.CfgSourceScanned {
		t.Errorf("the extraction record must travel with the contract, got %+v", mv.Record)
	}
	if mv.Readme == "" {
		t.Error("the written documentation must be carried")
	}
	// Human prose must be a distinct field, never merged into the structure.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["readme"]; !ok {
		t.Error("readme must be its own field")
	}
	if _, ok := raw["resources"]; !ok {
		t.Error("resources must be its own field")
	}
}

func TestGenerateModuleRejectsBadInput(t *testing.T) {
	contractRoot := t.TempDir()
	if _, err := GenerateModule(filepath.Join(t.TempDir(), "absent"), contractRoot); err == nil {
		t.Error("a missing module.json should fail")
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "module.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateModule(dir, contractRoot); err == nil {
		t.Error("malformed module.json should fail")
	}

	dir2 := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir2, "module.json"), []byte(`{"owner":"a","name":"b"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateModule(dir2, contractRoot); err == nil {
		t.Error("a module.json with no version should fail")
	}
}

func TestGenerateModulesEmitsACatalogue(t *testing.T) {
	moduleRoot, contractRoot := t.TempDir(), t.TempDir()
	writeModuleOut(t, moduleRoot, sampleModule())

	second := sampleModule()
	second.Name = "another"
	second.Record.StructureExtractable = false
	writeModuleOut(t, moduleRoot, second)

	paths, err := GenerateModules(moduleRoot, contractRoot, nil)
	if err != nil {
		t.Fatalf("GenerateModules: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("emitted %d contracts, want 2", len(paths))
	}

	b, err := os.ReadFile(filepath.Join(contractRoot, "registry", "modules.json"))
	if err != nil {
		t.Fatalf("modules.json not written: %v", err)
	}
	var catalog []ModuleCatalogEntry
	if err := json.Unmarshal(b, &catalog); err != nil {
		t.Fatal(err)
	}
	if len(catalog) != 2 {
		t.Fatalf("catalogue has %d entries, want 2", len(catalog))
	}
	if catalog[0].Name != "another" {
		t.Errorf("catalogue is not sorted, got %q first", catalog[0].Name)
	}
	// Every catalogue row must resolve to an index that exists.
	for _, e := range catalog {
		p := filepath.Join(contractRoot, ModuleContractPath(e.Owner, e.Name, e.Version), "index.json")
		if _, err := os.Stat(p); err != nil {
			t.Errorf("catalogue lists %s/%s@%s but %s is missing", e.Owner, e.Name, e.Version, p)
		}
	}
	// The catalogue says which modules were derived, so the SPA need not guess.
	for _, e := range catalog {
		if e.Name == "another" && e.Derived {
			t.Error("a module whose structure did not evaluate must not be marked derived")
		}
		if e.Name == "thing" && !e.Derived {
			t.Error("a derived module must be marked derived")
		}
	}
}

// TestGenerateModulesReportsAnAbsentRoot: a path that does not exist is a
// misconfiguration, and it must be visible. A silently empty catalogue is
// exactly how the pipeline scripts shipped an empty modules.json while
// reporting success.
func TestGenerateModulesReportsAnAbsentRoot(t *testing.T) {
	contractRoot := t.TempDir()
	var logs []string
	paths, err := GenerateModules(filepath.Join(t.TempDir(), "absent"), contractRoot,
		func(f string, a ...interface{}) { logs = append(logs, fmt.Sprintf(f, a...)) })
	if err != nil {
		t.Fatalf("a missing module root should not error: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("got %d paths, want none", len(paths))
	}

	var warned bool
	for _, l := range logs {
		if strings.Contains(l, "does not exist") {
			warned = true
		}
	}
	if !warned {
		t.Errorf("a missing module root must be reported; logs were %v", logs)
	}

	b, err := os.ReadFile(filepath.Join(contractRoot, "registry", "modules.json"))
	if err != nil {
		t.Fatalf("an empty catalogue must still be written: %v", err)
	}
	if string(b) != "[]\n" {
		t.Errorf("empty catalogue = %q, want []", b)
	}
}

// TestGenerateModulesIsQuietOnAnEmptyRoot: genuinely having no modules yet is
// not a misconfiguration and must not be reported as one.
func TestGenerateModulesIsQuietOnAnEmptyRoot(t *testing.T) {
	moduleRoot, contractRoot := t.TempDir(), t.TempDir()
	var logs []string
	if _, err := GenerateModules(moduleRoot, contractRoot,
		func(f string, a ...interface{}) { logs = append(logs, fmt.Sprintf(f, a...)) }); err != nil {
		t.Fatalf("an existing but empty module root should not error: %v", err)
	}
	for _, l := range logs {
		if strings.Contains(l, "does not exist") {
			t.Errorf("an existing root must not be reported as missing: %q", l)
		}
	}
}

func TestGenerateModulesSkipsUnreadableEntries(t *testing.T) {
	moduleRoot, contractRoot := t.TempDir(), t.TempDir()
	writeModuleOut(t, moduleRoot, sampleModule())
	bad := filepath.Join(moduleRoot, "acme", "broken", "0.1.0")
	if err := os.MkdirAll(bad, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bad, "module.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}

	var logs int
	paths, err := GenerateModules(moduleRoot, contractRoot, func(string, ...interface{}) { logs++ })
	if err != nil {
		t.Fatalf("one bad module must not abort generation: %v", err)
	}
	if len(paths) != 1 {
		t.Errorf("emitted %d contracts, want 1 (the healthy one)", len(paths))
	}
	if logs == 0 {
		t.Error("the skip should be logged")
	}
}

// --- enrichment -------------------------------------------------------------

func TestModuleCatalogueCarriesTheOwnerAvatar(t *testing.T) {
	moduleRoot, contractRoot := t.TempDir(), t.TempDir()
	writeModuleOut(t, moduleRoot, sampleModule())

	enr := Enrichment{Avatars: map[string]string{"acme": "avatars/acme.png"}}
	if _, err := GenerateModulesEnriched(moduleRoot, contractRoot, enr, nil); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(contractRoot, "registry", "modules.json"))
	if err != nil {
		t.Fatal(err)
	}
	var catalog []ModuleCatalogEntry
	if err := json.Unmarshal(b, &catalog); err != nil {
		t.Fatal(err)
	}
	if catalog[0].Avatar != "avatars/acme.png" {
		t.Errorf("avatar = %q, want the owner's reference", catalog[0].Avatar)
	}
}

func TestModuleCatalogueOmitsAnAbsentAvatar(t *testing.T) {
	moduleRoot, contractRoot := t.TempDir(), t.TempDir()
	writeModuleOut(t, moduleRoot, sampleModule())

	if _, err := GenerateModules(moduleRoot, contractRoot, nil); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(contractRoot, "registry", "modules.json"))
	if err != nil {
		t.Fatal(err)
	}
	// Absent, not empty: a consumer must not be able to render a broken image.
	if strings.Contains(string(b), "avatar") {
		t.Errorf("modules.json should carry no avatar key when none was obtained:\n%s", b)
	}
}
