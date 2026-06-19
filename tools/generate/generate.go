// Package generate emits the static JSON contract consumed by the forked
// registry-ui SPA, populated with schema-derived Nix constructors.
//
// Contract shape mirrors opentofu/registry-ui (backend/internal/server/openapi.yml):
//
//	registry/docs/providers/{ns}/{name}/{version}/index.json   # ProviderVersion: lists docs.{resources,datasources,functions,guides} + compat badge
//	registry/docs/providers/{ns}/{name}/{version}/<item>.md     # per-item document (Nix-rendered body)
//
// The Nivis difference: the per-item document body is the RENDERED `nivis gen`
// constructor (signature + types + computed outputs), NOT scraped HCL markdown.
// Each provider index.json also carries the compat record from tools/compat.
//
// See openspec/changes/contract-generator/specs/{docs-contract,nix-rendering}/spec.md.
package generate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nivis-project/registry/tools/compat"
	"github.com/nivis-project/registry/tools/extract"
)

// compatMetadata is the normalized provider-version metadata written by
// tools/extract; reused here so the contract and compat record stay in sync
// with the single source of truth for the shape.
type compatMetadata = extract.Metadata

// ContractPath returns the index.json path for a provider version, relative to
// the contract root.
func ContractPath(namespace, name, version string) string {
	return filepath.Join("registry", "docs", "providers", namespace, name, version, "index.json")
}

// DocItem mirrors registry-ui's ProviderDocItem.
type DocItem struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Subcategory string `json:"subcategory,omitempty"`
}

// Docs mirrors registry-ui's ProviderDocs (the lists the SPA routes over).
type Docs struct {
	Resources   []DocItem `json:"resources"`
	Datasources []DocItem `json:"datasources"`
	Functions   []DocItem `json:"functions"`
	Guides      []DocItem `json:"guides"`
}

// ProviderVersion mirrors registry-ui's ProviderVersion, extended with the Nivis
// `compat` record. The frontend keeps the registry-ui fields; `compat` is the
// only addition (the resource-page rendering is adapted, not the contract types).
type ProviderVersion struct {
	ID        string         `json:"id"` // version number
	Published string         `json:"published,omitempty"`
	Docs      Docs           `json:"docs"`
	Compat    compat.Record  `json:"compat"`
	Nivis     map[string]any `json:"nivis,omitempty"` // reserved for future Nivis metadata
}

// itemDocFileName is the per-item document filename for an item (Nix-rendered).
func itemDocFileName(itemName string) string { return itemName + ".md" }

// GenerateProvider walks one extracted provider-version directory
// (extractDir = <extract-out>/<ns>/<name>/<version>, containing <identity>/*.nix
// and metadata.json), renders each constructor, writes per-item documents under
// contractRoot, and writes the provider index.json with the compat record.
//
// It returns the emitted index.json path. A provider with no .nix files yields
// an index with empty lists (still valid) rather than an error.
func GenerateProvider(extractDir, contractRoot string) (string, error) {
	meta, err := readMetadata(extractDir)
	if err != nil {
		return "", err
	}
	identity := meta.Name
	idDir := filepath.Join(extractDir, identity)

	nixFiles, _ := filepath.Glob(filepath.Join(idDir, "*.nix"))
	sort.Strings(nixFiles)

	outDir := filepath.Join(contractRoot, "registry", "docs", "providers", meta.Namespace, meta.Name, meta.Version)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}

	var resources []DocItem
	for _, nf := range nixFiles {
		body, rerr := os.ReadFile(nf)
		if rerr != nil {
			return "", rerr
		}
		c := ParseConstructor(string(body))
		if c.Type == "" {
			// Fall back to the filename if the mkResource line was not found.
			c.Type = strings.TrimSuffix(filepath.Base(nf), ".nix")
		}
		// Render and write the per-item document.
		doc := RenderDoc(c)
		docPath := filepath.Join(outDir, itemDocFileName(c.Type))
		if werr := os.WriteFile(docPath, []byte(doc), 0o644); werr != nil {
			return "", werr
		}
		resources = append(resources, DocItem{
			Name:        c.Type,
			Title:       c.Type,
			Description: fmt.Sprintf("Nivis Nix constructor for %s (%d required, %d optional, %d computed)", c.Type, len(c.Required), len(c.Optional), len(c.Computed)),
			Subcategory: meta.Name,
		})
	}
	sort.Slice(resources, func(i, j int) bool { return resources[i].Name < resources[j].Name })

	// nivis gen v0.4.0 emits resources only; datasources/functions/guides are
	// empty lists for now (kept present so the contract validates).
	pv := ProviderVersion{
		ID: meta.Version,
		Docs: Docs{
			Resources:   resources,
			Datasources: []DocItem{},
			Functions:   []DocItem{},
			Guides:      []DocItem{},
		},
		Compat: compat.Compute(meta, len(resources) > 0),
	}

	indexPath := filepath.Join(outDir, "index.json")
	b, err := json.MarshalIndent(pv, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(indexPath, append(b, '\n'), 0o644); err != nil {
		return "", err
	}
	return indexPath, nil
}

