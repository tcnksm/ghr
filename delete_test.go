package main

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestDeleteURL(t *testing.T) {
	RegisterTestingT(t)

	info := &Info{
		ID:        123,
		OwnerName: "taichi",
		RepoName:  "tool",
	}

	url := deleteReleaseURL(info)
	Expect(url).To(Equal("https://api.github.com/repos/taichi/tool/releases/123"))
}
