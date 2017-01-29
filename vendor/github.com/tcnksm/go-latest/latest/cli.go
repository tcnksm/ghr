package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tcnksm/go-latest"
)

type CLI struct {
	// out/err stream is the stdout and stderr
	// to write message from CLI
	outStream, errStream io.Writer
}

// Run executes CLI and return its exit code
func (c *CLI) Run(args []string) int {
	var githubTag latest.GithubTag

	flags := flag.NewFlagSet(Name, flag.ExitOnError)
	flags.Usage = func() { fmt.Fprintf(c.errStream, helpText) }
	flags.SetOutput(c.errStream)

	flags.StringVar(&githubTag.Repository,
		"repo", "", "Repository name")
	flags.StringVar(&githubTag.Owner,
		"owner", "", "Repository owner name")

	flgNew := flags.Bool("new",
		false, "Check TAG(VERSION) is new and greater")
	flgFixVerStrFunc := flags.String("fix",
		"none", "Specify FixVersionStrFunc")
	flgVersion := flags.Bool("version",
		false, "Print version information")
	flgHelp := flags.Bool("help",
		false, "Print this message and quit")
	flgDebug := flags.Bool("debug",
		false, "Print verbose(debug) output")

	if err := flags.Parse(args[1:]); err != nil {
		fmt.Fprint(c.errStream, "Failed to parse flag\n")
		return 1
	}

	// Show version and quit
	if *flgVersion {
		fmt.Fprintf(c.errStream, "%s Version v%s build %s\n", Name, Version, GitCommit)
		return 0
	}

	// Show help and quit
	if *flgHelp {
		fmt.Fprintf(c.errStream, helpText)
		return 0
	}

	// Run as debug mode
	if os.Getenv(envDebug) != "" {
		*flgDebug = true
	}

	parsedArgs := flags.Args()
	if len(parsedArgs) != 1 {
		fmt.Fprintf(c.errStream, "Invalid arguments\n")
		return 1
	}
	target := parsedArgs[0]

	// Specify FixVersionStrFunc
	// e.g., if version is v0.3.1 it should be 0.3.1 (SemVer format)
	var f latest.FixVersionStrFunc
	switch *flgFixVerStrFunc {
	case "none":
		f = nil
	case "frontv":
		f = latest.DeleteFrontV()
		target = f(target)
	default:
		fmt.Fprintf(c.errStream, "Invalid fix func: %s\n", *flgFixVerStrFunc)
		return 1
	}

	githubTag.FixVersionStrFunc = f
	res, err := latest.Check(&githubTag, target)
	if err != nil {
		fmt.Fprintf(c.errStream, "Failed to check: %s\n", err.Error())
		return 1
	}

	// Default variables
	exitCode := 0
	output := fmt.Sprintf("%s is latest\n", target)

	// Check version is `new`
	if *flgNew {
		if !res.New {
			exitCode = 1
			output = fmt.Sprintf("%s is not new\n", target)
		} else {
			output = fmt.Sprintf("%s is new\n", target)
		}
	} else {
		if !res.Latest {
			exitCode = 1
			output = fmt.Sprintf("%s is not latest\n", target)
		}
	}

	if *flgDebug {
		fmt.Fprint(c.outStream, output)
	}

	return exitCode
}

const helpText = `Usage: latest [options] TAG

    latest command check TAG(VERSION) is latest. If is not latest,
    it returns non-zero value. It try to compare version by Semantic
    Versioning. 

Options:

    -owner=NAME    Set GitHub repository owner name.

    -repo=NAME     Set Github repository name.

    -new           Check TAG(VERSION) is new. 'new' means TAG(VERSION)
                   is not exist and greater than others.

    -fix=none      Specify FixVersionStrFunc (Fix version string to SemVer)
                   'none': does nothing (default)
                   'front': deletes front 'v' charactor

    -help          Print this message and quit.

    -debug         Print verbose(debug) output.

Example:

    $ latest -debug 0.2.0
`
