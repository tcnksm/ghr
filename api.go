package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

const (
	RELEASE_URL = "https://api.github.com/repos/%s/%s/releases"
	UPLOAD_URL  = "https://uploads.github.com/repos/%s/%s/releases/%d/assets"
)

type Releases struct {
	ID      int    `json:"id"`
	TagName string `json:"tag_name"`
}

type ReleaseRequest struct {
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish"`
	Draft           bool   `json:"draft"`
	Prerelease      bool   `json:"prerelease"`
}

type ReleaseResponse struct {
	ID int `json:"id"`
}

func debugResponseBody(body io.ReadCloser) {
	if os.Getenv("DEBUG") != "" {
		body, _ := ioutil.ReadAll(body)
		log.Println(string(body))
	}
}

func uploadAsset(info Info, path string) error {
	file, err := os.Stat(path)
	if err != nil {
		return err
	}

	if file.IsDir() {
		fmt.Fprintf(os.Stderr, "`%s` is directory, skip it\n", path)
		return nil
	}

	v := url.Values{}
	v.Set("name", file.Name())

	requestURL := fmt.Sprintf(UPLOAD_URL, info.OwnerName, info.RepoName, info.ID) + "?" + v.Encode()
	debug(requestURL)

	f, err := os.Open(path)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", requestURL, f)
	if err != nil {
		return err
	}

	req.ContentLength = file.Size()
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", info.Token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	debug(res.Status)

	if res.StatusCode != http.StatusCreated {
		if res.StatusCode == 422 {
			return fmt.Errorf("Github returned %s (this is probably because the release already exists)\n", res.Status)
		}
		return fmt.Errorf("Github returned %s\n", res.Status)
	}

	return nil
}

func GetReleaseID(info Info) (int, error) {

	requestURL := fmt.Sprintf(RELEASE_URL, info.OwnerName, info.RepoName)
	debug(requestURL)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return -1, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()

	debug(res.Status)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return -1, err
	}
	debug(string(body))

	var releases []Releases
	err = json.Unmarshal(body, &releases)
	if err != nil {
		return -1, err
	}

	for _, release := range releases {
		if release.TagName == info.TagName {
			return release.ID, nil
		}
	}

	return -1, nil
}

func CreateNewRelease(info Info) (int, error) {

	requestURL := fmt.Sprintf(RELEASE_URL, info.OwnerName, info.RepoName)
	debug(requestURL)

	params := ReleaseRequest{
		TagName:         info.TagName,
		TargetCommitish: info.TargetCommitish,
		Draft:           info.Draft,
		Prerelease:      info.Prerelease,
	}

	payload, err := json.Marshal(params)
	if err != nil {
		return -1, err
	}
	debug(string(payload))

	reader := bytes.NewReader(payload)
	req, err := http.NewRequest("POST", requestURL, reader)
	if err != nil {
		return -1, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", info.Token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()

	debug(res.Status)

	if res.StatusCode != http.StatusCreated {
		if res.StatusCode == 422 {
			return -1, fmt.Errorf("Github returned %s (this is probably because the release already exists)", res.Status)
		}
		return -1, fmt.Errorf("Github returned %s", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return -1, err
	}
	debug(string(body))

	var releaseResponse ReleaseResponse
	err = json.Unmarshal(body, &releaseResponse)
	if err != nil {
		return -1, err
	}

	return releaseResponse.ID, nil
}
