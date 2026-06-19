// Command extract runs the extraction pipeline over a set of provider addresses:
// verify-before-execute fetch → nivis gen → normalized metadata. It is resilient
// (a single provider's failure is skipped and logged, never aborting the run).
//
//	go run ./cmd/extract -nivis "$(command -v nivis)" -out ../extract-out random null Telmate/proxmox
//	go run ./cmd/extract -nivis ... -seed ../seed.json   # use seed.json addresses
//
// Provider addresses default to the OpenTofu registry; bare "<name>" is treated
// as "hashicorp/<name>" for convenience.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nivis-project/registry/tools/compat"
	"github.com/nivis-project/registry/tools/extract"
	"github.com/nivis-project/registry/tools/seed"
)

func main() {
	nivisBin := flag.String("nivis", "nivis", "path to the nivis CLI")
	outRoot := flag.String("out", "extract-out", "output root for generated constructors + metadata")
	cacheDir := flag.String("cache", ".cache/providers", "cache dir for verified binaries")
	seedPath := flag.String("seed", "", "read provider addresses from a seed.json instead of args")
	reportPath := flag.String("report", "", "write a JSON coverage report to this path")
	flag.Parse()

	addrs := normalizeAddrs(flag.Args())
	if *seedPath != "" {
		fromSeed, err := readSeed(*seedPath)
		if err != nil {
			log.Fatalf("extract: read seed %s: %v", *seedPath, err)
		}
		addrs = fromSeed
	}
	if len(addrs) == 0 {
		log.Fatal("extract: no provider addresses (pass args or -seed)")
	}

	c := extract.New(*nivisBin, *cacheDir)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	results := c.Batch(ctx, addrs, *outRoot, log.Printf)

	// Compute + persist a compat record per successful extraction, build a
	// coverage report, and summarize.
	report := coverageReport{Total: len(results)}
	for _, r := range results {
		if r.Skipped() {
			report.Skipped++
			report.Providers = append(report.Providers, coverageEntry{
				Address: r.Address, Extracted: false, Reason: r.SkipReason,
			})
			continue
		}
		report.OK++
		rec := compat.Compute(r.Metadata, true)
		recPath := r.OutDir + "/compat.json"
		if b, err := json.MarshalIndent(rec, "", "  "); err == nil {
			_ = os.WriteFile(recPath, append(b, '\n'), 0o644)
		}
		report.Providers = append(report.Providers, coverageEntry{
			Address: r.Address, Extracted: true, Version: r.Metadata.Version, Constructors: len(r.NixFiles),
		})
		log.Printf("compat %s: tier=%q protocols=%v arches=%v e2e=%s",
			r.Address, rec.Tier, rec.Protocols, rec.Architectures, rec.E2E)
	}
	log.Printf("extract: %d ok, %d skipped, %d total", report.OK, report.Skipped, report.Total)

	if *reportPath != "" {
		if b, err := json.MarshalIndent(report, "", "  "); err == nil {
			if werr := os.WriteFile(*reportPath, append(b, '\n'), 0o644); werr != nil {
				log.Printf("extract: could not write report %s: %v", *reportPath, werr)
			} else {
				log.Printf("extract: coverage report -> %s", *reportPath)
			}
		}
	}
	if report.OK == 0 {
		os.Exit(1)
	}
}

// coverageReport is the auditable outcome of a (possibly full-catalog) run.
type coverageReport struct {
	Total     int             `json:"total"`
	OK        int             `json:"ok"`
	Skipped   int             `json:"skipped"`
	Providers []coverageEntry `json:"providers"`
}

type coverageEntry struct {
	Address      string `json:"address"`
	Extracted    bool   `json:"extracted"`
	Version      string `json:"version,omitempty"`
	Constructors int    `json:"constructors,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

// normalizeAddrs expands a bare "<name>" to "hashicorp/<name>".
func normalizeAddrs(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if !strings.Contains(a, "/") {
			a = "hashicorp/" + a
		}
		out = append(out, a)
	}
	return out
}

func readSeed(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m seed.Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return seed.AddressList(m.Providers), nil
}
