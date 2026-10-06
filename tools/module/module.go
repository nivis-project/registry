package module

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	_ "embed"
)

//go:embed probe.nix
var probeExpr string

// DefaultNivisRef is the nivis the probe applies modules with. A module does
// not declare nivis as an input; the consumer supplies it, so the registry
// pins which one it documented against. This matches the rev in the repo's
// flake.lock.
const DefaultNivisRef = "git+https://github.com/wearetechnative/nivis?rev=2ef41acf5d40a4c8a914526b9808a7d9af8b0450"

// DefaultTimeout bounds one module's evaluation. Pure Nix evaluation cannot
// perform arbitrary IO, but it can fail to terminate, and unlike the provider
// path there is no checksum gate before we run third-party code.
const DefaultTimeout = 5 * time.Minute

// Coord is one resource or data source a module declares.
type Coord struct {
	Provider string `json:"provider"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	// Docs is the provider version documenting Type ("<ns>/<name>/<version>"),
	// empty when the registry does not catalogue it.
	Docs string `json:"docs,omitempty"`
}

// Record states what was derived and what was not, so a consumer never has to
// guess which half of a module page is machine-checked. It is the module
// counterpart of the provider compat record.
type Record struct {
	StructureExtractable bool   `json:"structure_extractable"`
	CfgSource            string `json:"cfg_source"` // "scanned" | "declared"
	ProvidersResolved    int    `json:"providers_resolved"`
	ProvidersTotal       int    `json:"providers_total"`
	FailureReason        string `json:"failure_reason,omitempty"`
}

// CfgSourceScanned marks a configuration surface recovered from source text
// rather than from a declaration.
const CfgSourceScanned = "scanned"

// Module is the per-module extraction outcome, written as module.json.
type Module struct {
	Owner       string   `json:"owner"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Tag         string   `json:"tag"`
	Rev         string   `json:"rev"`
	Reason      string   `json:"reason"`
	Note        string   `json:"note,omitempty"`
	Description string   `json:"description,omitempty"`
	Readme      string   `json:"readme,omitempty"`
	Resources   []Coord  `json:"resources"`
	DataSources []Coord  `json:"data_sources"`
	Outputs     []string `json:"outputs"`
	Composition []string `json:"composition,omitempty"`
	Cfg         []CfgKey `json:"cfg"`
	Record      Record   `json:"record"`
	SkipReason  string   `json:"skip_reason,omitempty"`
}

// Repo is the canonical "<owner>/<name>" identity.
func (m Module) Repo() string { return m.Owner + "/" + m.Name }

// Skipped reports whether this module was skipped rather than extracted.
func (m Module) Skipped() bool { return m.SkipReason != "" }

// probeOut mirrors probe.nix's JSON.
type probeOut struct {
	Src     string `json:"src"`
	Modules map[string]struct {
		Resources   []Coord  `json:"resources"`
		DataSources []Coord  `json:"dataSources"`
		Outputs     []string `json:"outputs"`
		Composition []string `json:"composition"`
	} `json:"modules"`
}

// flakeMeta mirrors the fields of `nix flake metadata --json` we use. The
// description is flake METADATA, not an output, so it cannot come from the
// probe.
type flakeMeta struct {
	Description string `json:"description"`
}

// runner executes a command and returns its stdout. Tests replace it so the
// argument building and output parsing are exercised without spawning Nix.
type runner func(ctx context.Context, name string, args ...string) ([]byte, error)

// Client extracts modules.
type Client struct {
	nixBin   string
	gitBin   string
	nivisRef string
	timeout  time.Duration
	index    ProviderIndex
	run      runner
}

// Option configures a Client.
type Option func(*Client)

// WithNivisRef overrides the nivis flake reference modules are applied with.
func WithNivisRef(ref string) Option { return func(c *Client) { c.nivisRef = ref } }

// WithProviderIndex supplies the resource-type index used for provider links.
func WithProviderIndex(i ProviderIndex) Option { return func(c *Client) { c.index = i } }

// WithTimeout overrides the per-module evaluation bound.
func WithTimeout(d time.Duration) Option { return func(c *Client) { c.timeout = d } }

// WithRunner replaces the command runner (tests inject canned output).
func WithRunner(r runner) Option { return func(c *Client) { c.run = r } }

