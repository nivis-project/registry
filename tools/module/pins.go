// Package module catalogues nivis modules the way tools/extract catalogues
// providers. A module is a flake output, not a binary, so there is no schema
// call; its structure is derived by EVALUATING it with the configuration
// poisoned, which yields every resource, data source and output it declares.
//
// Three passes per module:
//
//	A. structure    nix-instantiate on probe.nix  (authoritative)
//	B. cfg surface  scan of the module source     (inferred, labelled as such)
//	C. provider links  lookup against the generated provider contract
//
// See openspec/changes/module-pipeline/specs/module-extraction/spec.md.
package module

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// DefaultPinsPath is the hand-maintained module list, at the repo root beside
// seed-pins.json.
const DefaultPinsPath = "module-pins.json"

// Pin is one hand-maintained entry in module-pins.json.
type Pin struct {
	Owner  string `json:"owner"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
	Note   string `json:"note,omitempty"`
}

// Repo is the canonical "<owner>/<name>" identity.
func (p Pin) Repo() string { return p.Owner + "/" + p.Name }

// GitURL is the clone URL the tag listing and the flake reference both use.
// Plain git over https, NOT the GitHub API: the API path fails under
// organisation policies that restrict fine-grained tokens.
func (p Pin) GitURL() string { return "https://github.com/" + p.Owner + "/" + p.Name }

// PinFile is the module-pins.json document, in catalogue order.
type PinFile struct {
	Modules []Pin `json:"modules"`
}

// LoadPins reads and validates a pin file. A hand-edited file must fail loudly
// rather than yield a silent default.
func LoadPins(path string) ([]Pin, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read module pins %s: %w", path, err)
	}
	var pf PinFile
	if err := json.Unmarshal(b, &pf); err != nil {
		return nil, fmt.Errorf("decode module pins %s: %w", path, err)
	}
	seen := map[string]bool{}
	for i, p := range pf.Modules {
		switch {
		case strings.TrimSpace(p.Owner) == "":
			return nil, fmt.Errorf("module pins %s entry %d: owner must not be empty", path, i)
		case strings.TrimSpace(p.Name) == "":
			return nil, fmt.Errorf("module pins %s entry %d: name must not be empty", path, i)
		case strings.TrimSpace(p.Reason) == "":
			return nil, fmt.Errorf("module pins %s entry %d (%s): reason must not be empty", path, i, p.Repo())
		case strings.ContainsAny(p.Owner+p.Name, "/ "):
			return nil, fmt.Errorf("module pins %s entry %d: owner and name must not contain %q or a space", path, i, "/")
		case seen[p.Repo()]:
			return nil, fmt.Errorf("module pins %s entry %d: duplicate module %s", path, i, p.Repo())
		}
		seen[p.Repo()] = true
	}
	return pf.Modules, nil
}
