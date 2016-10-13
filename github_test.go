package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-github/github"
)

const (
	TestOwner = "tcnksm"
	TestRepo  = "github-api-test"
)

func testGithubClient(t *testing.T) GitHub {
	token := os.Getenv(EnvGitHubToken)
	client, err := NewGitHubClient(TestOwner, TestRepo, token, defaultBaseURL)
	if err != nil {
		t.Fatal("NewGitHubClient failed:", err)
	}
	return client
}

func TestGitHubClient(t *testing.T) {
	t.Parallel()

	c := testGithubClient(t)
	testTag := "github-client"
	cases := []struct {
		Request *github.RepositoryRelease
	}{
		{
			&github.RepositoryRelease{
				TagName:    github.String(testTag),
				Draft:      github.Bool(false),
				Prerelease: github.Bool(false),
			},
		},

		{
			&github.RepositoryRelease{
				TagName:    github.String(testTag),
				Draft:      github.Bool(false),
				Prerelease: github.Bool(true),
			},
		},

		{
			&github.RepositoryRelease{
				TagName:    github.String(testTag),
				Draft:      github.Bool(true),
				Prerelease: github.Bool(false),
			},
		},
	}

	for i, tc := range cases {
		// Prevent a lot of requests at same time to GitHub API
		time.Sleep(1 * time.Second)

		created, err := c.CreateRelease(context.TODO(), tc.Request)
		if err != nil {
			t.Fatalf("#%d CreateRelease failed: %s", i, err)
		}

		// Draft release doesn't create tag. So it's not found.
		if !*created.Draft {
			got, err := c.GetRelease(context.TODO(), *created.TagName)
			if err != nil {
				t.Fatalf("#%d GetRelease failed: %s", i, err)
			}

			if *got.ID != *created.ID {
				t.Fatalf("got ID = %d, want %d", *got.ID, *created.ID)
			}
		}

		if err := c.DeleteRelease(context.TODO(), *created.ID); err != nil {
			t.Fatalf("#%d DeleteRelease failed: %s", i, err)
		}

		if *created.Draft {
			continue
		}

		if err := c.DeleteTag(context.TODO(), *tc.Request.TagName); err != nil {
			t.Fatalf("#%d DeleteTag failed: %s", i, err)
		}
	}
}

func TestGitHubClient_Upload(t *testing.T) {
	client := testGithubClient(t)
	testTag := "github-client-upload-asset"
	req := &github.RepositoryRelease{
		TagName: github.String(testTag),
		Draft:   github.Bool(true),
	}

	release, err := client.CreateRelease(context.TODO(), req)
	if err != nil {
		t.Fatal("CreateRelease failed:", err)
	}

	defer func() {
		if err := client.DeleteRelease(context.TODO(), *release.ID); err != nil {
			t.Fatalf("DeleteRelease failed: %s", err)
		}
	}()

	filename := filepath.Join("./testdata", "darwin_386")
	asset, err := client.UploadAsset(context.TODO(), *release.ID, filename)
	if err != nil {
		t.Fatal("UploadAsset failed:", err)
	}

	githubClient, ok := client.(*GitHubClient)
	if !ok {
		t.Fatal("Faield to asset to GithubClient")
	}

	rc, url, err := githubClient.Repositories.DownloadReleaseAsset(
		githubClient.Owner, githubClient.Repo, *asset.ID)
	if err != nil {
		t.Fatal("DownloadReleaseAsset failed:", err)
	}

	var buf bytes.Buffer
	if len(url) != 0 {
		res, err := http.Get(url)
		if err != nil {
			t.Fatal("http.Get failed:", err)
		}

		if _, err := io.Copy(&buf, res.Body); err != nil {
			t.Fatal("Copy failed:", err)
		}
		res.Body.Close()

	} else {
		if _, err := io.Copy(&buf, rc); err != nil {
			t.Fatal("Copy failed:", err)
		}
		rc.Close()

	}

	if got, want := buf.String(), "darwin_386\n"; got != want {
		t.Fatalf("file body is %q, want %q", got, want)
	}
}

func TestGitHubClient_ListAssets(t *testing.T) {
	client := testGithubClient(t)
	testTag := "github-list-assets"
	req := &github.RepositoryRelease{
		TagName: github.String(testTag),
		Draft:   github.Bool(true),
	}

	release, err := client.CreateRelease(context.TODO(), req)
	if err != nil {
		t.Fatal("CreateRelease failed:", err)
	}

	defer func() {
		if err := client.DeleteRelease(context.TODO(), *release.ID); err != nil {
			t.Fatalf("DeleteRelease failed: %s", err)
		}
	}()

	for _, filename := range []string{"darwin_386", "darwin_amd64"} {
		filename := filepath.Join("./testdata", filename)
		if _, err := client.UploadAsset(context.TODO(), *release.ID, filename); err != nil {
			t.Fatal("UploadAsset failed:", err)
		}
	}

	assets, err := client.ListAssets(context.TODO(), *release.ID)
	if err != nil {
		t.Fatal("ListAssets failed:", err)
	}

	if got, want := len(assets), 2; got != want {
		t.Fatalf("ListAssets number = %d, want %d", got, want)
	}

	if err := client.DeleteAsset(context.TODO(), *assets[0].ID); err != nil {
		t.Fatal("DeleteAsset failed:", err)
	}

	assets, err = client.ListAssets(context.TODO(), *release.ID)
	if err != nil {
		t.Fatal("ListAssets failed:", err)
	}

	if got, want := len(assets), 1; got != want {
		t.Fatalf("ListAssets number = %d, want %d", got, want)
	}
}