// CatalogEntry is one provider version in the browsable catalog (catalog.json),
// which lets the SPA list providers without a live search backend (v1).
type CatalogEntry struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	Tier      string `json:"tier"`
}

// GenerateAll walks every provider-version directory under extractRoot
// (extractRoot/<ns>/<name>/<version>/), generates each, and writes the browsable
// catalog.json. It is resilient: a provider that fails generation is logged via
// log (if non-nil) and skipped.
func GenerateAll(extractRoot, contractRoot string, log func(string, ...interface{})) ([]string, error) {
	if log == nil {
		log = func(string, ...interface{}) {}
	}
	versionDirs, err := findVersionDirs(extractRoot)
	if err != nil {
		return nil, err
	}
	var out []string
	var catalog []CatalogEntry
	for _, vd := range versionDirs {
		idx, gerr := GenerateProvider(vd, contractRoot)
		if gerr != nil {
			log("skip %s: %v", vd, gerr)
			continue
		}
		log("ok   %s -> %s", vd, idx)
		out = append(out, idx)
		if meta, merr := readMetadata(vd); merr == nil {
			rec := compat.Compute(meta, true)
			catalog = append(catalog, CatalogEntry{
				Namespace: meta.Namespace, Name: meta.Name, Version: meta.Version, Tier: string(rec.Tier),
			})
		}
	}
	if err := writeCatalog(contractRoot, catalog); err != nil {
		return out, err
	}
	return out, nil
}

// writeCatalog writes the browsable provider list to registry/catalog.json.
func writeCatalog(contractRoot string, catalog []CatalogEntry) error {
	sort.Slice(catalog, func(i, j int) bool {
		if catalog[i].Namespace != catalog[j].Namespace {
			return catalog[i].Namespace < catalog[j].Namespace
		}
		return catalog[i].Name < catalog[j].Name
	})
	if catalog == nil {
		catalog = []CatalogEntry{}
	}
	dir := filepath.Join(contractRoot, "registry")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "catalog.json"), append(b, '\n'), 0o644)
}

// findVersionDirs returns every directory containing a metadata.json under root.
func findVersionDirs(root string) ([]string, error) {
	var dirs []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // tolerate unreadable subtrees
		}
		if !info.IsDir() && info.Name() == "metadata.json" {
			dirs = append(dirs, filepath.Dir(path))
		}
		return nil
	})
	sort.Strings(dirs)
	return dirs, err
}

// readMetadata loads the normalized metadata.json written by tools/extract.
func readMetadata(dir string) (compatMetadata, error) {
	b, err := os.ReadFile(filepath.Join(dir, "metadata.json"))
	if err != nil {
		return compatMetadata{}, fmt.Errorf("read metadata: %w", err)
	}
	var m compatMetadata
	if err := json.Unmarshal(b, &m); err != nil {
		return compatMetadata{}, fmt.Errorf("parse metadata: %w", err)
	}
	return m, nil
}
