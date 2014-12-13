package main

import (
	"fmt"
	"net/http"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
)

// GitHubAPIOpts are the options for GitHub API.
type GitHubAPIOpts struct {
	// Token is the GitHub API token
	Token string

	// ID is release ID. Each release has its unique ID
	ID int

	// Tag is release tag.
	TagName string

	// RepoName is repository name.
	RepoName string

	// OwnerName is repositoy owner name.
	OwnerName string

	// TargetCommitish specifies the commitish value that determines
	// where the Git tag is created from.
	TargetCommitish string

	// Body describes the contens of the tag.
	Body string

	// If Draft is true, release would be draft (unpublished).
	Draft bool

	// If Prerelease is true, release would be prerelease.
	Prerelease bool

	// UPloadURL is URL to upload artifacts.
	UploadURL string
}

// NewOAuthedClient create client with oauth
func NewOAuthedClient(token string) *github.Client {
	// Create OAuth client
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: token},
	}

	// Create GitHub API client with OAuth client
	return github.NewClient(t.Client())
}

// checkStatusOK checks http status returned by API is 200
func CheckStatusOK(res *github.Response) error {
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Github returned %s\n", res.Status)
	}
	return nil
}

// CheckStatusCreated checks http status returned by API is 201
func CheckStatusCreated(res *github.Response) error {
	if res.StatusCode != http.StatusCreated {
		if res.StatusCode == 422 {
			return fmt.Errorf("Github returned %s (this is probably because the release already exists)\n", res.Status)
		}
		return fmt.Errorf("Github returned %s\n", res.Status)
	}

	return nil
}

// checkStatusNoContent checks http status returned by API is 204
// In github API, this means DELETE request is success.
func CheckStatusNoContent(res *github.Response) error {
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Github returned %s\n", res.Status)
	}
	return nil
}
