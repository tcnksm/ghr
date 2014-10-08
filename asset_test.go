package main

import (
	. "github.com/onsi/gomega"
	"strings"
	"testing"
)

func TestListAssetURL(t *testing.T) {
	RegisterTestingT(t)

	info := &Info{
		OwnerName: "tcnksm",
		RepoName:  "ghr",
		ID:        1234,
	}

	url := listAssetsURL(info)
	Expect(url).To(Equal("https://api.github.com/repos/tcnksm/ghr/releases/1234/assets"))
}

func TestExtractDeleteTarget(t *testing.T) {
	RegisterTestingT(t)

	uploads := []string{"example1.zip", "example3.zip"}

	json := `
[
  {
    "url": "https://api.github.com/repos/octocat/Hello-World/releases/assets/1",
    "browser_download_url": "https://github.com/octocat/Hello-World/releases/download/v1.0.0/example1.zip",
    "id": 1,
    "name": "example1.zip",
    "label": "short description",
    "state": "uploaded"
  },
  {
    "url": "https://api.github.com/repos/octocat/Hello-World/releases/assets/2",
    "browser_download_url": "https://github.com/octocat/Hello-World/releases/download/v1.0.0/example2.zip",
    "id": 2,
    "name": "example2.zip",
    "label": "short description",
    "state": "uploaded"
  },
  {
    "url": "https://api.github.com/repos/octocat/Hello-World/releases/assets/3",
    "browser_download_url": "https://github.com/octocat/Hello-World/releases/download/v1.0.0/example3.zip",
    "id": 3,
    "name": "example3.zip",
    "label": "short description",
    "state": "uploaded"
  }
]
`
	targets, err := SearchDeleteTargets(strings.NewReader(json), uploads)
	Expect(err).NotTo(HaveOccurred())
	Expect(targets).To(
		Equal([]DeleteTarget{
			{Name: "example1.zip", AssetId: 1},
			{Name: "example3.zip", AssetId: 3}}))
}
