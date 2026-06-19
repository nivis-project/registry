package seed

import "testing"

// TestAnchorsPresent guards the invariants the seed must always satisfy
// (the real Select() will be tested against fixture popularity data).
func TestAnchorsPresent(t *testing.T) {
	if len(MustInclude) == 0 {
		t.Fatal("MustInclude must not be empty")
	}
	foundProxmox := false
	for _, p := range MustInclude {
		if p == "Telmate/proxmox" {
			foundProxmox = true
		}
	}
	if !foundProxmox {
		t.Error("seed must always include Telmate/proxmox (the most-popular proxmox provider)")
	}
	if MaxSeed != 50 {
		t.Errorf("MaxSeed = %d, want 50", MaxSeed)
	}
}

func TestUtilityAllowlistNonEmpty(t *testing.T) {
	if len(UtilityAllowlist) == 0 {
		t.Fatal("UtilityAllowlist must not be empty")
	}
}
