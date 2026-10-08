// Command generate walks the extraction output and emits the static contract:
// per-provider and per-module index documents, the two catalogues, and the
// owner avatars the catalogues reference.
//
//	go run ./cmd/generate -extract ../extract-out -modules ../module-out \
//	    -seed ../seed.json -contract ..
//
// The generate package itself does no network access. This command gathers the
// facts from outside the extraction tree (the seed, and the avatars) and hands
// them over, so generation stays testable without a network.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/nivis-project/registry/tools/avatar"
	"github.com/nivis-project/registry/tools/generate"
	"github.com/nivis-project/registry/tools/module"
	"github.com/nivis-project/registry/tools/seed"
)

func main() {
	extractRoot := flag.String("extract", "extract-out", "extraction-output root (from tools/extract)")
	moduleRoot := flag.String("modules", "module-out", "module-extraction root (from tools/module)")
	seedPath := flag.String("seed", "", "seed.json, for why each provider is catalogued and its upstream standing")
	modulePins := flag.String("module-pins", "", "module-pins.json, so module owners get an avatar too")
	contractRoot := flag.String("contract", ".", "contract root (writes <root>/registry/docs/...)")
	noAvatars := flag.Bool("no-avatars", false, "skip fetching owner avatars (offline runs)")
	flag.Parse()

	enr := generate.Enrichment{
		Providers: map[string]generate.ProviderFacts{},
		Avatars:   map[string]string{},
	}

	if *seedPath != "" {
		facts, owners, err := readSeed(*seedPath)
		if err != nil {
			log.Fatalf("generate: %v", err)
		}
		enr.Providers = facts
		log.Printf("generate: seed supplied facts for %d provider(s)", len(facts))

		// Module owners are not provider namespaces, so without this the
		// owner of every module has no avatar.
		if *modulePins != "" {
			pins, perr := module.LoadPins(*modulePins)
			if perr != nil {
				log.Fatalf("generate: %v", perr)
			}
			for _, p := range pins {
				owners = append(owners, p.Owner)
			}
		}

		if !*noAvatars {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			enr.Avatars = avatar.New().FetchAll(ctx, owners, *contractRoot, log.Printf)
			cancel()
			log.Printf("generate: %d owner avatar(s) of %d", len(enr.Avatars), countOwners(owners))
		}
	}

	idxs, err := generate.GenerateAllEnriched(*extractRoot, *contractRoot, enr, log.Printf)
	if err != nil {
		log.Fatalf("generate: %v", err)
	}
	log.Printf("generate: emitted %d provider contract(s)", len(idxs))

	mods, err := generate.GenerateModulesEnriched(*moduleRoot, *contractRoot, enr, log.Printf)
	if err != nil {
		log.Fatalf("generate: modules: %v", err)
	}
	log.Printf("generate: emitted %d module contract(s)", len(mods))

	if len(idxs) == 0 {
		os.Exit(1)
	}
}

// readSeed returns the per-provider facts keyed by address, and every owner
// whose avatar is worth fetching.
func readSeed(path string) (map[string]generate.ProviderFacts, []string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var m seed.Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, nil, err
	}
	facts := make(map[string]generate.ProviderFacts, len(m.Providers))
	owners := make([]string, 0, len(m.Providers))
	for _, p := range m.Providers {
		facts[p.Address] = generate.ProviderFacts{Reason: p.Reason, Publisher: p.Tier}
		ns, _, err := seed.ParseAddress(p.Address)
		if err == nil {
			owners = append(owners, ns)
		}
	}
	return facts, owners, nil
}

func countOwners(owners []string) int {
	seen := map[string]bool{}
	for _, o := range owners {
		seen[o] = true
	}
	return len(seen)
}
