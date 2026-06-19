package extract

import (
	"reflect"
	"testing"
)

func TestGenArgs(t *testing.T) {
	got := GenArgs("/cache/terraform-provider-random", "random", "/out")
	want := []string{"gen", "--provider", "/cache/terraform-provider-random", "--identity", "random", "--out", "/out"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GenArgs = %v, want %v", got, want)
	}
}

func TestGenArgsMinimal(t *testing.T) {
	got := GenArgs("/bin/provider-alpha", "", "")
	want := []string{"gen", "--provider", "/bin/provider-alpha"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GenArgs minimal = %v, want %v", got, want)
	}
}
