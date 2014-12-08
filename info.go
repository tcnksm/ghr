package main

import (
	"fmt"
	"github.com/tcnksm/go-gitconfig"
	"os"
)

type Info struct {
	ID              int
	Token           string
	TagName         string
	RepoName        string
	OwnerName       string
	TargetCommitish string
	Body            string
	Draft           bool
	Prerelease      bool
}

// setToken sets GITHUB TOKEN.
func setToken(token *string) error {
	// Use flag value
	if *token != "" {
		return nil
	}

	// Use Environmental value
	if os.Getenv("GITHUB_TOKEN") != "" {
		*token = os.Getenv("GITHUB_TOKEN")
		return nil
	}

	// Use .gitconfig value
	*token, _ = gitconfig.GithubToken()
	if *token == "" {
		return fmt.Errorf("Please set your Github API Token in the GITHUB_TOKEN env var\n")
	}

	return nil
}

func NewInfo(tag string) (*Info, error) {

	var (
		err error
	)

	err = setToken(token)
	if err != nil {
		return nil, err
	}
	debug("token:", *token)

	if *owner == "" {
		*owner, err = gitconfig.Username()
		if err != nil {
			return nil, fmt.Errorf("Cound not retrieve git user name\n")
		}
	}
	debug("owner:", *owner)

	if *repo == "" {
		*repo, err = gitconfig.Repository()
		if err != nil {
			return nil, fmt.Errorf("Cound not retrieve repository name\n")
		}
	}
	debug("repository:", *repo)

	return &Info{
		TagName:         tag,
		Token:           *token,
		OwnerName:       *owner,
		RepoName:        *repo,
		TargetCommitish: "master",
		Draft:           *flDraft,
		Prerelease:      *flPrerelease,
	}, nil
}
