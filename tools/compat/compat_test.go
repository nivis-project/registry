package compat

import (
	"reflect"
	"testing"
)

// A provider not in the allowlist must never be "verified".
func TestUnverifiedIsByDesign(t *testing.T) {
	if got := TierFor("hashicorp/random", true); got != TierByDesign {
		t.Errorf("TierFor(unverified) = %q, want %q", got, TierByDesign)
	}
}

// Architectures are an intersection: linux/arm64 must drop out when not published.
func TestIntersectArches(t *testing.T) {
	published := []string{"linux/amd64", "darwin/amd64", "darwin/arm64", "windows/amd64"}
	got := IntersectArches(published)
	want := []string{"linux/amd64", "darwin/amd64", "darwin/arm64"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("IntersectArches = %v, want %v (linux/arm64 must drop, windows is not Nivis-supported)", got, want)
	}
}
