package extract

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenArgs(t *testing.T) {
	got := GenArgs("/bin/prov", "alpha", "/out")
	want := []string{"gen", "--provider", "/bin/prov", "--identity", "alpha", "--out", "/out"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("GenArgs = %v, want %v", got, want)
	}
	// identity/out omitted when empty.
	got = GenArgs("/bin/prov", "", "")
	if strings.Join(got, " ") != "gen --provider /bin/prov" {
		t.Errorf("GenArgs minimal = %v", got)
	}
}

func TestPlatformString(t *testing.T) {
	if (Platform{OS: "linux", Arch: "amd64"}).String() != "linux/amd64" {
		t.Error("Platform.String wrong")
	}
}

// makeZip builds an in-memory provider zip containing a terraform-provider-<name>
// executable with the given content, returning the zip bytes.
func makeZip(t *testing.T, provName, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("terraform-provider-" + provName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func sha256hex(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// TestVerifyArchiveMatch confirms a correct checksum passes.
func TestVerifyArchiveMatch(t *testing.T) {
	zipData := makeZip(t, "random", "fake-binary")
	const fname = "terraform-provider-random.zip"
	sums := fmt.Sprintf("%s  %s\n", sha256hex(zipData), fname)
	if err := VerifyArchive(zipData, fname, sums); err != nil {
		t.Errorf("VerifyArchive should pass: %v", err)
	}
}

// TestVerifyArchiveMismatch is the spec scenario "Checksum mismatch aborts
// extraction": a wrong checksum must error so the binary is never executed.
func TestVerifyArchiveMismatch(t *testing.T) {
	zipData := makeZip(t, "random", "fake-binary")
	const fname = "terraform-provider-random.zip"
	sums := fmt.Sprintf("%s  %s\n", sha256hex([]byte("DIFFERENT")), fname)
	err := VerifyArchive(zipData, fname, sums)
	if err == nil {
		t.Fatal("VerifyArchive must error on checksum mismatch")
	}
	if !strings.Contains(err.Error(), "mismatch") {
		t.Errorf("error should mention mismatch, got %v", err)
	}
}

func TestVerifyArchiveMissingEntry(t *testing.T) {
	zipData := makeZip(t, "random", "x")
	if err := VerifyArchive(zipData, "terraform-provider-random.zip", "abc  other.zip\n"); err == nil {
		t.Error("VerifyArchive must error when filename absent from SHA256SUMS")
	}
}

// fakeRegistry stands up a local OpenTofu-shaped registry serving a single
// provider version, its download/shasums, and the zip — so Fetch can be tested
// end-to-end without network and with controllable checksums.
type fakeRegistry struct {
	srv         *httptest.Server
	provName    string
	zipData     []byte
	corruptSums bool // serve a wrong checksum to exercise the abort path
}

func newFakeRegistry(t *testing.T, provName string, corruptSums bool) *fakeRegistry {
	t.Helper()
	fr := &fakeRegistry{provName: provName, zipData: makeZip(t, provName, "BINARY:"+provName), corruptSums: corruptSums}
	mux := http.NewServeMux()
	base := "/v1/providers/acme/" + provName
	mux.HandleFunc(base+"/versions", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"versions":[{"version":"1.2.3","protocols":["6.0"],"platforms":[{"os":"linux","arch":"amd64"},{"os":"darwin","arch":"arm64"}]}]}`)
	})
	mux.HandleFunc(base+"/1.2.3/download/", func(w http.ResponseWriter, r *http.Request) {
		u := "http://" + r.Host
		fmt.Fprintf(w, `{"download_url":%q,"shasums_url":%q,"filename":"terraform-provider-%s.zip"}`,
			u+"/dl/prov.zip", u+"/dl/SHA256SUMS", provName)
	})
	mux.HandleFunc("/dl/prov.zip", func(w http.ResponseWriter, r *http.Request) {
		w.Write(fr.zipData)
	})
	mux.HandleFunc("/dl/SHA256SUMS", func(w http.ResponseWriter, r *http.Request) {
		sum := sha256hex(fr.zipData)
		if fr.corruptSums {
			sum = sha256hex([]byte("not the real bytes"))
		}
		fmt.Fprintf(w, "%s  terraform-provider-%s.zip\n", sum, provName)
	})
	fr.srv = httptest.NewServer(mux)
	t.Cleanup(fr.srv.Close)
	return fr
}

func (fr *fakeRegistry) host() string {
	return strings.TrimPrefix(fr.srv.URL, "http://")
}

// clientForFake wires a Client to talk plain HTTP to the fake registry (the
// production code uses https; the transport rewrites the scheme).
func clientForFake(fr *fakeRegistry, nivisBin, cacheDir string) *Client {
	rt := &schemeRewriteTransport{base: http.DefaultTransport}
	return New(nivisBin, cacheDir,
		WithRegistryHost(fr.host()),
		WithHTTPClient(&http.Client{Transport: rt}),
	)
}

// schemeRewriteTransport downgrades https->http so the test can hit httptest.
type schemeRewriteTransport struct{ base http.RoundTripper }

func (t *schemeRewriteTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.Scheme == "https" {
		r.URL.Scheme = "http"
	}
	return t.base.RoundTrip(r)
}

// TestFetchVerifyBeforeExecute_Mismatch is the end-to-end abort path: a corrupt
// SHA256SUMS must make Fetch fail and leave no binary on disk.
func TestFetchVerifyBeforeExecute_Mismatch(t *testing.T) {
	fr := newFakeRegistry(t, "random", true /*corrupt*/)
	cache := t.TempDir()
	c := clientForFake(fr, "/nonexistent-nivis", cache)

	_, _, err := c.Fetch(context.Background(), "acme", "random")
	if err == nil {
		t.Fatal("Fetch must fail when the archive checksum does not match")
	}
	if !strings.Contains(err.Error(), "mismatch") {
		t.Errorf("expected checksum mismatch error, got %v", err)
	}
	// And nothing was unpacked.
	if entries, _ := os.ReadDir(cache); len(entries) != 0 {
		t.Errorf("no binary should be written on verify failure, found %v", entries)
	}
}

// TestFetchVerifyBeforeExecute_OK confirms a matching checksum lets Fetch unpack
// the binary and return the captured metadata (protocols/platforms).
func TestFetchVerifyBeforeExecute_OK(t *testing.T) {
	fr := newFakeRegistry(t, "random", false)
	cache := t.TempDir()
	c := clientForFake(fr, "/nonexistent-nivis", cache)

	bin, meta, err := c.Fetch(context.Background(), "acme", "random")
	if err != nil {
		t.Fatalf("Fetch should succeed: %v", err)
	}
	if _, statErr := os.Stat(bin); statErr != nil {
		t.Errorf("binary not on disk: %v", statErr)
	}
	if meta.Version != "1.2.3" {
		t.Errorf("version = %q, want 1.2.3", meta.Version)
	}
	if len(meta.Protocols) != 1 || meta.Protocols[0] != "6.0" {
		t.Errorf("protocols = %v, want [6.0]", meta.Protocols)
	}
	if len(meta.Platforms) != 2 {
		t.Errorf("platforms = %v, want 2 entries", meta.Platforms)
	}
}

// nivisOnPath returns the nivis binary path if available, else "".
func nivisOnPath() string {
	p, err := exec.LookPath("nivis")
	if err != nil {
		return ""
	}
	return p
}

// buildFakeProvider builds a nivis fake provider (cmd/provider-<which>) from a
// nivis source checkout located via NIVIS_SRC; returns ("", false) to skip.
func buildFakeProvider(t *testing.T, which string) (string, bool) {
	t.Helper()
	src := os.Getenv("NIVIS_SRC")
	if src == "" {
		return "", false
	}
	if _, err := os.Stat(filepath.Join(src, "cmd", "provider-"+which)); err != nil {
		return "", false
	}
	bin := filepath.Join(t.TempDir(), "provider-"+which)
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/provider-"+which)
	cmd.Dir = src
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Logf("build fake provider failed: %v\n%s", err, out)
		return "", false
	}
	return bin, true
}

// TestRunGen_HermeticFakeProvider runs nivis gen on a locally built fake
// provider — the hermetic fixture (no network, no creds). Skipped unless both
// nivis (PATH) and a nivis source checkout (NIVIS_SRC) are available.
func TestRunGen_HermeticFakeProvider(t *testing.T) {
	nivis := nivisOnPath()
	if nivis == "" {
		t.Skip("nivis not on PATH; run inside `nix develop`")
	}
	bin, ok := buildFakeProvider(t, "beta")
	if !ok {
		t.Skip("NIVIS_SRC not set or fake provider not buildable")
	}
	c := New(nivis, t.TempDir())
	out := t.TempDir()
	files, err := c.RunGen(context.Background(), bin, "beta", out)
	if err != nil {
		t.Fatalf("RunGen: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected at least one .nix constructor")
	}
	body, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "mkResource") {
		t.Errorf("constructor missing mkResource:\n%s", body)
	}
}

// TestBatchResilient is the spec scenario "One bad provider does not stop the
// batch": a mix of a bad address and an unresolvable one must yield one Result
// per input with each marked skipped, never aborting.
func TestBatchResilient(t *testing.T) {
	c := New("/nonexistent-nivis", t.TempDir(), WithRegistryHost("127.0.0.1:0"),
		WithHTTPClient(&http.Client{Transport: &schemeRewriteTransport{base: http.DefaultTransport}}))
	var logs []string
	logger := func(f string, a ...interface{}) { logs = append(logs, fmt.Sprintf(f, a...)) }

	addrs := []string{"not-an-address", "acme/wont-resolve"}
	results := c.Batch(context.Background(), addrs, t.TempDir(), logger)
	if len(results) != len(addrs) {
		t.Fatalf("Batch returned %d results, want %d", len(results), len(addrs))
	}
	if !results[0].Skipped() {
		t.Error("invalid address should be skipped with a reason")
	}
	if !results[1].Skipped() {
		t.Error("unresolvable provider should be skipped with a reason")
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 log lines, got %d: %v", len(logs), logs)
	}
}

func TestLatestVersionIndex(t *testing.T) {
	type ver = struct {
		Version   string     `json:"version"`
		Protocols []string   `json:"protocols"`
		Platforms []Platform `json:"platforms"`
	}
	vr := versionsResp{Versions: []ver{
		{Version: "1.2.0"},
		{Version: "1.10.0"},
		{Version: "1.9.0"},
	}}
	if got := vr.Versions[latestVersionIndex(vr)].Version; got != "1.10.0" {
		t.Errorf("latest = %q, want 1.10.0 (numeric, not lexical)", got)
	}
}
