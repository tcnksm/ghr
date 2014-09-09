package main

import (
	"github.com/tcnksm/go-gitconfig"

	"fmt"
	"regexp"
	"strings"
)

func GetOwnerName() (string, error) {
	owner, err := gitconfig.Username()
	if err != nil || owner == "" {
		return "", fmt.Errorf("Cound not retrieve git user name\n")
	}
	return owner, err
}

func GetRepoName() (string, error) {
	url, err := gitconfig.OriginURL()
	if err != nil || url == "" {
		return "", fmt.Errorf("Cound not retrieve remote repository url\n")
	}
	repo := retrieveRepoName(url)
	if repo == "" {
		return "", fmt.Errorf("Cound not retrieve repository name\n")
	}
	return repo, nil
}

var RepoNameRegexp = regexp.MustCompile(`.+/([^/]+)(\.git)?$`)

func retrieveRepoName(url string) string {
	matched := RepoNameRegexp.FindStringSubmatch(url)
	return strings.TrimRight(matched[1], ".git")
}
