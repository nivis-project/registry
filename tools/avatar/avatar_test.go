package avatar

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func square(c color.Color) image.Image {
	m := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			m.Set(x, y, c)
		}
	}
	return m
}

func pngBytes(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, square(color.RGBA{10, 20, 30, 255})); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func jpegBytes(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := jpeg.Encode(&b, square(color.RGBA{200, 100, 50, 255}), nil); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func isPNG(b []byte) bool { return bytes.HasPrefix(b, []byte("\x89PNG\r\n\x1a\n")) }

// serve returns a client pointed at a stub, plus a counter of requests per path.
func serve(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New(WithHTTPClient(&http.Client{Transport: rewrite{srv.URL}}))
}

// rewrite sends every request to the stub, whatever host the code asked for.
type rewrite struct{ base string }

func (r rewrite) RoundTrip(req *http.Request) (*http.Response, error) {
	u := *req.URL
	u.Scheme = "http"
	u.Host = strings.TrimPrefix(r.base, "http://")
	req2 := req.Clone(req.Context())
	req2.URL = &u
	return http.DefaultTransport.RoundTrip(req2)
}

// TestNormaliseConvertsJPEG is the case that bites in production: the source
// serves whatever was uploaded regardless of the .png asked for, and five of
// the thirty-six real avatars come back as JPEG.
func TestNormaliseConvertsJPEG(t *testing.T) {
	out, err := Normalise(jpegBytes(t))
	if err != nil {
		t.Fatalf("Normalise: %v", err)
	}
	if !isPNG(out) {
		t.Error("a JPEG source must be stored as PNG")
	}
}

func TestNormaliseRejectsNonImage(t *testing.T) {
	if _, err := Normalise([]byte("<html>not an image</html>")); err == nil {
		t.Error("a non-image payload must fail rather than be stored")
	}
}

func TestFetchAllDeduplicatesByOwner(t *testing.T) {
	hits := map[string]int{}
	c := serve(t, func(w http.ResponseWriter, r *http.Request) {
		hits[r.URL.Path]++
		w.Write(pngBytes(t))
	})

	// Three providers from one namespace, plus a module owner.
	refs := c.FetchAll(context.Background(),
		[]string{"hashicorp", "hashicorp", "hashicorp", "wearetechnative"},
		t.TempDir(), nil)

	if len(refs) != 2 {
		t.Fatalf("got %d avatars, want 2 (one per distinct owner)", len(refs))
	}
	if hits["/hashicorp.png"] != 1 {
		t.Errorf("fetched hashicorp %d times, want once", hits["/hashicorp.png"])
	}
}

func TestFetchAllWritesAndReferences(t *testing.T) {
	root := t.TempDir()
	c := serve(t, func(w http.ResponseWriter, r *http.Request) { w.Write(jpegBytes(t)) })

	refs := c.FetchAll(context.Background(), []string{"ovh"}, root, nil)

	if refs["ovh"] != "avatars/ovh.png" {
		t.Errorf("reference = %q, want a contract-relative avatars/ovh.png", refs["ovh"])
	}
	stored, err := os.ReadFile(filepath.Join(root, "registry", "avatars", "ovh.png"))
	if err != nil {
		t.Fatalf("avatar not written: %v", err)
	}
	if !isPNG(stored) {
		t.Error("the stored file must be PNG whatever the source sent")
	}
}

// TestFetchAllSkipsFailures: the source being unreachable must not cost us the
// contract. A skipped owner has no reference, so its entries render without an
// avatar rather than with a broken image.
func TestFetchAllSkipsFailures(t *testing.T) {
	c := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "missing") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(pngBytes(t))
	})

	var logs []string
	refs := c.FetchAll(context.Background(), []string{"missing", "present"}, t.TempDir(),
		func(f string, a ...interface{}) { logs = append(logs, f) })

	if _, ok := refs["missing"]; ok {
		t.Error("an owner with no avatar must have no reference")
	}
	if _, ok := refs["present"]; !ok {
		t.Error("one failure must not stop the others")
	}
	if len(logs) != 1 {
		t.Errorf("the skip should be logged once, got %d lines", len(logs))
	}
}

func TestFetchAllSkipsUndecodableResponse(t *testing.T) {
	c := serve(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<!doctype html><title>rate limited</title>"))
	})
	refs := c.FetchAll(context.Background(), []string{"owner"}, t.TempDir(), nil)
	if len(refs) != 0 {
		t.Error("a response that is not an image must be skipped, not stored")
	}
}

// TestOverrideSeam: the override map is empty by design, but the lookup must
// consult it, so adding one later is a data change rather than a rework.
func TestOverrideSeam(t *testing.T) {
	if len(Overrides) != 0 {
		t.Fatal("Overrides ships empty; this change does not add any")
	}
	local := filepath.Join(t.TempDir(), "brand.png")
	if err := os.WriteFile(local, pngBytes(t), 0o644); err != nil {
		t.Fatal(err)
	}
	Overrides["acme"] = local
	t.Cleanup(func() { delete(Overrides, "acme") })

	// No stub server: reaching the network here would mean the override was
	// ignored.
	c := New(WithHTTPClient(&http.Client{Transport: failTransport{}}))
	refs := c.FetchAll(context.Background(), []string{"acme"}, t.TempDir(), nil)
	if refs["acme"] != "avatars/acme.png" {
		t.Error("an override must be used instead of fetching")
	}
}

type failTransport struct{}

func (failTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, http.ErrUseLastResponse
}

func TestSourceURLUsesTheOwner(t *testing.T) {
	if got := SourceURL("wearetechnative"); !strings.Contains(got, "/wearetechnative.png") {
		t.Errorf("SourceURL = %q", got)
	}
}
