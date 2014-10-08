package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
)

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

func GetReleaseID(info *Info) (int, error) {
	requestURL := releaseURL(info)
	debug(requestURL)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return -1, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}
	debug(res.Status)

	err = checkStatusOK(res.StatusCode, res.Status)
	if err != nil {
		return -1, err
	}

	defer res.Body.Close()
	return SearchIDByTag(res.Body, info.TagName)
}

func DeleteRelease(info *Info) error {
	requestURL := deleteReleaseURL(info)
	debug(requestURL)

	req, err := http.NewRequest("DELETE", requestURL, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", info.Token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	debug(res.Status)

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Github returned %s\n", res.Status)
	}
	return nil
}

func CreateNewRelease(info *Info) (int, error) {

	requestURL := releaseURL(info)
	debug(requestURL)

	requestBody, err := releaseRequest(info)
	req, err := http.NewRequest("POST", requestURL, bytes.NewReader(requestBody))
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
	debug(res.Status)

	err = checkStatusCreated(res.StatusCode, res.Status)
	if err != nil {
		return -1, err
	}

	defer res.Body.Close()
	return CreatedID(res.Body)
}

func UploadAsset(info *Info, path string) error {

	file, err := os.Stat(path)
	requestURL := uploadURL(info, file.Name())

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
	debug(res.Status)

	err = checkStatusCreated(res.StatusCode, res.Status)
	if err != nil {
		return err
	}

	return nil
}
