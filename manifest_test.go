package main

import (
	"testing"
)

func TestNewManifestFromFile(t *testing.T) {
	man, err := NewManifestFromFile("./testdata/chaoskube-release.yaml")
	if err != nil {
		t.Fatalf("Unexpected error : %q", err)
	}
	if man.Name != "chaoskube" {
		t.Errorf("Unexpected Name : %q", man.Name)
	}
}
