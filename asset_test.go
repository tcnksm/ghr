package main

import (
	"io/ioutil"
	"testing"
)

func TestGetLocalAssets(t *testing.T) {

	// Create tmp assets
	path := genAssets(t)

	out, err := GetLocalAssets(path)
	if err != nil {
		t.Error("expected err is not happened")
	}

	if len(out) != 3 {
		t.Errorf("expected %d assets to eq %d assets", len(out), 3)
	}

}

func TestGetLocalAssets_noAssets(t *testing.T) {
	// Create empty directory
	path, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("err: %s", err)
	}

	out, err := GetLocalAssets(path)
	if err == nil {
		t.Errorf("expected error is happened, got '%s'", err)
	}

	if len(out) != 0 {
		t.Errorf("expected %d assets to eq %d assets", len(out), 0)
	}

}
