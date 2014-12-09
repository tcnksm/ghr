package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	flag "github.com/dotcloud/docker/pkg/mflag"
	"path/filepath"
)

var (
	owner        = flag.String([]string{"u", "-username"}, "", "GitHub username")
	repo         = flag.String([]string{"r", "-repository"}, "", "Repository name")
	token        = flag.String([]string{"t", "-token"}, "", "Github API Token")
	parallel     = flag.Int([]string{"p", "--parallel"}, -1, "Parallelization factor")
	flReplace    = flag.Bool([]string{"-replace"}, false, "Replace asset if target is already uploaded")
	flDelete     = flag.Bool([]string{"-delete"}, false, "Delete release if it exists")
	flDraft      = flag.Bool([]string{"-draft"}, false, "Create unpublised release")
	flPrerelease = flag.Bool([]string{"-prerelease"}, false, "Create prerelease")
	flVersion    = flag.Bool([]string{"v", "-version"}, false, "Print version information and quit")
	flHelp       = flag.Bool([]string{"h", "-help"}, false, "Print this message and quit")
	flDebug      = flag.Bool([]string{"-debug"}, false, "Run as DEBUG mode")
)

func main() {
	// call ghrMain in a separate function
	// so that it can use defer and have them
	// run before the exit.
	os.Exit(ghrMain())
}

func ghrMain() int {
	var err error

	flag.Parse()

	if *flDebug {
		os.Setenv("DEBUG", "1")
	}
	debug("Run as DEBUG mode")
	debug("Version:", Version)
	debug("Execution:", os.Args)

	if *flHelp {
		showHelp()
		return 0
	}

	if *flVersion {
		showVersion()
		return 0
	}

	if len(flag.Args()) != 2 {
		showHelp()
		return 1
	}

	tag := flag.Arg(0)
	inputPath := flag.Arg(1)

	// Info stores all configuration values
	// for using GitHub API.
	info, err := NewInfo(tag)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return 1
	}

	files, err := Artifacts(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return 1
	}

	// Create release on GitHub if not exsits.
	// If exists, just set its releaseID to *Info
	// If `--delete`, deleting it if exists.
	err = SetRelease(info, *flDelete)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return 1
	}

	// If `--replace`, deleting artifacts if exists (Not here)
	// Here just extracting all artifacts which are already on
	// GitHub and related to releaseID.
	var deleteTargets []DeleteTarget
	if *flReplace {
		deleteTargets, err = GetDeleteTargets(info, ArtifactNames(files))
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
		}
	}

	// Limit amount of parallelism by number of logic CPU
	if *parallel <= 0 {
		*parallel = runtime.NumCPU()
	}

	// Use CPU efficiently
	cpu := runtime.NumCPU()
	runtime.GOMAXPROCS(cpu)

	var errorLock sync.Mutex
	var wg sync.WaitGroup

	errors := make([]string, 0)
	semaphore := make(chan int, *parallel)

	for _, path := range files {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			f, _ := os.Stat(path)
			if f.IsDir() {
				fmt.Fprintf(os.Stderr, "%s is directory, skip it\n", path)
				return
			}

			semaphore <- 1

			// Only when `--replace`,
			// Deleting artifact on GitHub if exists in advance
			if id := DeleteTargetID(deleteTargets, path); id != ID_NOT_FOUND {
				fmt.Fprintf(os.Stderr, "--> Deleting: %15s\n", filepath.Base(path))
				if err := DeleteAsset(info, id); err != nil {
					errorLock.Lock()
					defer errorLock.Unlock()
					errors = append(errors,
						fmt.Sprintf("deleting %s error: %s", filepath.Base(path), err))
				}
			}

			// Upload artifacts
			fmt.Fprintf(os.Stderr, "--> Uploading: %15s\n", path)
			if err := UploadAsset(info, path); err != nil {
				errorLock.Lock()
				defer errorLock.Unlock()
				errors = append(errors,
					fmt.Sprintf("upload %s error: %s", path, err))
			}

			<-semaphore
		}(path)
	}
	wg.Wait()

	// List all errors in uploading or deleting
	if len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "%d errors occurred:\n", len(errors))
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "--> %s\n", err)
		}
		return 1
	}

	return 0
}

func debug(v ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		log.Println(v...)
	}
}

func showVersion() {
	fmt.Fprintf(os.Stderr, "ghr version %s, build %s \n", Version, GitCommit)
}

func showHelp() {
	fmt.Fprintf(os.Stderr, helpText)
}

const helpText = `Usage: ghr [option] <tag> <artifacts>

ghr - easy to release to Github in parallel

Options:

  -u, --username     Github username
  -t, --token        Github API Token
  -r, --repository   Github repository name
  -p, --parallel=-1  Amount of parallelism, defaults to number of CPUs
　--replace          Replace asset if target already exists
　--delete           Delete release and its git tag if same version exists
  --draft            Create unpublised release
  --prerelease       Create prerelease
  -h, --help         Print this message and quit
  -v, --version      Print version information and quit
  --debug=false      Run as DEBUG mode

Example:
  $ ghr v1.0.0 dist/
  $ ghr --replace v1.0.0 dist/
  $ ghr v1.0.2 dist/tool.zip
`
