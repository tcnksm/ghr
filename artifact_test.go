package main

import (
	. "github.com/onsi/gomega"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestArtifacts(t *testing.T) {
	RegisterTestingT(t)

	files := []string{
		"ghr_darwin_386.zip",
		"ghr_linux_amd64.zip",
		"ghr_darwin_amd64.zip",
	}

	path := withCreateArtifact(files)
	artifacts, err := Artifacts(path)
	Expect(err).NotTo(HaveOccurred())
	Expect(artifacts).To(Equal([]string{
		path + "/ghr_darwin_386.zip",
		path + "/ghr_darwin_amd64.zip",
		path + "/ghr_linux_amd64.zip",
	}))

}

func TestArtifactNames(t *testing.T) {
	RegisterTestingT(t)

	files := []string{
		"ghr_darwin_386.zip",
		"ghr_linux_amd64.zip",
		"ghr_darwin_amd64.zip",
	}

	path := withCreateArtifact(files)
	artifacts, _ := Artifacts(path)
	names := ArtifactNames(artifacts)
	Expect(names).To(Equal([]string{
		"ghr_darwin_386.zip",
		"ghr_darwin_amd64.zip",
		"ghr_linux_amd64.zip",
	}))
}

func withCreateArtifact(files []string) string {
	tmpDir, err := ioutil.TempDir("", "ghr")
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		ioutil.WriteFile(
			filepath.Join(tmpDir, f),
			[]byte("test contents"),
			0777,
		)
	}
	return tmpDir
}