// New returns a module extraction Client.
func New(opts ...Option) *Client {
	c := &Client{
		nixBin:   "nix",
		gitBin:   "git",
		nivisRef: DefaultNivisRef,
		timeout:  DefaultTimeout,
		index:    ProviderIndex{},
		run:      execRun,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func execRun(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if ok := asExitError(err, &ee); ok && len(ee.Stderr) > 0 {
			return nil, fmt.Errorf("%s: %w: %s", name, err, trimStderr(ee.Stderr))
		}
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return out, nil
}

// FlakeRef is the locked reference the probe evaluates: plain git over https
// with an explicit rev, never the GitHub API and never a moving branch.
func FlakeRef(gitURL, rev string) string { return "git+" + gitURL + "?rev=" + rev }

// probe runs the structure pass.
func (c *Client) probe(ctx context.Context, moduleRef string) (probeOut, error) {
	dir, err := os.MkdirTemp("", "nivis-module-probe")
	if err != nil {
		return probeOut{}, err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "probe.nix")
	if err := os.WriteFile(path, []byte(probeExpr), 0o644); err != nil {
		return probeOut{}, err
	}

	out, err := c.run(ctx, "nix-instantiate", "--eval", "--strict", "--json", path,
		"--argstr", "nivisRef", c.nivisRef,
		"--argstr", "moduleRef", moduleRef)
	if err != nil {
		return probeOut{}, err
	}
	var po probeOut
	if err := json.Unmarshal(out, &po); err != nil {
		return probeOut{}, fmt.Errorf("decode probe output: %w", err)
	}
	return po, nil
}

// describe reads the flake description.
func (c *Client) describe(ctx context.Context, moduleRef string) string {
	out, err := c.run(ctx, c.nixBin, "flake", "metadata", "--json", moduleRef)
	if err != nil {
		return ""
	}
	var fm flakeMeta
	if err := json.Unmarshal(out, &fm); err != nil {
		return ""
	}
	return fm.Description
}

// link resolves coordinates against the provider index, returning how many of
// them the registry documents.
func (c *Client) link(coords []Coord) ([]Coord, int) {
	out := make([]Coord, 0, len(coords))
	resolved := 0
	for _, co := range coords {
		if addr, ok := c.index.Resolve(co.Type); ok {
			co.Docs = addr
			resolved++
		}
		out = append(out, co)
	}
	return out, resolved
}

// Extract runs all three passes for one pinned module.
//
// A module that cannot be evaluated is still returned, with its structure
// omitted and the reason recorded, rather than falling back to a render that
// would look as authoritative as a derived one.
func (c *Client) Extract(ctx context.Context, p Pin) (Module, error) {
	m := Module{Owner: p.Owner, Name: p.Name, Reason: p.Reason, Note: p.Note}

	tag, ok, err := c.LatestTag(ctx, p)
	if err != nil {
		return m, fmt.Errorf("list tags for %s: %w", p.Repo(), err)
	}
	if !ok {
		m.SkipReason = "no release tag published"
		return m, nil
	}
	m.Tag, m.Rev, m.Version = tag.Name, tag.Rev, tag.Version()

	ref := FlakeRef(p.GitURL(), tag.Rev)
	m.Description = c.describe(ctx, ref)

	po, probeErr := c.probe(ctx, ref)
	if probeErr != nil {
		m.Record = Record{
			StructureExtractable: false,
			CfgSource:            CfgSourceScanned,
			FailureReason:        probeErr.Error(),
		}
		m.Resources, m.DataSources, m.Outputs = []Coord{}, []Coord{}, []string{}
		m.Cfg = []CfgKey{}
		return m, nil
	}

	entry, ok := po.Modules["default"]
	if !ok {
		m.SkipReason = "flake exposes no nivisModules.default"
		return m, nil
	}

	resources, rResolved := c.link(entry.Resources)
	dataSources, dResolved := c.link(entry.DataSources)
	m.Resources, m.DataSources = resources, dataSources
	m.Outputs = nonNil(entry.Outputs)
	m.Composition = entry.Composition

	if src, err := os.ReadFile(filepath.Join(po.Src, "module.nix")); err == nil {
		m.Cfg = ScanCfg(string(src))
	} else {
		m.Cfg = []CfgKey{}
	}
	if readme, err := os.ReadFile(filepath.Join(po.Src, "README.md")); err == nil {
		m.Readme = string(readme)
	}

	m.Record = Record{
		StructureExtractable: true,
		CfgSource:            CfgSourceScanned,
		ProvidersResolved:    rResolved + dResolved,
		ProvidersTotal:       len(resources) + len(dataSources),
	}
	return m, nil
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// OutPath is where a module's extraction output lands, mirroring the provider
// layout <root>/<owner>/<name>/<version>/.
func OutPath(root string, m Module) string {
	return filepath.Join(root, m.Owner, m.Name, m.Version)
}

// Write persists a module's extraction output as module.json.
func Write(root string, m Module) (string, error) {
	dir := OutPath(root, m)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "module.json")
	return path, os.WriteFile(path, append(b, '\n'), 0o644)
}

// Logger receives one line per module in a batch.
type Logger func(format string, args ...interface{})

// Batch extracts every pinned module resiliently: one Result per input, a
// single failure recorded as a skip, never aborting or hanging the run.
func (c *Client) Batch(ctx context.Context, pins []Pin, log Logger) []Module {
	if log == nil {
		log = func(string, ...interface{}) {}
	}
	out := make([]Module, 0, len(pins))
	for _, p := range pins {
		modCtx, cancel := context.WithTimeout(ctx, c.timeout)
		m, err := c.Extract(modCtx, p)
		cancel()
		switch {
		case err != nil:
			m.SkipReason = err.Error()
			log("skip %s: %v", p.Repo(), err)
		case m.Skipped():
			log("skip %s: %s", p.Repo(), m.SkipReason)
		case !m.Record.StructureExtractable:
			log("partial %s @ %s: structure not derivable: %s", p.Repo(), m.Version, m.Record.FailureReason)
		default:
			log("ok   %s @ %s: %d resource(s), %d data source(s), %d/%d linked",
				p.Repo(), m.Version, len(m.Resources), len(m.DataSources),
				m.Record.ProvidersResolved, m.Record.ProvidersTotal)
		}
		out = append(out, m)
	}
	return out
}
