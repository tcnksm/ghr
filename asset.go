package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/go-github/github"
)

const (
	// AssetIDNotFound is default value of `Asset.ID`.
	AssetIDNotFound int = 0
)

// Asset is the uploading target object.
type Asset struct {

	// Path is artifact local path.
	Path string

	// Name is artifact base name.
	Name string

	// ID is assets ID on GitHub.
	// We use this for checking asset is already uploaded or not.
	ID int
}

// UploadAssets uploads assets parallelly.
func UploadAssets(assets []*Asset, ghrOpts *GhrOpts, apiOpts *GitHubAPIOpts) (errors []string) {

	var errorLock sync.Mutex
	var wg sync.WaitGroup

	errors = make([]string, 0)
	semaphore := make(chan int, ghrOpts.Parallel)
	for _, asset := range assets {
		wg.Add(1)
		go func(asset *Asset) {
			defer wg.Done()
			semaphore <- 1

			// If `--replace` flag is set and asset is already
			// uploaded on Github. ghr delete it in advance.
			if ghrOpts.Replace && asset.ID != AssetIDNotFound {
				fmt.Fprintf(os.Stderr, "--> Deleting: %15s\n", asset.Name)
				if err := DeleteAsset(asset, apiOpts); err != nil {
					errorLock.Lock()
					defer errorLock.Unlock()
					errors = append(errors,
						fmt.Sprintf("delete %s error: %s", asset.Name, err))
				}
			}

			// Upload asset.
			fmt.Fprintf(os.Stderr, "--> Uploading: %15s\n", asset.Name)
			if err := UploadAsset(asset, apiOpts); err != nil {
				errorLock.Lock()
				defer errorLock.Unlock()
				errors = append(errors,
					fmt.Sprintf("upload %s error: %s", asset.Name, err))
			}

			<-semaphore
		}(asset)
	}
	wg.Wait()

	return errors
}

// UploadAsset upload asset to GitHub release
func UploadAsset(asset *Asset, apiOpts *GitHubAPIOpts) (err error) {
	// Create client
	client := NewOAuthedClient(apiOpts)

	// Set upload URL.
	// Upload URL is provided when Creating release.
	client.UploadURL = ExtractUploadURL(apiOpts)

	// OpenFile
	file, err := os.Open(asset.Path)
	if err != nil {
		return fmt.Errorf("failed to open file: %s\n", asset.Name)
	}

	// Set asset Name
	opts := &github.UploadOptions{Name: asset.Name}

	// Release Asset
	_, res, err := client.Repositories.UploadReleaseAsset(apiOpts.OwnerName, apiOpts.RepoName, apiOpts.ID, opts, file)
	if err != nil {
		return err
	}

	err = CheckStatusCreated(res)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAsset deletes asset on Github
func DeleteAsset(asset *Asset, apiOpts *GitHubAPIOpts) (err error) {
	// Create client
	client := NewOAuthedClient(apiOpts)

	// Delete asset on GitHub
	res, err := client.Repositories.DeleteReleaseAsset(apiOpts.OwnerName, apiOpts.RepoName, asset.ID)
	if err != nil {
		return err
	}

	err = CheckStatusNoContent(res)
	if err != nil {
		return err
	}

	return nil
}

// FetchAssetID fetches all assets which are already uploaded on Github.
func FetchAssetID(assets []*Asset, apiOpts *GitHubAPIOpts) error {
	// Create client
	client := NewOAuthedClient(apiOpts)

	// Get all assets on Github related to its relase ID
	releasedAssets, res, err := client.Repositories.ListReleaseAssets(apiOpts.OwnerName, apiOpts.RepoName, apiOpts.ID, nil)
	if err != nil {
		return err
	}

	err = CheckStatusOK(res)
	if err != nil {
		return err
	}

	// Check asset which is already on GitHub
	for _, ra := range releasedAssets {
		for _, a := range assets {
			if *ra.Name == a.Name {
				a.ID = *ra.ID
			}
		}
	}

	return nil
}

// GetLocalAssets extract local files to upload.
// If path is directory, ghr uploads all files there.
// If path is file, ghr only upload it.
func GetLocalAssets(path string) ([]*Asset, error) {
	var assets []*Asset

	// Get file Info
	fInfo, err := os.Stat(path)
	if err != nil {
		return assets, err
	}

	// If path is a sigle file just return it.
	if !fInfo.IsDir() {
		return append(assets, &Asset{Path: path, Name: filepath.Base(path)}), nil
	}

	// Glob all files in path
	files, err := filepath.Glob(path + "/*")
	if err != nil {
		return assets, err
	}

	for _, f := range files {

		// Don't include directory
		if f, _ := os.Stat(f); f.IsDir() {
			continue
		}

		assets = append(assets, &Asset{Path: f, Name: filepath.Base(f)})
	}

	return assets, nil
}
