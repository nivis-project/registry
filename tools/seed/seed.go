// Package seed derives and pins the provider seed manifest (seed.json) for the
// Nivis registry.
//
// Source of truth for popularity: the PUBLIC Terraform registry listing API,
//
//	https://registry.terraform.io/v1/providers?limit=N&offset=M
//
// which returns per-provider `tier` (official/partner/community) and `downloads`.
// NOTE: the `filter=` query parameter is INERT — it does not filter. Page through
// with limit/offset and sort client-side by (tier, downloads).
//
// The manifest is: the pins declared in seed-pins.json UNION the top remaining
// providers by downloads (favoring official/partner), ordered deterministically
// so seed.json is byte-stable across runs.
//
// See openspec/changes/extraction-pipeline/specs/seed-selection/spec.md.
package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// Pin is one hand-maintained entry in seed-pins.json.
type Pin struct {
	Address string `json:"address"`
	Reason  string `json:"reason"`
	Note    string `json:"note,omitempty"`
}

// PinFile is the seed-pins.json document: the providers that are always in the
// seed regardless of download rank, in selection order, each carrying the
// reason it was pinned.
type PinFile struct {
	Pinned []Pin `json:"pinned"`
}

// DefaultPinsPath is where the pin file lives relative to the repo root.
const DefaultPinsPath = "seed-pins.json"

// PopularFill is how many popularity-ranked providers follow the pins. The cap
// is derived from it (see MaxSeedFor) so that pinning a provider grows the
// manifest instead of evicting the lowest-ranked popular one.
const PopularFill = 38

// LoadPins reads and decodes a pin file.
func LoadPins(path string) ([]Pin, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read pins %s: %w", path, err)
	}
	var pf PinFile
	if err := json.Unmarshal(b, &pf); err != nil {
		return nil, fmt.Errorf("decode pins %s: %w", path, err)
	}
	for i, pin := range pf.Pinned {
		if _, _, err := ParseAddress(pin.Address); err != nil {
			return nil, fmt.Errorf("pins %s entry %d: %w", path, i, err)
		}
		if pin.Reason == "" {
			return nil, fmt.Errorf("pins %s entry %d (%s): reason must not be empty", path, i, pin.Address)
		}
	}
	return pf.Pinned, nil
}

// MaxSeedFor returns the manifest cap for a given pin set.
func MaxSeedFor(pins []Pin) int { return len(pins) + PopularFill }

// DefaultListURL is the public popularity API. The filter= param is inert; page
// with limit/offset and sort client-side.
const DefaultListURL = "https://registry.terraform.io/v1/providers"

// Provider is one entry from the popularity API, reduced to the fields the seed
// selection needs.
type Provider struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Tier      string `json:"tier"`
	Downloads int64  `json:"downloads"`
}

// Address is the canonical "<namespace>/<name>" identity used throughout the
// pipeline and as the seed.json key.
func (p Provider) Address() string { return p.Namespace + "/" + p.Name }

// SeedEntry is one pinned provider in seed.json. Reason records WHY it was
// included (anchor/utility/popular) so the manifest is self-documenting.
type SeedEntry struct {
	Address   string `json:"address"`
	Tier      string `json:"tier"`
	Downloads int64  `json:"downloads"`
	Reason    string `json:"reason"`
}

// Manifest is the pinned seed.json document.
type Manifest struct {
	GeneratedFrom string      `json:"generated_from"` // the popularity API URL
	Count         int         `json:"count"`
	Providers     []SeedEntry `json:"providers"`
}

// tierRank orders the tiers for selection: official > partner > community >
// anything else. Lower is more preferred.
func tierRank(tier string) int {
	switch tier {
	case "official":
		return 0
	case "partner":
		return 1
	case "community":
		return 2
	default:
		return 3
	}
}

// listResponse mirrors the popularity API envelope.
type listResponse struct {
	Meta struct {
		NextOffset int    `json:"next_offset"`
		NextURL    string `json:"next_url"`
	} `json:"meta"`
	Providers []Provider `json:"providers"`
}

// Fetch pages the popularity API until at least `want` distinct provider
// addresses have been collected (or pages run out), returning them de-duplicated
// (the listing yields one row per provider version; keep the highest-download
// row per address). It never trusts filter=; it pages and the caller sorts.
func Fetch(ctx context.Context, client *http.Client, listURL string, want int) ([]Provider, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	if listURL == "" {
		listURL = DefaultListURL
	}
	const pageSize = 100
	byAddr := map[string]Provider{}
	offset := 0
	for len(byAddr) < want {
		url := fmt.Sprintf("%s?limit=%d&offset=%d", listURL, pageSize, offset)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("seed fetch %s: %w", url, err)
		}
		var lr listResponse
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(&lr)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("seed decode %s: %w", url, err)
		}
		if len(lr.Providers) == 0 {
			break // ran out of pages
		}
		for _, p := range lr.Providers {
			if cur, ok := byAddr[p.Address()]; !ok || p.Downloads > cur.Downloads {
				byAddr[p.Address()] = p
			}
		}
		if lr.Meta.NextOffset <= offset {
			break // no forward progress; stop
		}
		offset = lr.Meta.NextOffset
	}
	out := make([]Provider, 0, len(byAddr))
	for _, p := range byAddr {
		out = append(out, p)
	}
	return out, nil
}

// Select turns the raw popularity rows into the pinned seed: every pin is
// present (even if absent from the fetched rows), in file order, followed by the
// most-popular remaining providers sorted by (tier, -downloads, address) up to
// MaxSeedFor(pins).
func Select(pins []Pin, rows []Provider) []SeedEntry {
	byAddr := map[string]Provider{}
	for _, p := range rows {
		byAddr[p.Address()] = p
	}

	maxSeed := MaxSeedFor(pins)
	var out []SeedEntry
	seen := map[string]bool{}
	add := func(addr, reason string) {
		if seen[addr] || len(out) >= maxSeed {
			return
		}
		seen[addr] = true
		p := byAddr[addr]
		out = append(out, SeedEntry{
			Address:   addr,
			Tier:      p.Tier,
			Downloads: p.Downloads,
			Reason:    reason,
		})
	}

	for _, pin := range pins {
		add(pin.Address, pin.Reason)
	}

	popular := make([]Provider, 0, len(rows))
	for _, p := range rows {
		if !seen[p.Address()] {
			popular = append(popular, p)
		}
	}
	sort.Slice(popular, func(i, j int) bool {
		pi, pj := popular[i], popular[j]
		if ri, rj := tierRank(pi.Tier), tierRank(pj.Tier); ri != rj {
			return ri < rj
		}
		if pi.Downloads != pj.Downloads {
			return pi.Downloads > pj.Downloads
		}
		return pi.Address() < pj.Address()
	})
	for _, p := range popular {
		add(p.Address(), "popular")
	}

	return out
}

// Write serializes the manifest to path deterministically (sorted-key,
// indented JSON with a trailing newline) so seed.json is byte-stable.
func Write(path, listURL string, entries []SeedEntry) error {
	m := Manifest{
		GeneratedFrom: listURL,
		Count:         len(entries),
		Providers:     entries,
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

// AddressList returns just the provider addresses from a manifest's entries,
// preserving order — the input the extraction batch consumes.
func AddressList(entries []SeedEntry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Address)
	}
	return out
}

// ParseAddress splits "<namespace>/<name>" into its parts.
func ParseAddress(addr string) (namespace, name string, err error) {
	parts := strings.SplitN(addr, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid provider address %q (want namespace/name)", addr)
	}
	return parts[0], parts[1], nil
}
