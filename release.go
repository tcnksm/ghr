package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

type ReleaseRequest struct {
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish"`
	Draft           bool   `json:"draft"`
	Prerelease      bool   `json:"prerelease"`
}

type Released struct {
	ID      int    `json:"id"`
	TagName string `json:"tag_name"`
}

type Created struct {
	ID int `json:"id"`
}

const (
	RELEASE_URL = "https://api.github.com/repos/%s/%s/releases"
)

func releaseURL(info *Info) string {
	return fmt.Sprintf(RELEASE_URL, info.OwnerName, info.RepoName)
}

func releaseRequest(info *Info) ([]byte, error) {
	params := &ReleaseRequest{
		TagName:         info.TagName,
		TargetCommitish: info.TargetCommitish,
		Draft:           info.Draft,
		Prerelease:      info.Prerelease,
	}

	payload, err := json.Marshal(params)
	if err != nil {
		return nil, nil
	}

	return payload, nil
}

func CreatedID(r io.Reader) (int, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return -1, err
	}

	var created Created
	err = json.Unmarshal(body, &created)
	if err != nil {
		return -1, err
	}
	return created.ID, nil
}

func SearchIDByTag(r io.Reader, tag string) (int, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return -1, err
	}

	var releases []Released
	err = json.Unmarshal(body, &releases)
	if err != nil {
		return -1, err
	}

	for _, release := range releases {
		if release.TagName == tag {
			return release.ID, nil
		}
	}

	return -1, nil
}
