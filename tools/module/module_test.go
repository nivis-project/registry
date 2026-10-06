package module

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeSrc writes a module source tree the probe would have fetched, so the
// configuration scan and README pickup run against real files.
func fakeSrc(t *testing.T, moduleNix, readme string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "module.nix"), []byte(moduleNix), 0o644); err != nil {
		t.Fatal(err)
	}
	if readme != "" {
		if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// scriptedRunner answers each command by the binary and subcommand asked for,
// so Extract's argument building and output parsing are exercised without
// spawning Nix.
type scriptedRunner struct {
	tags     string
	tagsErr  error
	probe    string
	probeErr error
	metadata string
	calls    []string
}

func (s *scriptedRunner) opt() Option {
	return WithRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		s.calls = append(s.calls, name+" "+strings.Join(args, " "))
		switch {
		case strings.HasSuffix(name, "git"):
			return []byte(s.tags), s.tagsErr
		case name == "nix-instantiate":
			return []byte(s.probe), s.probeErr
		default:
			return []byte(s.metadata), nil
		}
	})
}

func probeJSON(t *testing.T, src string, resources, dataSources []Coord, outputs, composition []string) string {
	t.Helper()
	type entry struct {
		Resources   []Coord  `json:"resources"`
		DataSources []Coord  `json:"dataSources"`
		Outputs     []string `json:"outputs"`
		Composition []string `json:"composition"`
	}
	b, err := json.Marshal(map[string]interface{}{
		"src": src,
		"modules": map[string]entry{
			"default": {resources, dataSources, outputs, composition},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestExtractDerivesStructureAndLinks(t *testing.T) {
	src := fakeSrc(t,
		`# a comment mentioning cfg.notARealKey
{ cfg }: { a = cfg.appName; b = cfg.buildSpec or null; }`,
		"# amplify\nprose here")

	r := &scriptedRunner{
		tags: "aaa\trefs/tags/v0.1.0\nbbb\trefs/tags/v0.1.0^{}\n",
		probe: probeJSON(t, src,
			[]Coord{{Provider: "aws", Type: "aws_iam_role", Name: "r"}},
			[]Coord{{Provider: "aws", Type: "aws_unknown_thing", Name: "d"}},
			[]string{"app_id"}, nil),
		metadata: `{"description":"a nivis module"}`,
	}
	c := New(r.opt(), WithProviderIndex(ProviderIndex{"aws_iam_role": "hashicorp/aws/6.67.0"}))

	m, err := c.Extract(context.Background(), Pin{Owner: "o", Name: "m", Reason: "curated"})
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if m.Skipped() {
		t.Fatalf("unexpected skip: %s", m.SkipReason)
	}
	if m.Version != "0.1.0" || m.Tag != "v0.1.0" || m.Rev != "bbb" {
		t.Errorf("version/tag/rev = %q/%q/%q, want 0.1.0/v0.1.0/bbb (the PEELED rev)", m.Version, m.Tag, m.Rev)
	}
	if m.Description != "a nivis module" {
		t.Errorf("description = %q", m.Description)
	}
	if m.Readme == "" {
		t.Error("README must be carried, separately from the derived structure")
	}
	if got := m.Resources[0].Docs; got != "hashicorp/aws/6.67.0" {
		t.Errorf("resource link = %q, want the provider version", got)
	}
	if got := m.DataSources[0].Docs; got != "" {
		t.Errorf("an uncatalogued type must stay unresolved, got %q", got)
	}
	if m.Record.ProvidersResolved != 1 || m.Record.ProvidersTotal != 2 {
		t.Errorf("record resolved %d/%d, want 1/2", m.Record.ProvidersResolved, m.Record.ProvidersTotal)
	}
	if !m.Record.StructureExtractable || m.Record.CfgSource != CfgSourceScanned {
		t.Errorf("record = %+v, want derived structure and a scanned cfg source", m.Record)
	}
	got := map[string]bool{}
	for _, k := range m.Cfg {
		got[k.Path] = k.Required
	}
	if req, ok := got["appName"]; !ok || !req {
		t.Error("appName must be required")
	}
	if req, ok := got["buildSpec"]; !ok || req {
		t.Error("buildSpec must be optional")
	}
	if _, ok := got["notARealKey"]; ok {
		t.Error("a key named only in a comment must not reach the contract")
	}

	// The flake reference must be locked to the peeled rev, never a branch.
	var sawRef bool
	for _, call := range r.calls {
		if strings.Contains(call, "git+https://github.com/o/m?rev=bbb") {
			sawRef = true
		}
	}
	if !sawRef {
		t.Errorf("probe was not given a rev-locked flake reference; calls: %v", r.calls)
	}
}

// TestExtractRecordsAProbeFailure is the Pass A fallback: a module that
// branches its resource list on cfg forces the poisoned value and cannot be
// evaluated. It must still be listed, with the structure omitted and the
// reason stated, never with a guess.
func TestExtractRecordsAProbeFailure(t *testing.T) {
	r := &scriptedRunner{
		tags:     "aaa\trefs/tags/v0.1.0^{}\n",
		probeErr: errors.New("error: module read cfg while probing its structure"),
	}
	c := New(r.opt())

	m, err := c.Extract(context.Background(), Pin{Owner: "o", Name: "m"})
	if err != nil {
		t.Fatalf("a probe failure must not be a hard error: %v", err)
	}
	if m.Record.StructureExtractable {
		t.Error("structure must not be reported as derived")
	}
	if m.Record.FailureReason == "" {
		t.Error("the failure reason must be recorded")
	}
	if len(m.Resources) != 0 || len(m.Outputs) != 0 {
		t.Error("no structure may be published for a module that did not evaluate")
	}
	if m.Version != "0.1.0" {
		t.Errorf("the version is still known, got %q", m.Version)
	}
}

func TestExtractSkipsUntaggedModule(t *testing.T) {
	c := New((&scriptedRunner{tags: ""}).opt())
	m, err := c.Extract(context.Background(), Pin{Owner: "o", Name: "m"})
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if !m.Skipped() || !strings.Contains(m.SkipReason, "no release tag") {
		t.Errorf("an untagged module must be skipped with that reason, got %q", m.SkipReason)
	}
}

func TestExtractSkipsFlakeWithoutDefaultModule(t *testing.T) {
	r := &scriptedRunner{
		tags:  "aaa\trefs/tags/v0.1.0^{}\n",
		probe: `{"src":"/nonexistent","modules":{"other":{"resources":[],"dataSources":[],"outputs":[],"composition":[]}}}`,
	}
	m, err := r.extract(t)
	if err != nil {
		t.Fatal(err)
	}
	if !m.Skipped() || !strings.Contains(m.SkipReason, "nivisModules.default") {
		t.Errorf("want a skip naming the missing default output, got %q", m.SkipReason)
	}
}

// TestExtractHandlesMissingDataSourcesKey: nivis-aws-form-action returns no
// dataSources key at all. That is a valid module, not a failure.
func TestExtractHandlesMissingDataSourcesKey(t *testing.T) {
	src := fakeSrc(t, `{ cfg }: { a = cfg.x; }`, "")
	r := &scriptedRunner{
		tags:  "aaa\trefs/tags/v0.1.0^{}\n",
		probe: `{"src":"` + src + `","modules":{"default":{"resources":[],"outputs":["o"],"composition":["apiEndpointRef"]}}}`,
	}
	m, err := r.extract(t)
	if err != nil {
		t.Fatal(err)
	}
	if m.Skipped() || !m.Record.StructureExtractable {
		t.Fatalf("a module without a dataSources key must still extract: %q", m.SkipReason)
	}
	if m.DataSources == nil || len(m.DataSources) != 0 {
		t.Errorf("data sources = %v, want an empty set", m.DataSources)
	}
	if len(m.Composition) != 1 || m.Composition[0] != "apiEndpointRef" {
		t.Errorf("composition keys must be preserved, got %v", m.Composition)
	}
}

func (r *scriptedRunner) extract(t *testing.T) (Module, error) {
	t.Helper()
	return New(r.opt()).Extract(context.Background(), Pin{Owner: "o", Name: "m"})
}

func TestBatchIsResilient(t *testing.T) {
	src := fakeSrc(t, `{ cfg }: { a = cfg.x; }`, "")
	good := probeJSON(t, src, []Coord{{Provider: "aws", Type: "t", Name: "n"}}, nil, []string{"o"}, nil)

	calls := 0
	c := New(WithRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if strings.HasSuffix(name, "git") {
			calls++
			if calls == 1 {
				return nil, errors.New("repository not found")
			}
			return []byte("aaa\trefs/tags/v0.1.0^{}\n"), nil
		}
		return []byte(good), nil
	}))

	var logs []string
	pins := []Pin{{Owner: "o", Name: "broken"}, {Owner: "o", Name: "fine"}}
	got := c.Batch(context.Background(), pins, func(f string, a ...interface{}) {
		logs = append(logs, f)
	})

	if len(got) != len(pins) {
		t.Fatalf("Batch returned %d results, want one per input", len(got))
	}
	if !got[0].Skipped() {
		t.Error("the failing module must be skipped with a reason")
	}
	if got[1].Skipped() {
		t.Errorf("the healthy module must still extract: %s", got[1].SkipReason)
	}
	if len(logs) != 2 {
		t.Errorf("want one log line per module, got %d", len(logs))
	}
}

// TestBatchBoundsEvaluation: a module that never terminates is abandoned, not
// allowed to hang the run.
func TestBatchBoundsEvaluation(t *testing.T) {
	c := New(
		WithTimeout(20*time.Millisecond),
		WithRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
			if strings.HasSuffix(name, "git") {
				return []byte("aaa\trefs/tags/v0.1.0^{}\n"), nil
			}
			<-ctx.Done()
			return nil, ctx.Err()
		}),
	)
	done := make(chan []Module, 1)
	go func() { done <- c.Batch(context.Background(), []Pin{{Owner: "o", Name: "slow"}}, nil) }()

	select {
	case got := <-done:
		if len(got) != 1 {
			t.Fatalf("got %d results", len(got))
		}
		if got[0].Record.StructureExtractable {
			t.Error("a timed-out evaluation must not report a derived structure")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Batch hung on a non-terminating module")
	}
}

func TestWriteRoundTrips(t *testing.T) {
	root := t.TempDir()
	m := Module{Owner: "o", Name: "m", Version: "0.1.0", Outputs: []string{"a"}}
	path, err := Write(root, m)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if want := filepath.Join(root, "o", "m", "0.1.0", "module.json"); path != want {
		t.Errorf("path = %q, want %q (mirroring the provider layout)", path, want)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var back Module
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("module.json does not round-trip: %v", err)
	}
	if back.Repo() != "o/m" || back.Version != "0.1.0" {
		t.Errorf("round-tripped to %+v", back)
	}
}

func TestProbeExprIsEmbedded(t *testing.T) {
	for _, want := range []string{"nivisModules", "cfg = throw", "getFlake"} {
		if !strings.Contains(probeExpr, want) {
			t.Errorf("embedded probe.nix is missing %q", want)
		}
	}
}
