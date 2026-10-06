package module

import (
	"os"
	"path/filepath"
	"testing"
)

func repoPins(t *testing.T) []Pin {
	t.Helper()
	pins, err := LoadPins(filepath.Join("..", "..", DefaultPinsPath))
	if err != nil {
		t.Fatalf("LoadPins(repo): %v", err)
	}
	return pins
}

// TestRepoPinsValid guards the committed module list.
func TestRepoPinsValid(t *testing.T) {
	pins := repoPins(t)
	if len(pins) == 0 {
		t.Fatal("module-pins.json must not be empty")
	}
	want := map[string]bool{
		"wearetechnative/nivis-aws-amplify-site": false,
		"wearetechnative/nivis-aws-form-action":  false,
	}
	for _, p := range pins {
		if _, ok := want[p.Repo()]; ok {
			want[p.Repo()] = true
		}
		if p.Reason == "" {
			t.Errorf("%s has no reason", p.Repo())
		}
		if got := p.GitURL(); got != "https://github.com/"+p.Repo() {
			t.Errorf("GitURL = %q", got)
		}
	}
	for repo, found := range want {
		if !found {
			t.Errorf("module-pins.json missing %q", repo)
		}
	}
}

func TestLoadPinsRejectsBadEntries(t *testing.T) {
	for name, body := range map[string]string{
		"no owner":       `{"modules":[{"owner":"","name":"m","reason":"curated"}]}`,
		"no name":        `{"modules":[{"owner":"o","name":"","reason":"curated"}]}`,
		"no reason":      `{"modules":[{"owner":"o","name":"m","reason":""}]}`,
		"slash in name":  `{"modules":[{"owner":"o","name":"a/b","reason":"curated"}]}`,
		"duplicate":      `{"modules":[{"owner":"o","name":"m","reason":"a"},{"owner":"o","name":"m","reason":"b"}]}`,
		"malformed json": `{"modules":[`,
	} {
		path := filepath.Join(t.TempDir(), "pins.json")
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadPins(path); err == nil {
			t.Errorf("%s: LoadPins should have failed", name)
		}
	}
	if _, err := LoadPins(filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Error("LoadPins on a missing file should fail")
	}
}
