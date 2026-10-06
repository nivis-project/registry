package module

import (
	"os"
	"path/filepath"
	"testing"
)

// fakeExtractTree builds an extraction-output tree shaped like the real one:
// <namespace>/<name>/<version>/<identity>/<resource_type>.nix
func fakeExtractTree(t *testing.T, entries map[string][]string) string {
	t.Helper()
	root := t.TempDir()
	for addr, types := range entries {
		dir := filepath.Join(root, filepath.FromSlash(addr), "ident")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, ty := range types {
			if err := os.WriteFile(filepath.Join(dir, ty+".nix"), []byte("{}"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(dir, "README.txt"), []byte("ignored"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestBuildProviderIndexMapsTypes(t *testing.T) {
	root := fakeExtractTree(t, map[string][]string{
		"hashicorp/aws/6.67.0":            {"aws_amplify_app", "aws_iam_role"},
		"nivis-project/hcloudimage/0.1.0": {"hcloudimage_image"},
	})
	idx, err := BuildProviderIndex(root)
	if err != nil {
		t.Fatalf("BuildProviderIndex: %v", err)
	}
	if len(idx) != 3 {
		t.Errorf("index has %d entries, want 3 (non-.nix files ignored)", len(idx))
	}
	for ty, want := range map[string]string{
		"aws_amplify_app":   "hashicorp/aws/6.67.0",
		"hcloudimage_image": "nivis-project/hcloudimage/0.1.0",
	} {
		got, ok := idx.Resolve(ty)
		if !ok || got != want {
			t.Errorf("Resolve(%q) = %q,%v, want %q", ty, got, ok, want)
		}
	}
}

// TestResolveReportsUnknownTypes: an uncatalogued type is a provable fact, not
// a reason to guess a provider from its name prefix.
func TestResolveReportsUnknownTypes(t *testing.T) {
	idx, err := BuildProviderIndex(fakeExtractTree(t, map[string][]string{
		"hashicorp/aws/6.67.0": {"aws_iam_role"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := idx.Resolve("aws_not_catalogued"); ok {
		t.Errorf("Resolve returned %q for an uncatalogued type; it must not infer from the aws_ prefix", got)
	}
}

func TestBuildProviderIndexIsDeterministic(t *testing.T) {
	root := fakeExtractTree(t, map[string][]string{
		"bbb/two/1.0.0": {"shared_type"},
		"aaa/one/1.0.0": {"shared_type"},
	})
	for i := 0; i < 5; i++ {
		idx, err := BuildProviderIndex(root)
		if err != nil {
			t.Fatal(err)
		}
		if got, _ := idx.Resolve("shared_type"); got != "aaa/one/1.0.0" {
			t.Fatalf("run %d resolved to %q, want the lexically first address", i, got)
		}
	}
}

func TestBuildProviderIndexOnMissingRoot(t *testing.T) {
	idx, err := BuildProviderIndex(filepath.Join(t.TempDir(), "absent"))
	if err != nil {
		t.Fatalf("a missing extraction root should be empty, not an error: %v", err)
	}
	if len(idx) != 0 {
		t.Errorf("got %d entries, want none", len(idx))
	}
}
