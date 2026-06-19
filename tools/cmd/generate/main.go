// Command generate walks an extraction-output tree and emits the static
// registry-ui-compatible JSON contract (per-provider index.json with the compat
// badge + per-item Nix-rendered documents).
//
//	go run ./cmd/generate -extract ../extract-out -contract ..
//
// The contract is written under <contract>/registry/docs/providers/...
package main

import (
	"flag"
	"log"
	"os"

	"github.com/nivis-project/registry/tools/generate"
)

func main() {
	extractRoot := flag.String("extract", "extract-out", "extraction-output root (from tools/extract)")
	contractRoot := flag.String("contract", ".", "contract root (writes <root>/registry/docs/providers/...)")
	flag.Parse()

	idxs, err := generate.GenerateAll(*extractRoot, *contractRoot, log.Printf)
	if err != nil {
		log.Fatalf("generate: %v", err)
	}
	log.Printf("generate: emitted %d provider contract(s)", len(idxs))
	if len(idxs) == 0 {
		os.Exit(1)
	}
}
