// Command module catalogues the pinned nivis modules: resolve each one's
// latest release tag, derive its structure by evaluating it with the
// configuration poisoned, scan its configuration surface, and link every
// resource type to the provider that documents it.
//
//	go run ./cmd/module -pins ../module-pins.json -extract ../extract-out -out ../module-out
//
// Network is used to list tags and to fetch the pinned flake revisions.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/nivis-project/registry/tools/module"
)

func main() {
	pinsPath := flag.String("pins", module.DefaultPinsPath, "path to the hand-maintained module pin file")
	extractRoot := flag.String("extract", "extract-out", "provider extraction root, for the resource-type index")
	outRoot := flag.String("out", "module-out", "output root for per-module extraction results")
	nivisRef := flag.String("nivis-ref", module.DefaultNivisRef, "nivis flake reference modules are applied with")
	timeout := flag.Duration("timeout", module.DefaultTimeout, "per-module evaluation timeout")
	flag.Parse()

	pins, err := module.LoadPins(*pinsPath)
	if err != nil {
		log.Fatalf("module: %v", err)
	}

	index, err := module.BuildProviderIndex(*extractRoot)
	if err != nil {
		log.Fatalf("module: build provider index: %v", err)
	}
	log.Printf("module: resource-type index has %d entries from %s", len(index), *extractRoot)

	c := module.New(
		module.WithProviderIndex(index),
		module.WithNivisRef(*nivisRef),
		module.WithTimeout(*timeout),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	results := c.Batch(ctx, pins, log.Printf)

	var ok, skipped int
	for _, m := range results {
		if m.Skipped() {
			skipped++
			continue
		}
		ok++
		if _, err := module.Write(*outRoot, m); err != nil {
			log.Printf("warn %s: write failed: %v", m.Repo(), err)
		}
	}
	log.Printf("module: %d ok, %d skipped, %d total", ok, skipped, len(results))
}
