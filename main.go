package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

func debug(v ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		log.Println(v...)
	}
}

func artifacts(path string) ([]string, error) {
	var files []string
	file, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if file.IsDir() {
		files, err = filepath.Glob(path + "/*")
		if err != nil {
			return nil, err
		}
	} else {
		files = append(files, path)
	}

	return files, nil
}

func main() {
	// call ghrMain in a separate function
	// so that it can use defer and have them
	// run before the exit.
	os.Exit(ghrMain())
}

func ghrMain() int {

	if os.Getenv("GITHUB_TOKEN") == "" {
		fmt.Fprintf(os.Stderr, "Please set your Github API Token in the GITHUB_TOKEN env var\n")
		return 1
	}

	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: ghr <tag> <artifact>\n")
		return 1
	}

	tag := os.Args[1]
	inputPath := os.Args[2]

	owner, err := GetOwnerName()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return 1
	}

	repo, err := GetRepoName()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return 1
	}

	info := NewInfo()
	info.Token = os.Getenv("GITHUB_TOKEN")
	info.TagName = tag
	info.OwnerName = owner
	info.RepoName = repo
	debug(info)

	id, err := GetReleaseID(info)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return 1
	}

	if id == -1 {
		id, err = CreateNewRelease(info)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return 1
		}

		if id == -1 {
			fmt.Fprintf(os.Stderr, "Counld not retrieve release ID\n")
		}
	}

	info.ID = id
	debug(id)

	files, err := artifacts(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return 1
	}

	var errorLock sync.Mutex
	var wg sync.WaitGroup
	errors := make([]string, 0)

	for _, path := range files {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			f, _ := os.Stat(path)
			if f.IsDir() {
				fmt.Fprintf(os.Stderr, "%s is directory, skip it\n", path)
				return
			}

			fmt.Printf("--> Uploading: %15s\n", path)
			if err := UploadAsset(info, path); err != nil {
				errorLock.Lock()
				defer errorLock.Unlock()
				errors = append(errors,
					fmt.Sprintf("%s error: %s", path, err))
			}
		}(path)
	}
	wg.Wait()

	if len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "%d errors occurred:\n", len(errors))
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "--> %s\n", err)
		}
		return 1
	}

	return 0
}
