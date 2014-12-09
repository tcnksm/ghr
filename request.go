package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
)

func githubAPIClient(
	url string,
	method string,
	token string,
	check func(int, string) error) (res *http.Response, err error) {

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	addAPIHeader(req, token)
	return requestWithCheck(req, check)
}

// addHeader adds header for GitHub API
func addAPIHeader(req *http.Request, token string) {
	req.Close = true
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))
}

func requestWithCheck(
	req *http.Request,
	check func(int, string) error) (*http.Response, error) {

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	debug("HttpStatus:", res.Status)

	err = check(res.StatusCode, res.Status)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func checkStatusOK(code int, msg string) error {
	if code != http.StatusOK {
		return fmt.Errorf("Github returned %s\n", msg)
	}
	return nil
}

func checkStatusCreated(code int, msg string) error {
	if code != http.StatusCreated {
		if code == 422 {
			return fmt.Errorf("Github returned %s (this is probably because the release already exists)\n", msg)
		}
		return fmt.Errorf("Github returned %s\n", msg)
	}

	return nil
}

func checkStatusNoContent(code int, msg string) error {
	if code != http.StatusNoContent {
		return fmt.Errorf("Github returned %s\n", msg)
	}
	return nil
}

func GetReleaseID(info *Info) (int, error) {
	requestURL := releaseURL(info)
	debug("requestURL:", requestURL)

	res, err := githubAPIClient(requestURL, "GET", info.Token, checkStatusOK)
	if err != nil {
		return ID_NOT_FOUND, err
	}
	defer res.Body.Close()
	return SearchIDByTag(res.Body, info.TagName)
}

func GetDeleteTargets(info *Info, uploads []string) ([]DeleteTarget, error) {
	requestURL := listAssetsURL(info)
	debug("requestURL:", requestURL)

	res, err := githubAPIClient(requestURL, "GET", info.Token, checkStatusOK)
	if err != nil {
		return []DeleteTarget{}, err
	}
	defer res.Body.Close()
	return SearchDeleteTargets(res.Body, uploads)
}

func DeleteRelease(info *Info) error {
	requestURL := deleteReleaseURL(info)
	debug("requestURL:", requestURL)

	_, err := githubAPIClient(requestURL, "DELETE", info.Token, checkStatusNoContent)
	if err != nil {
		return err
	}

	return nil
}

func DeleteTag(info *Info) error {
	requestURL := deleteTagURL(info)
	debug("requestURL:", requestURL)

	_, err := githubAPIClient(requestURL, "DELETE", info.Token, checkStatusNoContent)
	if err != nil {
		return err
	}
	return nil
}

func DeleteAsset(info *Info, assetId int) error {
	requestURL := deleteAssetURL(info, assetId)
	debug("requestURL:", requestURL)

	_, err := githubAPIClient(requestURL, "DELETE", info.Token, checkStatusNoContent)
	if err != nil {
		return err
	}
	return nil
}

func CreateNewRelease(info *Info) (int, error) {
	requestURL := releaseURL(info)
	debug("requestURL:", requestURL)

	requestBody, err := releaseRequest(info)
	req, err := http.NewRequest("POST", requestURL, bytes.NewReader(requestBody))
	if err != nil {
		return ID_NOT_FOUND, err
	}

	addAPIHeader(req, info.Token)
	res, err := requestWithCheck(req, checkStatusCreated)
	if err != nil {
		return ID_NOT_FOUND, err
	}

	defer res.Body.Close()
	return CreatedID(res.Body)
}

func UploadAsset(info *Info, path string) error {
	file, err := os.Stat(path)
	requestURL := uploadURL(info, file.Name())
	debug("requestURL:", requestURL)

	f, err := os.Open(path)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", requestURL, f)
	if err != nil {
		return err
	}

	addAPIHeader(req, info.Token)
	req.ContentLength = file.Size()
	req.Header.Set("Content-Type", "application/octet-stream")

	_, err = requestWithCheck(req, checkStatusCreated)
	if err != nil {
		return err
	}
	return nil
}
