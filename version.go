package main

import (
	"os"
	"strconv"
	"time"

	"github.com/tcnksm/go-latest"
)

// Name is application name
const Name = "ghr"

// Version is application version
const Version string = "v0.5.0"

// GitCommit describes latest commit hash.
// This is automatically extracted by git describe --always.
var GitCommit string

// verCheckCh is channel which gets go-latest.Response
var verCheckCh = make(chan *latest.CheckResponse)

// CheckTimeout is timeout of go-latest.Check executiom
var CheckTimeout time.Duration

// defaultCheckTimeout is default timeout of go-latest.Check
// execution.
var defaultCheckTimeout = 2 * time.Second

// envCheckTimeout is environmental varible to
// set go-latest.Check execution timeout.
const envCheckTimeout = "GHR_CHECK_WAIT"

func init() {

	CheckTimeout = defaultCheckTimeout
	if timeStr := os.Getenv(envCheckTimeout); timeStr != "" {
		t, err := strconv.Atoi(timeStr)
		// If wait to conv, ignore env value
		if err == nil {
			CheckTimeout = time.Duration(t)
		}
	}

	go func() {
		fixFunc := latest.DeleteFrontV()
		githubTag := &latest.GithubTag{
			Owner:             "tcnksm",
			Repository:        "ghr",
			FixVersionStrFunc: fixFunc,
		}

		// Ignore error, because it's not important for ghr fucntion
		res, _ := latest.Check(githubTag, fixFunc(Version))
		verCheckCh <- res
	}()

}
