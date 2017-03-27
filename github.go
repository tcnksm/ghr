package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

var (
	RelaseNotFound = errors.New("release is not found")
)

type GitHub interface {
	CreateRelease(ctx context.Context, req *github.RepositoryRelease) (*github.RepositoryRelease, error)
	GetRelease(ctx context.Context, tag string) (*github.RepositoryRelease, error)
	DeleteRelease(ctx context.Context, releaseID int) error
	DeleteTag(ctx context.Context, tag string) error

	UploadAsset(ctx context.Context, releaseID int, filename string) (*github.ReleaseAsset, error)
	DeleteAsset(ctx context.Context, assetID int) error
	ListAssets(ctx context.Context, releaseID int) ([]*github.ReleaseAsset, error)

	SetUploadURL(urlStr string) error
}

type GitHubClient struct {
	Owner, Repo string
	*github.Client
}

func NewGitHubClient(owner, repo, token string, urlStr string) (GitHub, error) {
	if len(owner) == 0 {
		return nil, errors.New("missing GitHub repository owner")
	}

	if len(owner) == 0 {
		return nil, errors.New("missing GitHub repository name")
	}

	if len(token) == 0 {
		return nil, errors.New("missing GitHub API token")
	}

	if len(urlStr) == 0 {
		return nil, errors.New("missgig GitHub API URL")
	}

	baseURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Github API URL")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)
	client.BaseURL = baseURL

	return &GitHubClient{
		Owner:  owner,
		Repo:   repo,
		Client: client,
	}, nil
}

func (c *GitHubClient) SetUploadURL(urlStr string) error {
	i := strings.Index(urlStr, "repos/")
	parsedURL, err := url.ParseRequestURI(urlStr[:i])
	if err != nil {
		return errors.Wrap(err, "faield to parse upload URL")
	}

	c.UploadURL = parsedURL
	return nil
}

func (c *GitHubClient) CreateRelease(ctx context.Context, req *github.RepositoryRelease) (*github.RepositoryRelease, error) {

	release, res, err := c.Repositories.CreateRelease(c.Owner, c.Repo, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a release")
	}

	if res.StatusCode != http.StatusCreated {
		return nil, errors.Errorf("create release: invalid status: %s", res.Status)
	}

	return release, nil
}

func (c *GitHubClient) GetRelease(ctx context.Context, tag string) (*github.RepositoryRelease, error) {
	// Check Release is already exist or not
	release, res, err := c.Repositories.GetReleaseByTag(c.Owner, c.Repo, tag)
	if err != nil {
		if res == nil {
			return nil, errors.Wrapf(err, "failed to get release tag: %s", tag)
		}

		// TODO(tcnksm): Handle invalid token
		if res.StatusCode != http.StatusNotFound {
			return nil, errors.Wrapf(err,
				"get release tag: invalid status: %s", res.Status)
		}

		return nil, RelaseNotFound
	}

	return release, nil
}

func (c *GitHubClient) DeleteRelease(ctx context.Context, releaseID int) error {
	res, err := c.Repositories.DeleteRelease(c.Owner, c.Repo, releaseID)
	if err != nil {
		return errors.Wrap(err, "failed to delete release")
	}

	if res.StatusCode != http.StatusNoContent {
		return errors.Errorf("delete release: invalid status: %s", res.Status)
	}

	return nil
}

func (c *GitHubClient) DeleteTag(ctx context.Context, tag string) error {
	ref := fmt.Sprintf("tags/%s", tag)
	res, err := c.Git.DeleteRef(c.Owner, c.Repo, ref)
	if err != nil {
		return errors.Wrapf(err, "failed to delete tag: %s", ref)
	}

	if res.StatusCode != http.StatusNoContent {
		return errors.Errorf("delete tag: invalid status: %s", res.Status)
	}

	return nil
}

func (c *GitHubClient) UploadAsset(ctx context.Context, releaseID int, filename string) (*github.ReleaseAsset, error) {

	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get abs path")
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	opts := &github.UploadOptions{
		// Use base name by default
		Name: filepath.Base(filename),
	}

	asset, res, err := c.Repositories.UploadReleaseAsset(c.Owner, c.Repo, releaseID, opts, f)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to upload release asset: %s", filename)
	}

	switch res.StatusCode {
	case http.StatusCreated:
		return asset, nil
	case 422:
		return nil, errors.Errorf(
			"upload release asset: invalid status code: %s",
			"422 (this is probably because the asset already uploaded)")
	default:
		return nil, errors.Errorf(
			"upload release asset: invalid status code: %s", res.Status)
	}
}

func (c *GitHubClient) DeleteAsset(ctx context.Context, assetID int) error {
	res, err := c.Repositories.DeleteReleaseAsset(c.Owner, c.Repo, assetID)
	if err != nil {
		return errors.Wrap(err, "failed to delete release asset")
	}

	if res.StatusCode != http.StatusNoContent {
		return errors.Errorf("delete release assets: invalid status code: %s", res.Status)
	}

	return nil
}

func (c *GitHubClient) ListAssets(ctx context.Context, releaseID int) ([]*github.ReleaseAsset, error) {
	result := []*github.ReleaseAsset{}
	page := 1

	for {
		assets, res, err := c.Repositories.ListReleaseAssets(c.Owner, c.Repo, releaseID, &github.ListOptions{Page: page})
		if err != nil {
			return nil, errors.Wrap(err, "failed to list assets")
		}

		if res.StatusCode != http.StatusOK {
			return nil, errors.Errorf("list release assets: invalid status code: %s", res.Status)
		}

		result = append(result, assets...)

		if res.NextPage <= page {
			break
		}

		page = res.NextPage
	}

	return result, nil
}
