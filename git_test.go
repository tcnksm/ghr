package main

import (
	. "github.com/onsi/gomega"

	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGitConfig(t *testing.T) {
	RegisterTestingT(t)

	reset := withGitConfigFile(`
[user]
    name = tc

[remote "origin"]
    url = https://github.com/tcnksm/test.git
`)

	defer reset()

	owner, err := gitConfig("user.name")
	Expect(err).NotTo(HaveOccurred())
	Expect(owner).To(Equal("tc"))

	url, err := gitConfig("remote.origin.url")
	Expect(err).NotTo(HaveOccurred())
	Expect(url).To(Equal("https://github.com/tcnksm/test.git"))
}

func TestGitRepoName(t *testing.T) {
	RegisterTestingT(t)

	repo := GitRepoName("https://github.com/tcnksm/ghr.git")
	Expect(repo).To(Equal("ghr"))
}

func withGitConfigFile(content string) func() {
	tmpdir, err := ioutil.TempDir("", "ghr-test")
	if err != nil {
		panic(err)
	}

	tmpGitConfigFile := filepath.Join(tmpdir, "gitconfig")

	ioutil.WriteFile(
		tmpGitConfigFile,
		[]byte(content),
		0777,
	)

	prevGitConfigEnv := os.Getenv("GIT_CONFIG")
	os.Setenv("GIT_CONFIG", tmpGitConfigFile)

	return func() {
		os.Setenv("GIT_CONFIG", prevGitConfigEnv)
	}
}
