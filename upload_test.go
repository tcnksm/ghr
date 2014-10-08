package main

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestUploadURL(t *testing.T) {
	RegisterTestingT(t)

	info := &Info{
		ID:        123,
		OwnerName: "tc",
		RepoName:  "tool",
	}

	url := uploadURL(info, "tool.zip")
	Expect(url).To(Equal("https://uploads.github.com/repos/tc/tool/releases/123/assets?name=tool.zip"))
}
