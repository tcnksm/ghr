package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

type Assets struct {
	Name    string `json:"name"`
	AssetId int    `json:"id"`
}

type DeleteTarget Assets

const (
	// GET /repos/:owner/:repo/releases/:id/assets
	LIST_ASSET_URL = "https://api.github.com/repos/%s/%s/releases/%d/assets"
)

func listAssetsURL(info *Info) string {
	return fmt.Sprintf(LIST_ASSET_URL, info.OwnerName, info.RepoName, info.ID)
}

// DeleteTargets extract asset IDs which is already uploaded
// It decide `already uploaded` based on filename to upload
func SearchDeleteTargets(r io.Reader, uploads []string) ([]DeleteTarget, error) {

	targets := []DeleteTarget{}
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return targets, err
	}

	var assetsUploaded []Assets
	err = json.Unmarshal(body, &assetsUploaded)
	if err != nil {
		return targets, err
	}

	for _, upload := range uploads {
		for _, asset := range assetsUploaded {
			if upload == asset.Name {
				targets = append(targets,
					DeleteTarget{Name: asset.Name, AssetId: asset.AssetId})
			}
		}
	}
	return targets, nil
}
