package version

import "testing"

func TestNormalizeStripsTheTagPrefix(t *testing.T) {
	for in, want := range map[string]string{
		"v0.1.0": "0.1.0",
		"0.1.0":  "0.1.0",
		"v1":     "1",
		"":       "",
	} {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsPrerelease(t *testing.T) {
	for _, v := range []string{"1.0.0-rc1", "2.0.0-beta", "0.1.0-alpha.2", "1.0.0-pre", "3.0.0rc1"} {
		if !IsPrerelease(v) {
			t.Errorf("%q should be a prerelease", v)
		}
	}
	for _, v := range []string{"1.0.0", "v2.3.4", "0.0.1", "10.20.30"} {
		if IsPrerelease(v) {
			t.Errorf("%q should be stable", v)
		}
	}
}

func TestLatestStablePrefersStableOverHigherPrerelease(t *testing.T) {
	got, ok := LatestStable([]string{"1.0.0", "2.0.0-rc1", "0.9.0"})
	if !ok || got != "1.0.0" {
		t.Errorf("got %q, want 1.0.0: a stable release outranks a higher prerelease", got)
	}
}

func TestLatestStableFallsBackToPrerelease(t *testing.T) {
	got, ok := LatestStable([]string{"2.0.0-rc1", "1.0.0-beta"})
	if !ok || got != "2.0.0-rc1" {
		t.Errorf("got %q, want the highest prerelease when nothing stable exists", got)
	}
}

func TestLatestStableOrdersNumerically(t *testing.T) {
	got, _ := LatestStable([]string{"0.9.0", "0.10.0", "0.2.0"})
	if got != "0.10.0" {
		t.Errorf("got %q, want 0.10.0: 10 outranks 9 numerically, not lexically", got)
	}
}

func TestLatestStableHandlesTagPrefixes(t *testing.T) {
	got, _ := LatestStable([]string{"v0.1.0", "v0.0.9"})
	if got != "v0.1.0" {
		t.Errorf("got %q, want v0.1.0 (the tag spelling is preserved)", got)
	}
}

func TestLatestStableOnEmpty(t *testing.T) {
	if _, ok := LatestStable(nil); ok {
		t.Error("an empty list has no latest version")
	}
	if _, ok := LatestStableIndex([]string{}); ok {
		t.Error("an empty list has no index")
	}
}

func TestLatestStableIsDeterministic(t *testing.T) {
	in := []string{"1.0.0", "1.0.0", "0.9.0"}
	first, _ := LatestStableIndex(in)
	for i := 0; i < 5; i++ {
		again, _ := LatestStableIndex(in)
		if again != first {
			t.Fatalf("index not stable: %d vs %d", first, again)
		}
	}
}

func TestLessIsATotalOrder(t *testing.T) {
	if Less("1.0.0", "1.0.0") {
		t.Error("a version is not less than itself")
	}
	if !Less("1.0.0-rc1", "1.0.0") {
		t.Error("a prerelease sorts below its release")
	}
	if Less("1.0.1", "1.0.0") {
		t.Error("1.0.1 outranks 1.0.0")
	}
}
