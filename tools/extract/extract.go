// Package extract obtains a provider binary and runs `nivis gen` to capture
// schema-derived Nix constructors plus a normalized metadata JSON.
//
// Binary acquisition MUST verify before executing. This mirrors the nivis model
// (internal/registry in github.com/wearetechnative/nivis): resolve the version
// via the OpenTofu registry, download, verify the SHA256 checksum, THEN run.
// We never execute an unverified binary.
//
// nivis gen contract (verified):
//
//	nivis gen --provider <BINARY_PATH> [--identity <id>] [--out <dir>]
//	  -> emits <out>/<id>/<resource_type>.nix typed constructors with
//	     doc-comment headers (computed outputs, nested-block shapes).
//	  Takes a LOCAL BINARY PATH, not a registry address.
//	  Available via the nivis flake input: nix run github:wearetechnative/nivis#nivis -- gen ...
//
// Batch extraction MUST be resilient: a single provider failure is skipped and
// logged, never hanging or aborting the run.
//
// See openspec/changes/extraction-pipeline/specs/provider-extraction/spec.md.
package extract

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// DefaultRegistryHost is the OpenTofu registry used for resolution.
const DefaultRegistryHost = "registry.opentofu.org"

// GenArgs builds the `nivis gen` argument vector for a verified binary path.
func GenArgs(binaryPath, identity, outDir string) []string {
	args := []string{"gen", "--provider", binaryPath}
	if identity != "" {
		args = append(args, "--identity", identity)
	}
	if outDir != "" {
		args = append(args, "--out", outDir)
	}
	return args
}

// Platform is one published os/arch target from the registry metadata.
type Platform struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// String renders a platform as the "os/arch" form used by compat.
func (p Platform) String() string { return p.OS + "/" + p.Arch }

// Metadata is the normalized provider-version metadata captured from the
// registry during resolution — the raw material for the compat record.
type Metadata struct {
	Address   string     `json:"address"`
	Namespace string     `json:"namespace"`
	Name      string     `json:"name"`
	Version   string     `json:"version"`
	Protocols []string   `json:"protocols"`
	Platforms []Platform `json:"platforms"`
}

// Result is the per-provider extraction outcome.
type Result struct {
	Address    string   // <namespace>/<name>
	Identity   string   // nivis gen --identity (defaults to name)
	Metadata   Metadata // registry metadata
	OutDir     string   // directory containing <identity>/*.nix
	NixFiles   []string // generated .nix files (absolute paths)
	SkipReason string   // non-empty if extraction was skipped (resilient batch)
}

// Skipped reports whether this provider was skipped rather than extracted.
func (r Result) Skipped() bool { return r.SkipReason != "" }

// Client resolves and verifies provider binaries, then runs nivis gen.
type Client struct {
	http     *http.Client
	host     string // registry host
	nivisBin string // path to the nivis CLI
	cacheDir string // where verified binaries unpack
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient overrides the HTTP client (tests inject a stub).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.http = h } }

// WithRegistryHost overrides the registry host (tests point at a local server).
func WithRegistryHost(host string) Option { return func(c *Client) { c.host = host } }

