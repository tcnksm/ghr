package main

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/google/go-github/github"
)

func TestGHR_CreateRelease(t *testing.T) {
	t.Parallel()

	githubClient := testGithubClient(t)
	GHR := &GHR{
		GitHub:    githubClient,
		outStream: ioutil.Discard,
	}

	testTag := "create-release"
	req := &github.RepositoryRelease{
		TagName: &testTag,
		Draft:   github.Bool(false),
		Body:    github.String("This is test release"),
	}

	recreate := false
	release, err := GHR.CreateRelease(context.TODO(), req, recreate)
	if err != nil {
		t.Fatal("CreateRelease failed:", err)
	}

	defer GHR.DeleteRelease(context.TODO(), *release.ID, testTag)
}

func TestGHR_CreateReleaseWithExistingRelease(t *testing.T) {
	t.Parallel()

	githubClient := testGithubClient(t)
	GHR := &GHR{
		GitHub:    githubClient,
		outStream: ioutil.Discard,
	}

	testTag := "create-with-existing"
	existingReq := &github.RepositoryRelease{
		TagName: github.String(testTag),
		Draft:   github.Bool(false),
	}
	cases := []struct {
		request    *github.RepositoryRelease
		recreate   bool
		newRelease bool
	}{
		// 0: When same tag as existing release is used
		{
			&github.RepositoryRelease{
				TagName: github.String(testTag),
				Draft:   github.Bool(false),
			},
			false,
			false,
		},

		// 1: When draft release is requested
		{
			&github.RepositoryRelease{
				TagName: github.String(testTag),
				Draft:   github.Bool(true),
			},
			false,
			true,
		},

		// 2: When recreate is requtested
		{
			&github.RepositoryRelease{
				TagName: github.String(testTag),
				Draft:   github.Bool(false),
			},
			true,
			true,
		},

		// 3: When different tag is requtested
		{
			&github.RepositoryRelease{
				TagName: github.String("v2.0.0"),
				Draft:   github.Bool(false),
			},
			false,
			true,
		},
	}

	for i, tc := range cases {
		// Prevent a lot of requests at same time to GitHub API
		time.Sleep(1 * time.Second)

		// Create an existing release before
		existing, err := githubClient.CreateRelease(context.TODO(), existingReq)
		if err != nil {
			t.Fatalf("#%d CreateRelease failed: %s", i, err)
		}

		// Create a release for THIS TEST
		created, err := GHR.CreateRelease(context.TODO(), tc.request, tc.recreate)
		if err != nil {
			t.Fatalf("#%d GHR.CreateRelease failed: %s", i, err)
		}

		// Clean up existing release
		if !tc.recreate {
			err = GHR.DeleteRelease(context.TODO(), *existing.ID, *existingReq.TagName)
			if err != nil {
				t.Fatalf("#%d GHR.DeleteRelease (existing) failed: %s", i, err)
			}
		}

		if !tc.newRelease {
			if *created.ID != *existing.ID {
				t.Fatalf("#%d created ID %d, want %d (same as existing release ID)",
					i, *created.ID, *existing.ID)
			}
			continue
		}

		// Clean up newly created release before. When draft request,
		// tag is not created. So it need to be deleted separately by Github client.
		if *tc.request.Draft {
			// Clean up newly created release before checking
			if err := githubClient.DeleteRelease(context.TODO(), *created.ID); err != nil {
				t.Fatalf("#%d GitHub.DeleteRelease (created) failed: %s", i, err)
			}
		} else {
			err := GHR.DeleteRelease(context.TODO(), *created.ID, *tc.request.TagName)
			if err != nil {
				t.Fatalf("#%d GHR.DeleteRelease (created) failed: %s", i, err)
			}
		}

		if *created.ID == *existing.ID {
			t.Fatalf("#%d expect created ID %d to be different from existing ID %d",
				i, *created.ID, *existing.ID)
		}
	}
}

func TestGHR_UploadAssets(t *testing.T) {
	githubClient := testGithubClient(t)
	GHR := &GHR{
		GitHub:    githubClient,
		outStream: ioutil.Discard,
	}

	testTag := "ghr-upload-assets"
	req := &github.RepositoryRelease{
		TagName: github.String(testTag),
		Draft:   github.Bool(true),
	}

	// Create an existing release before
	release, err := githubClient.CreateRelease(context.TODO(), req)
	if err != nil {
		t.Fatalf("CreateRelease failed: %s", err)
	}

	defer func() {
		if err := githubClient.DeleteRelease(context.TODO(), *release.ID); err != nil {
			t.Fatal("DeleteRelease failed:", err)
		}
	}()

	localTestAssets, err := LocalAssets(TestDir)
	if err != nil {
		t.Fatal("LocalAssets failed:", err)
	}

	if err := GHR.UploadAssets(context.TODO(), *release.ID, localTestAssets, 4); err != nil {
		t.Fatal("GHR.UploadAssets failed:", err)
	}

	assets, err := githubClient.ListAssets(context.TODO(), *release.ID)
	if err != nil {
		t.Fatal("ListAssets failed:", err)
	}

	if got, want := len(assets), 4; got != want {
		t.Fatalf("upload assets number = %d, want %d", got, want)
	}

	// Delete all assets
	parallel := 4
	if err := GHR.DeleteAssets(context.TODO(), *release.ID, localTestAssets, parallel); err != nil {
		t.Fatal("GHR.DeleteAssets failed:", err)
	}

	assets, err = githubClient.ListAssets(context.TODO(), *release.ID)
	if err != nil {
		t.Fatal("ListAssets failed:", err)
	}

	if got, want := len(assets), 0; got != want {
		t.Fatalf("upload assets number = %d, want %d", got, want)
	}
}
