package generate

import "testing"

func TestContractPath(t *testing.T) {
	got := ContractPath("hashicorp", "random", "v3.9.0")
	want := "registry/docs/providers/hashicorp/random/v3.9.0/index.json"
	if got != want {
		t.Errorf("ContractPath = %q, want %q", got, want)
	}
}
