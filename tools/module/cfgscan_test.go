package module

import "testing"

func paths(keys []CfgKey) map[string]bool {
	m := map[string]bool{}
	for _, k := range keys {
		m[k.Path] = k.Required
	}
	return m
}

// TestStripCommentsIgnoresProse is the bug this function exists for:
// nivis-aws-amplify-site's header comment names a configuration key in
// English. A scan that reads comments invents keys a module never uses.
func TestStripCommentsIgnoresProse(t *testing.T) {
	src := `
# The GitHub token is read from SSM at deploy time (cfg.inventedByAComment);
/* and cfg.inventedByABlockComment too */
{ cfg }: { name = cfg.realKey; }
`
	got := paths(ScanCfg(src))
	if _, ok := got["inventedByAComment"]; ok {
		t.Error("a key mentioned only in a line comment must not be reported")
	}
	if _, ok := got["inventedByABlockComment"]; ok {
		t.Error("a key mentioned only in a block comment must not be reported")
	}
	if _, ok := got["realKey"]; !ok {
		t.Error("a key actually read must be reported")
	}
}

// TestStripCommentsKeepsStrings: a "#" inside a string literal is data, not a
// comment, and stripping it would corrupt the source being scanned.
func TestStripCommentsKeepsStrings(t *testing.T) {
	src := `{ cfg }: { a = "not # a comment ${cfg.inString}"; b = ''also /* not */ ${cfg.inIndented}''; }`
	got := paths(ScanCfg(src))
	for _, want := range []string{"inString", "inIndented"} {
		if _, ok := got[want]; !ok {
			t.Errorf("key %q inside a string literal must still be seen", want)
		}
	}
}

func TestScanCfgMarksOptionality(t *testing.T) {
	src := `{ cfg }: {
    required  = cfg.appName;
    fallback  = cfg.buildSpec or null;
    listish   = (cfg.preRules or [ ]) ++ [ ];
  }`
	got := paths(src2keys(src))
	if req, ok := got["appName"]; !ok || !req {
		t.Error("a key read unconditionally must be required")
	}
	for _, opt := range []string{"buildSpec", "preRules"} {
		if req, ok := got[opt]; !ok || req {
			t.Errorf("%q is read with a fallback and must be optional", opt)
		}
	}
}

// TestScanCfgNestedPaths: a nested path is reported in full, and inherits its
// ancestor's optionality because it is only ever read when the ancestor was
// supplied.
func TestScanCfgNestedPaths(t *testing.T) {
	src := `{ cfg }: {
    a = cfg.ses.from;
    b = if cfg.formProxy or null == null then null else cfg.formProxy.pathPrefix;
  }`
	got := paths(src2keys(src))
	if req, ok := got["ses.from"]; !ok || !req {
		t.Error("ses.from must be reported in full and required")
	}
	if _, ok := got["ses"]; ok {
		t.Error("only the full path is reported, not a bare ancestor that is never read alone")
	}
	if req, ok := got["formProxy.pathPrefix"]; !ok || req {
		t.Error("a key under an optional ancestor must itself be optional")
	}
}

func TestScanCfgDeterministicOrder(t *testing.T) {
	src := `{ cfg }: { z = cfg.zebra; a = cfg.apple; m = cfg.mango; }`
	first := src2keys(src)
	for i := 0; i < 5; i++ {
		again := src2keys(src)
		for j := range first {
			if first[j].Path != again[j].Path {
				t.Fatalf("scan order is not stable at %d: %q vs %q", j, first[j].Path, again[j].Path)
			}
		}
	}
	if first[0].Path != "apple" {
		t.Errorf("keys should be sorted, got %q first", first[0].Path)
	}
}

func src2keys(src string) []CfgKey { return ScanCfg(src) }
