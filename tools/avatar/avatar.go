// Package avatar obtains the brand avatar of whoever publishes a provider,
// module or blueprint.
//
// Avatars are fetched HERE, when the contract is generated, and served from the
// registry's own origin. Hot-linking them would put a request to a third party
// on every card of every page view, leaking each visitor and breaking whenever
// that service is slow or unreachable. The fonts are self-hosted for the same
// reason, and the visual-identity spec requires it.
//
// Keyed by OWNER, not by entry: 59 providers collapse to 35 avatars, and modules
// and blueprints reuse the same files.
//
// See openspec/changes/catalog-enrichment/specs/owner-avatars/spec.md.
package avatar

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg" // registers JPEG with image.Decode
	"image/png"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"
)

// Dir is where avatars live inside the contract, relative to its root.
const Dir = "registry/avatars"

// Size is the edge length requested from the source. Avatars are shown at 28px
// in lists and 36px on a detail page, so 80px covers a 2x display.
const Size = 80

// Overrides maps an owner to a local image file to use instead of fetching one.
//
// Deliberately empty. Some marks are wordmarks that read poorly at 28px however
// they are presented, and a hand-maintained replacement is the eventual answer.
// The seam exists so that adding one later is a data change: nothing in the
// contract or the SPA has to learn where an avatar came from.
var Overrides = map[string]string{}

// SourceURL returns where an owner's avatar is fetched from. The namespace of a
// provider is its GitHub organisation by construction, because providers are
// published from github.com/<namespace>/terraform-provider-<name>.
func SourceURL(owner string) string {
	return fmt.Sprintf("https://github.com/%s.png?size=%d", owner, Size)
}

// Ref is the contract-relative reference stored on a catalogue entry.
func Ref(owner string) string { return path.Join("avatars", owner+".png") }

// Client fetches and stores avatars.
type Client struct {
	http *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient overrides the HTTP client (tests inject a stub).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.http = h } }

// New returns an avatar Client.
func New(opts ...Option) *Client {
	c := &Client{http: &http.Client{Timeout: 30 * time.Second}}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Normalise decodes an avatar and re-encodes it as PNG.
//
// The source serves whatever was uploaded regardless of the extension asked
// for: five of the thirty-six current avatars come back as JPEG despite the
// ".png" in the URL. Normalising means a consumer never has to negotiate
// formats, and it needs nothing outside the standard library.
func Normalise(data []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode avatar: %w", err)
	}
	var out bytes.Buffer
	if err := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(&out, img); err != nil {
		return nil, fmt.Errorf("encode avatar: %w", err)
	}
	return out.Bytes(), nil
}

// fetch retrieves one owner's avatar bytes, honouring an override first.
func (c *Client) fetch(ctx context.Context, owner string) ([]byte, error) {
	if local, ok := Overrides[owner]; ok {
		return os.ReadFile(local)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, SourceURL(owner), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s -> HTTP %d", SourceURL(owner), resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 4<<20))
}

// Logger receives one line per owner that could not be fetched.
type Logger func(format string, args ...interface{})

// FetchAll obtains an avatar for each owner and writes it under
// <contractRoot>/registry/avatars/, returning owner -> contract-relative
// reference for those that succeeded.
//
// An owner that cannot be fetched is SKIPPED, never fatal. The source being
// unreachable must not cost us the contract, the same rule the provider batch
// already follows. A skipped owner simply has no reference, so a consumer
// renders its entries without an avatar rather than with a broken image.
func (c *Client) FetchAll(ctx context.Context, owners []string, contractRoot string, log Logger) map[string]string {
	if log == nil {
		log = func(string, ...interface{}) {}
	}
	dir := filepath.Join(contractRoot, filepath.FromSlash(Dir))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log("warn avatars: %v", err)
		return map[string]string{}
	}

	refs := make(map[string]string, len(owners))
	for _, owner := range dedupe(owners) {
		raw, err := c.fetch(ctx, owner)
		if err != nil {
			log("skip avatar %s: %v", owner, err)
			continue
		}
		pngData, err := Normalise(raw)
		if err != nil {
			log("skip avatar %s: %v", owner, err)
			continue
		}
		if err := os.WriteFile(filepath.Join(dir, owner+".png"), pngData, 0o644); err != nil {
			log("skip avatar %s: %v", owner, err)
			continue
		}
		refs[owner] = Ref(owner)
	}
	return refs
}

// dedupe returns the distinct owners in deterministic order. Several providers
// share a namespace, and a module owner may also publish providers.
func dedupe(owners []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(owners))
	for _, o := range owners {
		if o == "" || seen[o] {
			continue
		}
		seen[o] = true
		out = append(out, o)
	}
	sort.Strings(out)
	return out
}
