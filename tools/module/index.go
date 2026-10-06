package module

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ProviderIndex maps a resource type to the provider version documenting it.
//
// It is built from the generated provider output, where a constructor's
// FILENAME is the resource type. That makes the lookup exact. Inferring a
// provider from a type-name prefix ("aws_" means hashicorp/aws) holds only
// until a fork or a vendored provider makes it false.
type ProviderIndex map[string]string

// BuildProviderIndex walks an extraction-output root
// (<namespace>/<name>/<version>/<identity>/<type>.nix) and indexes every
// constructor by resource type. When two providers emit the same type the
// lexically first address wins, so the index is deterministic.
func BuildProviderIndex(extractRoot string) (ProviderIndex, error) {
	type hit struct{ addr, typ string }
	var hits []hit

	err := filepath.WalkDir(extractRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".nix") {
			return nil
		}
		rel, relErr := filepath.Rel(extractRoot, path)
		if relErr != nil {
			return nil
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) < 5 {
			return nil
		}
		namespace, name, ver := parts[0], parts[1], parts[2]
		hits = append(hits, hit{
			addr: namespace + "/" + name + "/" + ver,
			typ:  strings.TrimSuffix(d.Name(), ".nix"),
		})
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return ProviderIndex{}, nil
		}
		return nil, err
	}

	sort.Slice(hits, func(i, j int) bool {
		if hits[i].typ != hits[j].typ {
			return hits[i].typ < hits[j].typ
		}
		return hits[i].addr < hits[j].addr
	})
	idx := make(ProviderIndex, len(hits))
	for _, h := range hits {
		if _, dup := idx[h.typ]; !dup {
			idx[h.typ] = h.addr
		}
	}
	return idx, nil
}

// Resolve returns the provider version documenting a resource type, and false
// when the registry does not catalogue it.
func (p ProviderIndex) Resolve(resourceType string) (string, bool) {
	addr, ok := p[resourceType]
	return addr, ok
}
