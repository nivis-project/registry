package module

import (
	"context"
	"errors"
	"testing"
)

func stubRun(out string, err error) Option {
	return WithRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return []byte(out), err
	})
}

// TestParseLsRemotePeelsAnnotatedTags is the detail that silently breaks
// everything: an annotated tag's plain ref names the TAG OBJECT, and only the
// "^{}" ref names the commit. Using the former fetches the wrong tree.
func TestParseLsRemotePeelsAnnotatedTags(t *testing.T) {
	out := "860c92a\trefs/tags/v0.1.0\n1b76f03\trefs/tags/v0.1.0^{}\nabc1234\trefs/tags/v0.0.9\n"
	tags := parseLsRemote(out)
	if len(tags) != 2 {
		t.Fatalf("got %d tags, want 2 (the peeled ref is not a separate tag)", len(tags))
	}
	byName := map[string]string{}
	for _, tg := range tags {
		byName[tg.Name] = tg.Rev
	}
	if byName["v0.1.0"] != "1b76f03" {
		t.Errorf("v0.1.0 rev = %q, want the peeled commit 1b76f03", byName["v0.1.0"])
	}
	if byName["v0.0.9"] != "abc1234" {
		t.Errorf("a lightweight tag keeps its own sha, got %q", byName["v0.0.9"])
	}
}

func TestParseLsRemoteIgnoresNonTags(t *testing.T) {
	out := "aaa\trefs/heads/main\nbbb\tHEAD\nccc\trefs/tags/v1.0.0\nnonsense\n"
	tags := parseLsRemote(out)
	if len(tags) != 1 || tags[0].Name != "v1.0.0" {
		t.Errorf("got %v, want only v1.0.0", tags)
	}
}

func TestTagVersionStripsThePrefix(t *testing.T) {
	if got := (Tag{Name: "v0.1.0"}).Version(); got != "0.1.0" {
		t.Errorf("Version() = %q, want 0.1.0 to match the provider convention", got)
	}
	if got := (Tag{Name: "1.2.3"}).Version(); got != "1.2.3" {
		t.Errorf("Version() = %q, want 1.2.3 unchanged", got)
	}
}

func TestLatestTagPrefersStable(t *testing.T) {
	out := "a\trefs/tags/v0.1.0\nb\trefs/tags/v0.2.0-rc1\nc\trefs/tags/v0.0.9\n"
	c := New(stubRun(out, nil))
	tag, ok, err := c.LatestTag(context.Background(), Pin{Owner: "o", Name: "m"})
	if err != nil || !ok {
		t.Fatalf("LatestTag: ok=%v err=%v", ok, err)
	}
	if tag.Name != "v0.1.0" {
		t.Errorf("picked %q, want the newest STABLE tag v0.1.0", tag.Name)
	}
}

func TestLatestTagFallsBackToPrerelease(t *testing.T) {
	c := New(stubRun("a\trefs/tags/v0.2.0-rc1\n", nil))
	tag, ok, _ := c.LatestTag(context.Background(), Pin{Owner: "o", Name: "m"})
	if !ok || tag.Name != "v0.2.0-rc1" {
		t.Errorf("got %q ok=%v, want the prerelease when nothing stable exists", tag.Name, ok)
	}
}

func TestLatestTagReportsNoTags(t *testing.T) {
	c := New(stubRun("", nil))
	if _, ok, err := c.LatestTag(context.Background(), Pin{Owner: "o", Name: "m"}); ok || err != nil {
		t.Errorf("ok=%v err=%v, want no tags and no error", ok, err)
	}
}

func TestLatestTagSurfacesGitFailure(t *testing.T) {
	c := New(stubRun("", errors.New("repository not found")))
	if _, _, err := c.LatestTag(context.Background(), Pin{Owner: "o", Name: "m"}); err == nil {
		t.Error("a git failure must surface, not look like an untagged repository")
	}
}

func TestFlakeRefIsLockedAndAvoidsTheAPI(t *testing.T) {
	got := FlakeRef("https://github.com/o/m", "deadbeef")
	want := "git+https://github.com/o/m?rev=deadbeef"
	if got != want {
		t.Errorf("FlakeRef = %q, want %q", got, want)
	}
}
