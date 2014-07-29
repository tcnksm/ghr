package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"
)

var RepoNameRegexp = regexp.MustCompile(`.+/([^/]+)\.git$`)

func GetRepoName() (string, error) {
	url, err := gitRemote()
	if err != nil || url == "" {
		return "", fmt.Errorf("Cound not retrieve remote repository url\n")
	}

	repo := retrieveRepoName(url)
	if repo == "" {
		return "", fmt.Errorf("Cound not retrieve repository name\n")
	}
	return repo, nil
}

func GetOwnerName() (string, error) {
	owner, err := gitOwner()
	if err != nil || owner == "" {
		return "", fmt.Errorf("Cound not retrieve git user name\n")
	}
	return owner, err
}

func retrieveRepoName(url string) string {
	matched := RepoNameRegexp.FindStringSubmatch(url)
	return matched[1]
}

// git config --local remote.origin.url
func gitRemote() (string, error) {
	return gitConfig("--local", "remote.origin.url")
}

// git config --global user.name
func gitOwner() (string, error) {
	return gitConfig("--global", "user.name")
}

func gitConfig(args ...string) (string, error) {
	gitArgs := append([]string{"config", "--get", "--null"}, args...)
	var stdout bytes.Buffer
	cmd := exec.Command("git", gitArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = ioutil.Discard

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimRight(stdout.String(), "\000"), nil
}
