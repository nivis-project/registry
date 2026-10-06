package module

import (
	"context"
	"sort"
	"strings"

	"github.com/nivis-project/registry/tools/version"
)

// Tag is one release tag of a module repository.
type Tag struct {
	Name string // as published, e.g. "v0.1.0"
	Rev  string // the COMMIT the tag points at
}

// Version is the tag without its leading "v", matching the provider
// convention where hcloudimage's v0.1.0 tag is version 0.1.0.
func (t Tag) Version() string { return version.Normalize(t.Name) }

// parseLsRemote turns `git ls-remote --tags` output into tags.
//
// An ANNOTATED tag yields two lines: "<tagobject> refs/tags/v1" and
// "<commit> refs/tags/v1^{}". The peeled line is the one that names the
// commit, so it always wins; using the unpeeled sha would reference the tag
// object and resolve to the wrong tree.
func parseLsRemote(out string) []Tag {
	rev := map[string]string{}
	peeled := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || !strings.HasPrefix(fields[1], "refs/tags/") {
			continue
		}
		name := strings.TrimPrefix(fields[1], "refs/tags/")
		isPeeled := strings.HasSuffix(name, "^{}")
		name = strings.TrimSuffix(name, "^{}")
		if name == "" {
			continue
		}
		if isPeeled || !peeled[name] {
			rev[name] = fields[0]
		}
		if isPeeled {
			peeled[name] = true
		}
	}
	names := make([]string, 0, len(rev))
	for n := range rev {
		names = append(names, n)
	}
	sort.Strings(names)
	tags := make([]Tag, 0, len(names))
	for _, n := range names {
		tags = append(tags, Tag{Name: n, Rev: rev[n]})
	}
	return tags
}

// LatestTag lists a repository's tags and returns the newest stable release.
// It reports false when the repository has published none.
func (c *Client) LatestTag(ctx context.Context, p Pin) (Tag, bool, error) {
	out, err := c.run(ctx, c.gitBin, "ls-remote", "--tags", p.GitURL())
	if err != nil {
		return Tag{}, false, err
	}
	tags := parseLsRemote(string(out))
	if len(tags) == 0 {
		return Tag{}, false, nil
	}
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	idx, ok := version.LatestStableIndex(names)
	if !ok {
		return Tag{}, false, nil
	}
	return tags[idx], true, nil
}
