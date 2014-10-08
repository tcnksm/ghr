package main

import (
	"fmt"
)

const (
	// DELETE /repos/:owner/:repo/releases/:id
	DELETE_RELEASE_URL = "https://api.github.com/repos/%s/%s/releases/%d"
)

func deleteReleaseURL(info *Info) string {
	return fmt.Sprintf(DELETE_RELEASE_URL, info.OwnerName, info.RepoName, info.ID)
}
