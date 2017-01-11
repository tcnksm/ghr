package main

import (
	"bytes"
	"fmt"
	"time"

	latest "github.com/tcnksm/go-latest"
)

// Name is application name
const Name = "ghr"

// Version is application version
const Version string = "v0.5.4"

// GitCommit describes latest commit hash.
// This is automatically extracted by git describe --always.
var GitCommit string

func OutputVersion() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s version %s", Name, Version)
	if len(GitCommit) != 0 {
		fmt.Fprintf(&buf, " (%s)", GitCommit)
	}
	fmt.Fprintf(&buf, "\n")

	// Check latest version is release or not.
	verCheckCh := make(chan *latest.CheckResponse)
	go func() {
		fixFunc := latest.DeleteFrontV()
		githubTag := &latest.GithubTag{
			Owner:             "tcnksm",
			Repository:        "ghr",
			FixVersionStrFunc: fixFunc,
		}

		res, err := latest.Check(githubTag, fixFunc(Version))
		if err != nil {
			// Don't return error
			Debugf("[ERROR] Check lastet version is failed: %s", err)
			return
		}
		verCheckCh <- res
	}()

	select {
	case <-time.After(defaultCheckTimeout):
	case res := <-verCheckCh:
		if res.Outdated {
			fmt.Fprintf(&buf,
				"Latest version of ghr is v%s, please upgrade!\n",
				res.Current)
		}
	}

	return buf.String()
}
