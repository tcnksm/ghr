package main

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"
)

var RepoNameRegexp = regexp.MustCompile(`.+/([^/]+)\.git$`)

func GitRepoName(url string) string {
	matched := RepoNameRegexp.FindStringSubmatch(url)
	return matched[1]
}

func GitRemote() (string, error) {
	return gitConfig("--local", "remote.origin.url")
}

func GitOwner() (string, error) {
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
