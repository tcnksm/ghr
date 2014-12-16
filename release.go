package main

import (
	"github.com/google/go-github/github"
)

const (
	ReleaseIDNotFound int = 0
)

// CreateRelease creates release on GitHub.
// If release already exists, it just set `GitHubAPIOpts.ID`.
// If release already exists and `--delete` option is provided,
// delete it and re-create release.
func CreateRelease(ghrOpts *GhrOpts, apiOpts *GitHubAPIOpts) (err error) {

	// Get release ID
	err = GetReleaseID(apiOpts)
	if err != nil {
		return err
	}

	// Delte release if `--delete` is set
	if ghrOpts.Delete {
		err = DeleteRelease(apiOpts)
		if err != nil {
			return err
		}
	}

	// If release is exist, do nothing
	if apiOpts.ID != ReleaseIDNotFound {
		return nil
	}

	// Create client
	client := NewOAuthedClient(apiOpts.Token)

	// Create Release
	request := CreateReleaseRequest(apiOpts)
	rel, res, err := client.Repositories.CreateRelease(apiOpts.OwnerName, apiOpts.RepoName, request)
	if err != nil {
		return err
	}

	err = CheckStatusCreated(res)
	if err != nil {
		return err
	}

	Debug("CreateRelease:", rel)

	// Set ReleaseID and UploadURL
	apiOpts.ID = *rel.ID
	apiOpts.UploadURL = *rel.UploadURL

	return nil
}

// GetRleaseID gets release ID
// If it's not exist, it sets ReleaseIDNotFound(=0) to `GithubAPIOpts.ID`
func GetReleaseID(apiOpts *GitHubAPIOpts) (err error) {
	// Create client
	client := NewOAuthedClient(apiOpts.Token)

	// Fetch all releases on GitHub
	releases, res, err := client.Repositories.ListReleases(apiOpts.OwnerName, apiOpts.RepoName, nil)
	if err != nil {
		return err
	}

	// Check request is succeeded.
	err = CheckStatusOK(res)
	if err != nil {
		return err
	}

	// Check relase already exists or not
	for _, r := range releases {
		if *r.TagName == apiOpts.TagName {

			// Set ID if relase is already exist
			apiOpts.ID = *r.ID
			apiOpts.UploadURL = *r.UploadURL

			// Debug
			Debug("GetRelease(ID, UploadURL):", *r.ID, *r.UploadURL)

			return nil
		}
	}

	// Set const value to tell other func there is no release
	apiOpts.ID = ReleaseIDNotFound
	apiOpts.UploadURL = ""

	return nil
}

// DeleteRelease delete release which is related to release ID
// It also deletes its tag
func DeleteRelease(apiOpts *GitHubAPIOpts) (err error) {

	// Check Release is already exist or not
	if apiOpts.ID == ReleaseIDNotFound {
		return nil
	}

	// Create client
	client := NewOAuthedClient(apiOpts.Token)

	// Delete release.
	res, err := client.Repositories.DeleteRelease(apiOpts.OwnerName, apiOpts.RepoName, apiOpts.ID)
	if err != nil {
		return err
	}

	// Check deleting release is succeeded.
	err = CheckStatusNoContent(res)
	if err != nil {
		return err
	}

	// Delete tag related to its release
	ref := "tags/" + apiOpts.TagName
	res, err = client.Git.DeleteRef(apiOpts.OwnerName, apiOpts.RepoName, ref)
	if err != nil {
		return err
	}

	// Check deleting release is succeeded.
	err = CheckStatusNoContent(res)
	if err != nil {
		return err
	}

	// Set const value to tell other func there is no release
	apiOpts.ID = ReleaseIDNotFound
	apiOpts.UploadURL = ""

	return nil
}

// CreateReleaseRequest creates request for CreateRelease
func CreateReleaseRequest(apiOpts *GitHubAPIOpts) *github.RepositoryRelease {
	return &github.RepositoryRelease{
		TagName:         &apiOpts.TagName,
		Draft:           &apiOpts.Draft,
		Prerelease:      &apiOpts.Prerelease,
		TargetCommitish: &apiOpts.TargetCommitish,
	}
}
