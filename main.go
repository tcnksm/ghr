package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	flag "github.com/dotcloud/docker/pkg/mflag"
	"runtime"
)

var (
	owner        = flag.String([]string{"u", "-username"}, "", "GitHub username")
	repo         = flag.String([]string{"r", "-repository"}, "", "Repository name")
	token        = flag.String([]string{"t", "-token"}, "", "Github API Token")
	parallel     = flag.Int([]string{"p", "--parallel"}, -1, "Parallelization factor")
	flDraft      = flag.Bool([]string{"-draft"}, false, "Create unpublised release")
	flPrerelease = flag.Bool([]string{"-prerelease"}, false, "Create prerelease")
	flVersion    = flag.Bool([]string{"v", "-version"}, false, "Print version information and quit")
	flHelp       = flag.Bool([]string{"h", "-help"}, false, "Print this message and quit")
	flDebug      = flag.Bool([]string{"-debug"}, false, "Run as DEBUG mode")
)

func debug(v ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		log.Println(v...)
	}
}

func showVersion() {
	fmt.Fprintf(os.Stderr, "ghr %s\n", Version)
}

func showHelp() {
	fmt.Fprintf(os.Stderr, helpText)
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

	flag.Parse()

	if *flHelp {
		showHelp()
		return 0
	}

	if *flVersion {
		showVersion()
		return 0
	}

	if *flDebug {
		os.Setenv("DEBUG", "1")
		debug("Run as DEBUG mode")
	}

	if len(flag.Args()) != 2 {
		showHelp()
		return 1
	}

	tag := flag.Arg(0)
	inputPath := flag.Arg(1)

	var err error

	if *owner == "" {
		*owner, err = GetOwnerName()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return 1
		}
	}

	if *repo == "" {
		*repo, err = GetRepoName()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return 1
		}
	}

	if *token == "" {
		*token = os.Getenv("GITHUB_TOKEN")
		if *token == "" {
			fmt.Fprintf(os.Stderr, "Please set your Github API Token in the GITHUB_TOKEN env var\n")
			return 1
		}
	}

	// Limit amount of parallelism
	// by number of logic CPU
	if *parallel <= 0 {
		*parallel = runtime.NumCPU()
	}

	info := Info{
		TagName:         tag,
		Token:           *token,
		OwnerName:       *owner,
		RepoName:        *repo,
		TargetCommitish: "master",
		Draft:           *flDraft,
		Prerelease:      *flPrerelease,
	}
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
			fmt.Printf("--> Uploading: %15s\n", path)
			if err := UploadAsset(info, path); err != nil {
				errorLock.Lock()
				defer errorLock.Unlock()
				errors = append(errors,
					fmt.Sprintf("%s error: %s", path, err))
			}
			<-semaphore
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

const helpText = `Usage: ghr [option] <tag> <artifacts>

ghr - easy to release to Github in parallel

Options:

  -u, --username     Github username
  -t, --token        Github API Token
  -r, --repository   Github repository name
  -p, --parallel=-1  Amount of parallelism, defaults to number of CPUs
  --draft            Create unpublised release
  --prerelease       Create prerelease	
  -h, --help         Print this message and quit
  -v, --version      Print version information and quit
  --debug=false      Run as DEBUG mode

Example:
  $ ghr v1.0.0 pkg/dist/
  $ ghr v1.0.2 pkg/dist/tool.zip
`