// New returns an extraction Client. nivisBin is the path to the nivis CLI
// (resolve via PATH or the flake); cacheDir is where verified binaries unpack.
func New(nivisBin, cacheDir string, opts ...Option) *Client {
	c := &Client{
		http:     &http.Client{Timeout: 120 * time.Second},
		host:     DefaultRegistryHost,
		nivisBin: nivisBin,
		cacheDir: cacheDir,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// --- registry resolution (mirrors nivis internal/registry) ---

type versionsResp struct {
	Versions []struct {
		Version   string     `json:"version"`
		Protocols []string   `json:"protocols"`
		Platforms []Platform `json:"platforms"`
	} `json:"versions"`
}

type downloadResp struct {
	DownloadURL string `json:"download_url"`
	ShasumsURL  string `json:"shasums_url"`
	Filename    string `json:"filename"`
}

func (c *Client) getJSON(ctx context.Context, url string, dst interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("GET %s -> HTTP %d: %s", url, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

func (c *Client) getBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s -> HTTP %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// resolve lists versions, selects the latest, and returns the platform download
// plus the captured metadata for the current GOOS/GOARCH.
func (c *Client) resolve(ctx context.Context, namespace, name string) (downloadResp, Metadata, error) {
	base := fmt.Sprintf("https://%s/v1/providers/%s/%s", c.host, namespace, name)

	var vr versionsResp
	if err := c.getJSON(ctx, base+"/versions", &vr); err != nil {
		return downloadResp{}, Metadata{}, fmt.Errorf("list versions: %w", err)
	}
	if len(vr.Versions) == 0 {
		return downloadResp{}, Metadata{}, fmt.Errorf("no versions for %s/%s", namespace, name)
	}
	idx := latestVersionIndex(vr)
	latest := vr.Versions[idx]

	meta := Metadata{
		Address:   namespace + "/" + name,
		Namespace: namespace,
		Name:      name,
		Version:   latest.Version,
		Protocols: latest.Protocols,
		Platforms: latest.Platforms,
	}

	var dr downloadResp
	dlURL := fmt.Sprintf("%s/%s/download/%s/%s", base, latest.Version, runtime.GOOS, runtime.GOARCH)
	if err := c.getJSON(ctx, dlURL, &dr); err != nil {
		return downloadResp{}, meta, fmt.Errorf("resolve download %s %s/%s: %w", latest.Version, runtime.GOOS, runtime.GOARCH, err)
	}
	return dr, meta, nil
}

// latestVersionIndex returns the index of the highest version by numeric
// (semver-ish) ordering — same approach as nivis.
func latestVersionIndex(vr versionsResp) int {
	type kv struct {
		idx int
		key [3]int
		raw string
	}
	out := make([]kv, 0, len(vr.Versions))
	for i, v := range vr.Versions {
		var key [3]int
		parts := strings.FieldsFunc(v.Version, func(r rune) bool { return r < '0' || r > '9' })
		for j := 0; j < 3 && j < len(parts); j++ {
			n := 0
			for _, ch := range parts[j] {
				n = n*10 + int(ch-'0')
			}
			key[j] = n
		}
		out = append(out, kv{idx: i, key: key, raw: v.Version})
	}
	sort.Slice(out, func(i, j int) bool {
		for k := 0; k < 3; k++ {
			if out[i].key[k] != out[j].key[k] {
				return out[i].key[k] < out[j].key[k]
			}
		}
		return out[i].raw < out[j].raw
	})
	return out[len(out)-1].idx
}

// --- verification (mirrors nivis verify-before-execute) ---

func parseShasums(body string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(body, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 {
			out[fields[1]] = strings.ToLower(fields[0])
		}
	}
	return out
}

// verifySHA256 requires sha256(data) to equal the checksum listed for filename.
// A mismatch (or missing entry) is an error; callers MUST NOT use the bytes —
// nor execute the binary — when verifySHA256 returns non-nil.
func verifySHA256(data []byte, filename string, sums map[string]string) error {
	want, ok := sums[filename]
	if !ok {
		return fmt.Errorf("no checksum for %q in SHA256SUMS", filename)
	}
	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); got != want {
		return fmt.Errorf("checksum mismatch for %q: expected %s, got %s", filename, want, got)
	}
	return nil
}

// VerifyArchive is the public verify-before-execute gate used by Fetch and
// exercised directly by tests: it checks the archive bytes against the SHA256SUMS
// body, returning an error (and refusing to proceed) on any mismatch.
func VerifyArchive(zipData []byte, filename, shasumsBody string) error {
	return verifySHA256(zipData, filename, parseShasums(shasumsBody))
}

// extractBinary unzips the verified archive into dir and returns the executable
// path. Only the terraform-provider-* entry is extracted.
func extractBinary(dir string, zipData []byte) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return "", fmt.Errorf("open zip: %w", err)
	}
	for _, f := range zr.File {
		name := filepath.Base(f.Name)
		if !strings.HasPrefix(name, "terraform-provider-") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		out := filepath.Join(dir, name)
		dst, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			rc.Close()
			return "", err
		}
		_, copyErr := io.Copy(dst, rc) //nolint:gosec // archive is checksum-verified before this point
		rc.Close()
		dst.Close()
		if copyErr != nil {
			return "", copyErr
		}
		return out, nil
	}
	return "", fmt.Errorf("no terraform-provider-* executable in archive")
}

