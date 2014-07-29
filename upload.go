package main

import (
	"fmt"
	"net/url"
)

const (
	UPLOAD_URL = "https://uploads.github.com/repos/%s/%s/releases/%d/assets"
)

func uploadURL(info Info, name string) string {
	v := url.Values{}
	v.Set("name", name)

	return fmt.Sprintf(UPLOAD_URL, info.OwnerName, info.RepoName, info.ID) + "?" + v.Encode()
}
