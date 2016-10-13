package main

import "testing"

const (
	TestDir = "./testdata"
)

func TestLocalAssets(t *testing.T) {
	localAssets, err := LocalAssets(TestDir)
	if err != nil {
		t.Fatal("LocalAssets failed:", err)
	}

	if got, want := len(localAssets), 4; got != want {
		t.Fatalf("localAssets number = %d, want %d", got, want)
	}
}
