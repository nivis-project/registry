// Command seed derives the pinned provider manifest (seed.json) from the public
// Terraform registry popularity API and writes it deterministically.
//
//	go run ./cmd/seed -out ../seed.json
//
// Network is used only to read public provider METADATA (no binaries execute).
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/nivis-project/registry/tools/seed"
)

func main() {
	out := flag.String("out", "seed.json", "path to write the pinned seed manifest")
	listURL := flag.String("list-url", seed.DefaultListURL, "popularity API base URL")
	want := flag.Int("want", seed.MaxSeed*3, "how many distinct providers to fetch before selecting")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	rows, err := seed.Fetch(ctx, nil, *listURL, *want)
	if err != nil {
		log.Fatalf("seed: fetch failed: %v", err)
	}
	entries := seed.Select(rows)
	if err := seed.Write(*out, *listURL, entries); err != nil {
		log.Fatalf("seed: write failed: %v", err)
	}
	log.Printf("seed: wrote %d providers to %s (from %d fetched rows)", len(entries), *out, len(rows))
}
