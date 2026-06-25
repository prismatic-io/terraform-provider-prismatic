package provider

import "testing"

func TestReadComponentBundle(t *testing.T) {
	result, err := readComponentBundle("../../test/data/component/code/")
	if err != nil {
		t.Fatalf("Failed to read component bundle: %s", err)
	}
	if result == nil {
		t.Fatalf("Received nil result from bundle read")
	}
}
