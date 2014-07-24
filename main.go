package main

import (
	"fmt"
	"log"
	"os"
)

func debug(v ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		log.Println(v...)
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
		fmt.Fprintf(os.Stderr, "Usage: ghr <version> <artifact>\n")
		return 1
	}

	version := os.Args[1]
	debug(version)

	artifacts := os.Args[2]
	debug(artifacts)

	owner, _ := GitOwner()
	debug(owner)

	remoteURL, _ := GitRemote()
	debug(remoteURL)

	repo := GitRepoName(remoteURL)
	debug(repo)

	// git config --local remote.origin.url
	// https://github.com/tcnksm/ghr.git

	// git config --global user.name
	// tcnksm

	fmt.Fprintf(os.Stderr, "Success\n")
	return 0
}
