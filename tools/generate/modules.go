package generate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/nivis-project/registry/tools/module"
)

// ModuleContractPath is where one module version's index lands, mirroring the
// provider tree so the SPA reads both through the same mechanism.
func ModuleContractPath(owner, name, version string) string {
	return filepath.Join("registry", "docs", "modules", owner, name, version)
}

// ModuleVersion is a module index.json: the derived structure, the inferred
// configuration surface, the extraction record, and the module's own written
// documentation kept in a SEPARATE field so a consumer can present the derived
// and the human-written halves differently.
type ModuleVersion struct {
	ID          string          `json:"id"` // version
	Owner       string          `json:"owner"`
	Name        string          `json:"name"`
	Tag         string          `json:"tag,omitempty"`
	Rev         string          `json:"rev,omitempty"`
	Description string          `json:"description,omitempty"`
	Resources   []module.Coord  `json:"resources"`
	DataSources []module.Coord  `json:"data_sources"`
	Outputs     []string        `json:"outputs"`
	Composition []string        `json:"composition,omitempty"`
	Cfg         []module.CfgKey `json:"cfg"`
	Record      module.Record   `json:"record"`
	Readme      string          `json:"readme,omitempty"`
}

// ModuleCatalogEntry is one row of registry/modules.json.
type ModuleCatalogEntry struct {
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
	Derived     bool   `json:"derived"`
}

// GenerateModule emits one module's contract from its extraction output
// directory, returning the index path.
func GenerateModule(moduleDir, contractRoot string) (string, error) {
	b, err := os.ReadFile(filepath.Join(moduleDir, "module.json"))
	if err != nil {
		return "", fmt.Errorf("read module.json in %s: %w", moduleDir, err)
	}
	var m module.Module
	if err := json.Unmarshal(b, &m); err != nil {
		return "", fmt.Errorf("decode module.json in %s: %w", moduleDir, err)
	}
	if m.Version == "" {
		return "", fmt.Errorf("module.json in %s has no version", moduleDir)
	}

	mv := ModuleVersion{
		ID:          m.Version,
		Owner:       m.Owner,
		Name:        m.Name,
		Tag:         m.Tag,
		Rev:         m.Rev,
		Description: m.Description,
		Resources:   coordsOrEmpty(m.Resources),
		DataSources: coordsOrEmpty(m.DataSources),
		Outputs:     stringsOrEmpty(m.Outputs),
		Composition: m.Composition,
		Cfg:         cfgOrEmpty(m.Cfg),
		Record:      m.Record,
		Readme:      m.Readme,
	}

	dir := filepath.Join(contractRoot, ModuleContractPath(m.Owner, m.Name, m.Version))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	out, err := json.MarshalIndent(mv, "", "  ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "index.json")
	return path, os.WriteFile(path, append(out, '\n'), 0o644)
}

// GenerateModules walks a module-extraction root and emits every module's
// contract plus registry/modules.json. A skipped module has no structure to
// publish and is left out of the contract.
func GenerateModules(moduleRoot, contractRoot string, log func(string, ...interface{})) ([]string, error) {
	if log == nil {
		log = func(string, ...interface{}) {}
	}
	if _, statErr := os.Stat(moduleRoot); os.IsNotExist(statErr) {
		// Distinct from "no modules extracted": a path that does not exist is a
		// misconfiguration, and reporting it is what stops an empty catalogue
		// from being mistaken for a complete one.
		log("warn module root %s does not exist; writing an empty module catalogue", moduleRoot)
	}
	dirs, err := findModuleDirs(moduleRoot)
	if err != nil {
		return nil, err
	}

	var paths []string
	var catalog []ModuleCatalogEntry
	for _, dir := range dirs {
		path, err := GenerateModule(dir, contractRoot)
		if err != nil {
			log("skip %s: %v", dir, err)
			continue
		}
		b, readErr := os.ReadFile(filepath.Join(dir, "module.json"))
		if readErr != nil {
			continue
		}
		var m module.Module
		if json.Unmarshal(b, &m) != nil {
			continue
		}
		catalog = append(catalog, ModuleCatalogEntry{
			Owner:       m.Owner,
			Name:        m.Name,
			Version:     m.Version,
			Description: m.Description,
			Derived:     m.Record.StructureExtractable,
		})
		paths = append(paths, path)
		log("ok   %s/%s @ %s", m.Owner, m.Name, m.Version)
	}

	sort.Slice(catalog, func(i, j int) bool {
		a, b := catalog[i], catalog[j]
		if a.Owner != b.Owner {
			return a.Owner < b.Owner
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.Version < b.Version
	})
	if err := writeModuleCatalog(contractRoot, catalog); err != nil {
		return nil, err
	}
	return paths, nil
}

func writeModuleCatalog(contractRoot string, catalog []ModuleCatalogEntry) error {
	if catalog == nil {
		catalog = []ModuleCatalogEntry{}
	}
	dir := filepath.Join(contractRoot, "registry")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "modules.json"), append(b, '\n'), 0o644)
}

// findModuleDirs returns every <owner>/<name>/<version> directory holding a
// module.json, in deterministic order.
func findModuleDirs(root string) ([]string, error) {
	var out []string
	owners, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	for _, owner := range owners {
		if !owner.IsDir() {
			continue
		}
		names, err := os.ReadDir(filepath.Join(root, owner.Name()))
		if err != nil {
			continue
		}
		for _, name := range names {
			if !name.IsDir() {
				continue
			}
			versions, err := os.ReadDir(filepath.Join(root, owner.Name(), name.Name()))
			if err != nil {
				continue
			}
			for _, v := range versions {
				if !v.IsDir() {
					continue
				}
				dir := filepath.Join(root, owner.Name(), name.Name(), v.Name())
				if _, err := os.Stat(filepath.Join(dir, "module.json")); err == nil {
					out = append(out, dir)
				}
			}
		}
	}
	sort.Strings(out)
	return out, nil
}

func coordsOrEmpty(c []module.Coord) []module.Coord {
	if c == nil {
		return []module.Coord{}
	}
	return c
}

func stringsOrEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func cfgOrEmpty(c []module.CfgKey) []module.CfgKey {
	if c == nil {
		return []module.CfgKey{}
	}
	return c
}
