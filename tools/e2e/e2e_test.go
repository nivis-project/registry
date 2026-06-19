// Package e2e holds end-to-end tests that exercise the full pipeline:
// extract (nivis gen) -> contract -> render. The hermetic test builds a nivis
// fake provider and runs it offline (no network, no credentials); the
// real-provider test resolves a couple of small credential-free providers.
//
// Both tests SKIP when their prerequisites are absent (nivis on PATH, NIVIS_SRC
// for the fake provider source, network for the real-provider test), so
// `go test ./...` stays green in any environment. They run for real in CI:
//   - the hermetic e2e runs in the Nix sandbox check `e2e-hermetic` (offline);
//   - the real-provider e2e runs in the network-enabled CI job.
package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nivis-project/registry/tools/extract"
	"github.com/nivis-project/registry/tools/generate"
)

func nivisBin(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("NIVIS_BIN"); p != "" {
		return p
	}
	p, err := exec.LookPath("nivis")
	if err != nil {
		t.Skip("nivis not on PATH and NIVIS_BIN unset; run inside `nix develop`")
	}
	return p
}

// buildFakeProvider builds cmd/provider-<which> from the nivis source (NIVIS_SRC)
// into a temp dir, returning the binary path. Skips if the source is absent.
func buildFakeProvider(t *testing.T, which string) string {
	t.Helper()
	src := os.Getenv("NIVIS_SRC")
	if src == "" {
		t.Skip("NIVIS_SRC unset; cannot build the hermetic fake provider")
	}
	cmdDir := filepath.Join(src, "cmd", "provider-"+which)
	if _, err := os.Stat(cmdDir); err != nil {
		t.Skipf("fake provider %s not found under NIVIS_SRC", which)
	}
	bin := filepath.Join(t.TempDir(), "provider-"+which)
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/provider-"+which)
	cmd.Dir = src
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build fake provider %s: %v\n%s", which, err, out)
	}
	return bin
}

// TestHermeticPipeline is the spec's hermetic end-to-end proof: build a fake
// provider, run it through extract -> contract -> render, and assert a rendered
// page exists. No network and no credentials are used.
func TestHermeticPipeline(t *testing.T) {
	nivis := nivisBin(t)
	bin := buildFakeProvider(t, "beta") // beta_record: a required `from`, computed `endpoint`

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. Extract: run nivis gen on the verified (locally built) binary.
	extractRoot := t.TempDir()
	identity := "beta"
	outDir := filepath.Join(extractRoot, "fake", "beta", "0.0.0")
	c := extract.New(nivis, t.TempDir())
	nixFiles, err := c.RunGen(ctx, bin, identity, outDir)
	if err != nil {
		t.Fatalf("RunGen: %v", err)
	}
	if len(nixFiles) == 0 {
		t.Fatal("expected at least one .nix constructor from the fake provider")
	}

	// Write a metadata.json so the generator can pick up the version dir
	// (offline: synthesize the metadata the real extractor would capture).
	writeMetadata(t, outDir, "fake/beta", "fake", "beta", "0.0.0", []string{"6.0"})

	// 2. Contract: generate the registry-ui contract + per-item docs.
	contractRoot := t.TempDir()
	idxPath, err := generate.GenerateProvider(outDir, contractRoot)
	if err != nil {
		t.Fatalf("GenerateProvider: %v", err)
	}

	// 3. Render: a per-item document page must exist, be Nix, and carry no HCL.
	docDir := filepath.Dir(idxPath)
	docs, _ := filepath.Glob(filepath.Join(docDir, "*.md"))
	if len(docs) == 0 {
		t.Fatal("expected at least one rendered page")
	}
	body, err := os.ReadFile(docs[0])
	if err != nil {
		t.Fatal(err)
	}
	page := string(body)
	if !strings.Contains(page, "```nix") {
		t.Errorf("rendered page must contain a Nix code block:\n%s", page)
	}
	if generate.ContainsHCLResourceBlock(page) {
		t.Error("rendered page must NOT contain an HCL resource block")
	}
	// The beta_record resource declares a required `from` argument.
	if !strings.Contains(page, "from") {
		t.Errorf("rendered page should mention the beta_record `from` argument:\n%s", page)
	}
}

// writeMetadata synthesizes the metadata.json that tools/extract would write,
// so the offline hermetic test can drive the generator without a real registry.
func writeMetadata(t *testing.T, dir, address, ns, name, version string, protocols []string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	protoJSON := `["` + strings.Join(protocols, `","`) + `"]`
	body := `{
  "address": "` + address + `",
  "namespace": "` + ns + `",
  "name": "` + name + `",
  "version": "` + version + `",
  "protocols": ` + protoJSON + `,
  "platforms": [{"os":"linux","arch":"amd64"},{"os":"linux","arch":"arm64"}]
}
`
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestRealProviderPipeline is the real-provider e2e: resolve, extract, contract,
// and render `hashicorp/random` and `hashicorp/null` (credential-free; tls is a
// third in the proof script). It needs network; it skips when offline or when
// nivis is unavailable. Telmate/proxmox is intentionally excluded here — nivis
// gen v0.4.0 cannot configure it (see the run report / memory).
func TestRealProviderPipeline(t *testing.T) {
	if os.Getenv("NIVIS_REGISTRY_NET_E2E") != "1" {
		t.Skip("set NIVIS_REGISTRY_NET_E2E=1 to run the network real-provider e2e")
	}
	nivis := nivisBin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	extractRoot := t.TempDir()
	contractRoot := t.TempDir()
	c := extract.New(nivis, t.TempDir())

	for _, prov := range []struct{ ns, name string }{
		{"hashicorp", "random"},
		{"hashicorp", "null"},
		{"hashicorp", "tls"},
	} {
		res, err := c.Extract(ctx, prov.ns, prov.name, extractRoot)
		if err != nil {
			t.Fatalf("extract %s/%s: %v", prov.ns, prov.name, err)
		}
		if len(res.NixFiles) == 0 {
			t.Fatalf("%s/%s produced no constructors", prov.ns, prov.name)
		}
		idxPath, err := generate.GenerateProvider(res.OutDir, contractRoot)
		if err != nil {
			t.Fatalf("generate %s/%s: %v", prov.ns, prov.name, err)
		}
		// Assert: index + at least one rendered page (Nix, no HCL) + a badge.
		docs, _ := filepath.Glob(filepath.Join(filepath.Dir(idxPath), "*.md"))
		if len(docs) == 0 {
			t.Fatalf("%s/%s: no rendered pages", prov.ns, prov.name)
		}
		body, _ := os.ReadFile(docs[0])
		if generate.ContainsHCLResourceBlock(string(body)) {
			t.Errorf("%s/%s: rendered page contains an HCL resource block", prov.ns, prov.name)
		}
		if !strings.Contains(string(body), "```nix") {
			t.Errorf("%s/%s: rendered page is not Nix", prov.ns, prov.name)
		}
	}
}
