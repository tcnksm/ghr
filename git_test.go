package main

import (
	. "github.com/onsi/gomega"

	"testing"
)

func TestRetrieveRepoName(t *testing.T) {
	RegisterTestingT(t)

	repo := retrieveRepoName("https://github.com/tcnksm/ghr.git")
	Expect(repo).To(Equal("ghr"))

	repo = retrieveRepoName("https://github.com/tcnksm/ghr")
	Expect(repo).To(Equal("ghr"))
}
