package utils

import "testing"

func TestCompareSlice(t *testing.T) {
	a := []string{"10.0.0.1/32", "10.0.0.2/32", "10.0.0.3/32"}
	b := []string{"10.0.0.1/32", "10.0.0.2/32", "10.0.0.3/32"}
	gotA, gotB := CompareSlice(a, b)
	//Expect to be empty
	if len(gotA)+len(gotB) != 0 {
		t.Errorf("CompareSlice should return No diff for these slices %q, %q", a, b)
	}
	a = []string{"10.0.0.1/32", "10.0.0.2/32", "10.0.0.3/32"}
	b = []string{"10.0.0.1/32", "10.0.0.2/32", "10.0.0.4/32"}
	gotA, gotB = CompareSlice(a, b)

	//Expect to get values
	if gotA[0] != "10.0.0.3/32" {
		t.Errorf("When Comparing Slice %q and Slice %q , expected diff 10.0.0.3/32 got %q", a, b, gotA)
	}

	if gotB[0] != "10.0.0.4/32" {
		t.Errorf("When Comparing Slice %q and Slice %q , expected diff 10.0.0.4/32 got %q", a, b, gotB)
	}
}
