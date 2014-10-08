package main

import (
	"fmt"
)

const (
	// DELETE /repos/:owner/:repo/releases/:id
	DELETE_RELEASE_URL = "https://api.github.com/repos/%s/%s/releases/%d"

	// DELETE /repos/:owner/:repo/releases/assets/:id
	DELETE_ASSET_URL = "https://api.github.com/repos/%s/%s/releases/assets/%d"
)

func deleteReleaseURL(info *Info) string {
	return fmt.Sprintf(DELETE_RELEASE_URL, info.OwnerName, info.RepoName, info.ID)
}

func deleteAssetURL(info *Info, assetId int) string {
	return fmt.Sprintf(DELETE_ASSET_URL, info.OwnerName, info.RepoName, assetId)
}
