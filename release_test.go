package main

import (
	. "github.com/onsi/gomega"
	"strings"
	"testing"
)

func TestReleaseURL(t *testing.T) {
	RegisterTestingT(t)

	info := Info{
		OwnerName: "tc",
		RepoName:  "tool",
	}

	url := releaseURL(info)
	Expect(url).To(Equal("https://api.github.com/repos/tc/tool/releases"))
}

func TestCreatedID(t *testing.T) {
	RegisterTestingT(t)

	json := `
{
  "url": "https://api.github.com/repos/octocat/Hello-World/releases/1",
  "html_url": "https://github.com/octocat/Hello-World/releases/v1.0.0",
  "tarball_url": "https://api.github.com/repos/octocat/Hello-World/tarball/v1.0.0",
  "zipball_url": "https://api.github.com/repos/octocat/Hello-World/zipball/v1.0.0",
  "id": 123,
  "tag_name": "v1.0.0",
  "target_commitish": "master",
  "name": "v1.0.0",
  "body": "Description of the release",
  "draft": false,
  "prerelease": false,
  "created_at": "2013-02-27T19:35:32Z",
  "published_at": "2013-02-27T19:35:32Z"
}`

	id, err := CreatedID(strings.NewReader(json))
	Expect(err).NotTo(HaveOccurred())
	Expect(id).To(Equal(123))
}

func TestSearchIDByTag(t *testing.T) {
	RegisterTestingT(t)

	json := `[
{
    "url": "https://api.github.com/repos/octocat/Hello-World/releases/1",
    "html_url": "https://github.com/octocat/Hello-World/releases/v1.0.0",
    "id": 123,
    "tag_name": "v1.0.0",
    "target_commitish": "master",
    "name": "v1.0.0"
},
{
    "url": "https://api.github.com/repos/octocat/Hello-World/releases/1",
    "html_url": "https://github.com/octocat/Hello-World/releases/v1.0.2",
    "id": 456,
    "tag_name": "v1.0.2",
    "target_commitish": "master",
    "name": "v1.0.2"
}
]`

	id, err := SearchIDByTag(strings.NewReader(json), "v1.0.0")
	Expect(err).NotTo(HaveOccurred())
	Expect(id).To(Equal(123))

	id, err = SearchIDByTag(strings.NewReader(json), "v1.0.1")
	Expect(err).NotTo(HaveOccurred())
	Expect(id).To(Equal(-1))

	id, err = SearchIDByTag(strings.NewReader(json), "v1.0.2")
	Expect(err).NotTo(HaveOccurred())
	Expect(id).To(Equal(456))

	id, err = SearchIDByTag(strings.NewReader(`Not json string`), "v1.0.0")
	Expect(err).To(HaveOccurred())

}

func TestReleaseRequest(t *testing.T) {
	RegisterTestingT(t)

	info := Info{
		TagName:         "v1.0.0",
		TargetCommitish: "master",
		Draft:           false,
		Prerelease:      false,
	}

	json := []byte(`{"tag_name":"v1.0.0","target_commitish":"master","draft":false,"prerelease":false}`)

	body, err := releaseRequest(info)
	Expect(err).NotTo(HaveOccurred())
	Expect(body).To(Equal(json))
}
