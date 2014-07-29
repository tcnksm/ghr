package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Info struct {
	ID              int
	Token           string
	TagName         string
	RepoName        string
	OwnerName       string
	TargetCommitish string
	Draft           bool
	Prerelease      bool
}

func debug(v ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		log.Println(v...)
	}
}

func NewInfo() Info {
	return Info{
		TargetCommitish: "master",
		Draft:           false,
		Prerelease:      false,
	}
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
	debug(tag)

	path := os.Args[2]
	debug(path)

	owner, err := GetOwnerName()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return 1
	}
	debug(owner)

	repo, err := GetRepoName()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return 1
	}
	debug(repo)

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
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
			return 1
		}

		if id == -1 {
			fmt.Fprintf(os.Stderr, "Counld not retrieve release ID\n")
		}
	}
	debug(id)
	info.ID = id

	file, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return 1
	}

	var files []string

	if file.IsDir() {
		files, err = filepath.Glob(path + "/*")
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return 1
		}
	} else {
		files = append(files, path)
	}
	debug(files)

	var errorLock sync.Mutex
	var wg sync.WaitGroup
	errors := make([]string, 0)

	for _, p := range files {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			fmt.Printf("--> Uploading: %15s\n", p)
			// TODO check IsDir
			if err := UploadAsset(info, p); err != nil {
				errorLock.Lock()
				defer errorLock.Unlock()
				errors = append(errors,
					fmt.Sprintf("%s error: %s", p, err))
			}
		}(p)
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
