// Package version orders the "semver-ish" version strings the registry meets:
// provider versions from the OpenTofu registry and module release tags from
// git. Both callers need the same rule, so it lives here rather than being
// implemented twice.
package version

import (
	"sort"
	"strings"
)

// Normalize strips a leading "v" so a git tag and a registry version agree:
// the tag "v0.1.0" and the provider version "0.1.0" are the same version.
func Normalize(v string) string { return strings.TrimPrefix(v, "v") }

// IsPrerelease reports whether a version string is a prerelease.
func IsPrerelease(v string) bool {
	if strings.IndexByte(v, '-') >= 0 {
		return true
	}
	lower := strings.ToLower(v)
	for _, m := range []string{"rc", "beta", "alpha", "pre"} {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

// key extracts the leading numeric triple from a version string.
func key(v string) [3]int {
	var out [3]int
	parts := strings.FieldsFunc(v, func(r rune) bool { return r < '0' || r > '9' })
	for i := 0; i < 3 && i < len(parts); i++ {
		n := 0
		for _, ch := range parts[i] {
			n = n*10 + int(ch-'0')
		}
		out[i] = n
	}
	return out
}

// Less orders two versions: a prerelease always sorts below a stable release,
// then by numeric triple, then lexically so the order is total and stable.
func Less(a, b string) bool {
	if pa, pb := IsPrerelease(a), IsPrerelease(b); pa != pb {
		return pa
	}
	ka, kb := key(a), key(b)
	for i := 0; i < 3; i++ {
		if ka[i] != kb[i] {
			return ka[i] < kb[i]
		}
	}
	return a < b
}

// LatestStableIndex returns the index of the best version: the highest stable
// release, or the highest prerelease when no stable one exists. It reports
// false for an empty list.
func LatestStableIndex(versions []string) (int, bool) {
	if len(versions) == 0 {
		return 0, false
	}
	idx := make([]int, len(versions))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(i, j int) bool {
		return Less(versions[idx[i]], versions[idx[j]])
	})
	return idx[len(idx)-1], true
}

// LatestStable returns the best version string. It reports false for an empty
// list.
func LatestStable(versions []string) (string, bool) {
	i, ok := LatestStableIndex(versions)
	if !ok {
		return "", false
	}
	return versions[i], true
}