// Fetch resolves, downloads, SHA256-verifies, and unpacks a provider binary,
// returning its local path plus the captured registry metadata. The binary is
// NEVER written to disk before its archive passes verification.
func (c *Client) Fetch(ctx context.Context, namespace, name string) (binaryPath string, meta Metadata, err error) {
	dr, meta, err := c.resolve(ctx, namespace, name)
	if err != nil {
		return "", meta, err
	}
	zipData, err := c.getBytes(ctx, dr.DownloadURL)
	if err != nil {
		return "", meta, fmt.Errorf("download %s: %w", dr.Filename, err)
	}
	sumsBody, err := c.getBytes(ctx, dr.ShasumsURL)
	if err != nil {
		return "", meta, fmt.Errorf("download SHA256SUMS: %w", err)
	}
	if err := VerifyArchive(zipData, dr.Filename, string(sumsBody)); err != nil {
		return "", meta, err // verify-before-execute: refuse to unpack/run
	}
	dir := filepath.Join(c.cacheDir, namespace, name, meta.Version, runtime.GOOS+"_"+runtime.GOARCH)
	bin, err := extractBinary(dir, zipData)
	if err != nil {
		return "", meta, err
	}
	return bin, meta, nil
}

// RunGen executes `nivis gen` on a verified binary path, emitting .nix
// constructors under outDir/<identity>/, and returns the generated .nix files.
func (c *Client) RunGen(ctx context.Context, binaryPath, identity, outDir string) ([]string, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, c.nivisBin, GenArgs(binaryPath, identity, outDir)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("nivis gen: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	idDir := filepath.Join(outDir, identity)
	entries, err := os.ReadDir(idDir)
	if err != nil {
		return nil, fmt.Errorf("read gen output %s: %w", idDir, err)
	}
	var nixFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".nix") {
			nixFiles = append(nixFiles, filepath.Join(idDir, e.Name()))
		}
	}
	sort.Strings(nixFiles)
	if len(nixFiles) == 0 {
		return nil, fmt.Errorf("nivis gen produced no .nix files in %s", idDir)
	}
	return nixFiles, nil
}

// Extract runs the full per-provider pipeline: Fetch (verify-before-execute) →
// RunGen → write a normalized metadata.json. identity defaults to the provider
// name. The metadata.json sits beside the <identity>/ constructor directory.
func (c *Client) Extract(ctx context.Context, namespace, name, outRoot string) (Result, error) {
	identity := name
	address := namespace + "/" + name
	bin, meta, err := c.Fetch(ctx, namespace, name)
	if err != nil {
		return Result{Address: address, Identity: identity}, err
	}
	outDir := filepath.Join(outRoot, namespace, name, meta.Version)
	nixFiles, err := c.RunGen(ctx, bin, identity, outDir)
	if err != nil {
		return Result{Address: address, Identity: identity, Metadata: meta, OutDir: outDir}, err
	}
	// Persist the normalized metadata next to the constructors.
	metaPath := filepath.Join(outDir, "metadata.json")
	if mb, mErr := json.MarshalIndent(meta, "", "  "); mErr == nil {
		_ = os.WriteFile(metaPath, append(mb, '\n'), 0o644)
	}
	return Result{
		Address:  address,
		Identity: identity,
		Metadata: meta,
		OutDir:   outDir,
		NixFiles: nixFiles,
	}, nil
}

// Logger receives one line per provider in a batch (skip reasons, successes).
type Logger func(format string, args ...interface{})

// Batch extracts every address resiliently: a single provider's failure is
// recorded as a skip (with reason) and logged, and MUST NOT abort the run or
// hang. The returned slice has one Result per input address, in order.
func (c *Client) Batch(ctx context.Context, addresses []string, outRoot string, log Logger) []Result {
	if log == nil {
		log = func(string, ...interface{}) {}
	}
	results := make([]Result, 0, len(addresses))
	for _, addr := range addresses {
		parts := strings.SplitN(addr, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			log("skip %s: invalid address", addr)
			results = append(results, Result{Address: addr, SkipReason: "invalid address"})
			continue
		}
		// Per-provider timeout so one bad provider can never hang the batch.
		provCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		res, err := c.Extract(provCtx, parts[0], parts[1], outRoot)
		cancel()
		if err != nil {
			log("skip %s: %v", addr, err)
			res.SkipReason = err.Error()
			results = append(results, res)
			continue
		}
		log("ok   %s: %d constructor(s) @ %s", addr, len(res.NixFiles), res.Metadata.Version)
		results = append(results, res)
	}
	return results
}
