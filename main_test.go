package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// genAssets generates test assets in tmpDir
func genAssets(t *testing.T) string {
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	for _, f := range []string{"fileA", "fileB", "fileC"} {
		file := filepath.Join(tmpDir, f)
		ioutil.WriteFile(file, []byte("test"), 0777)
	}

	return tmpDir
}

// setEnv set enviromental variables and return restore function.
func setEnv(key, val string) func() {

	preVal := os.Getenv(key)
	os.Setenv(key, val)

	return func() {
		os.Setenv(key, preVal)
	}

}
