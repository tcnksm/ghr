package main

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type GHR struct {
	GitHub GitHub

	outStream io.Writer
}

func (g *GHR) CreateRelease(ctx context.Context, req *github.RepositoryRelease, recreate bool) (*github.RepositoryRelease, error) {

	// When draft release creation is requested,
	// create it witout any check (it can).
	if *req.Draft {
		fmt.Fprintln(g.outStream, "==> Create a draft release")
		return g.GitHub.CreateRelease(ctx, req)
	}

	// Check release is exist or not.
	// If release is not found, then create a new release.
	release, err := g.GitHub.GetRelease(ctx, *req.TagName)
	if err != nil {
		if err != RelaseNotFound {
			return nil, errors.Wrap(err, "failed to get release")
		}
		Debugf("Release (with tag %s) is not found: create a new one",
			*req.TagName)

		if recreate {
			fmt.Fprintf(g.outStream,
				"WARNING: '-recreate' is specified but release (%s) is not found",
				*req.TagName)
		}

		fmt.Fprintln(g.outStream, "==> Create a new release")
		return g.GitHub.CreateRelease(ctx, req)
	}

	// recreae is not true. Then use that exiting release.
	if !recreate {
		Debugf("Release (with tag %s) exists: use exsiting one",
			*req.TagName)

		fmt.Fprintf(g.outStream, "WARNING: found release (%s). Use existing one.",
			*req.TagName)
		return release, nil
	}

	// When recreate is requested, delete exsiting release
	// and create a new release.
	fmt.Fprintln(g.outStream, "==> Recreate a release")
	if err := g.DeleteRelease(ctx, *release.ID, *req.TagName); err != nil {
		return nil, err
	}

	return g.GitHub.CreateRelease(ctx, req)
}

func (g *GHR) DeleteRelease(ctx context.Context, ID int, tag string) error {

	err := g.GitHub.DeleteRelease(ctx, ID)
	if err != nil {
		return err
	}

	err = g.GitHub.DeleteTag(ctx, tag)
	if err != nil {
		return err
	}

	return nil
}

func (g *GHR) UploadAssets(ctx context.Context, releaseID int, localAssets []string, parallel int) error {
	start := time.Now()
	defer func() {
		Debugf("UploadAssets: time: %d ms", int(time.Since(start).Seconds()*1000))
	}()

	eg, ctx := errgroup.WithContext(ctx)
	semaphore := make(chan struct{}, parallel)
	for _, localAsset := range localAssets {
		localAsset := localAsset
		eg.Go(func() error {
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()

			fmt.Fprintf(g.outStream, "--> Uploading: %15s\n", filepath.Base(localAsset))
			_, err := g.GitHub.UploadAsset(ctx, releaseID, localAsset)
			if err != nil {
				return errors.Wrapf(err,
					"failed to upload asset: %s", localAsset)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "one of goroutines is failed")
	}

	return nil
}

func (g *GHR) DeleteAssets(ctx context.Context, releaseID int, localAssets []string, parallel int) error {
	start := time.Now()
	defer func() {
		Debugf("DeleteAssets: time: %d ms", int(time.Since(start).Seconds()*1000))
	}()

	eg, ctx := errgroup.WithContext(ctx)

	assets, err := g.GitHub.ListAssets(ctx, releaseID)
	if err != nil {
		return errors.Wrap(err, "failed to list assets")
	}

	semaphore := make(chan struct{}, parallel)
	for _, localAsset := range localAssets {
		for _, asset := range assets {
			// https://golang.org/doc/faq#closures_and_goroutines
			localAsset, asset := localAsset, asset

			// Uploaded asset name is same as basename of local file
			if *asset.Name == filepath.Base(localAsset) {
				eg.Go(func() error {
					semaphore <- struct{}{}
					defer func() {
						<-semaphore
					}()

					fmt.Fprintf(g.outStream, "--> Deleting: %15s\n", *asset.Name)
					if err := g.GitHub.DeleteAsset(ctx, *asset.ID); err != nil {
						return errors.Wrapf(err,
							"failed to delete asset: %s", *asset.Name)
					}
					return nil
				})
			}
		}
	}

	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "one of goroutines is failed")
	}

	return nil
}
