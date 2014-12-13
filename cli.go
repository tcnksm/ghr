package main

import (
	"fmt"
	"io"
	"os"
	"runtime"

	flag "github.com/dotcloud/docker/pkg/mflag"
	"github.com/tcnksm/go-gitconfig"
)

// Exit codes are in value that represnet an exit code for a paticular error.
const (
	ExitCodeOK int = 0

	// Errors start at 10
	ExitCodeError = 10 + iota
	ExitCodeParseFlagsError
	ExitCodeBadArgs
	ExitCodeTokenNotFound
	ExitCodeOwnerNotFound
	ExitCodeRepoNotFound
	ExitCodeRleaseError
)

// GhrOpts are the options for ghr related
type GhrOpts struct {
	// Parallel determines number of amount of parallelism.
	// Default is number of CPU.
	Parallel int

	// Replace determines repalce asset on GitHub if it exists.
	Replace bool

	// Detele determines delete release if it exists.
	Delete bool
}

// CLI is the command line object
type CLI struct {
	// outStream and errStream are the stdout and stderr
	// to write message from the CLI.
	outStream, errStream io.Writer
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var (
		// GitHub API related.
		githubAPIOpts GitHubAPIOpts

		// ghr related.
		ghrOpts GhrOpts

		// general
		version bool
		help    bool
		debug   bool

		err error
	)

	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)

	// Receive options for GitHub API.
	flags.StringVar(&githubAPIOpts.OwnerName, []string{"u", "-username"}, "",
		"GitHub username")
	flags.StringVar(&githubAPIOpts.RepoName, []string{"r", "-repository"}, "",
		"Repository name")
	flags.StringVar(&githubAPIOpts.Token, []string{"t", "-token"}, "",
		"GitHub API Token")
	flags.BoolVar(&githubAPIOpts.Draft, []string{"-draft"}, false,
		"Create unpublised release")
	flags.BoolVar(&githubAPIOpts.Prerelease, []string{"-prerelease"}, false,
		"Create prerelease")

	// Receive options to change ghr work.
	flags.IntVar(&ghrOpts.Parallel, []string{"p", "--parallel"}, -1,
		"Parallelization factor")
	flags.BoolVar(&ghrOpts.Replace, []string{"-replace"}, false,
		"Replace asset if target is already uploaded")
	flags.BoolVar(&ghrOpts.Delete, []string{"-delete"}, false,
		"Delete release if it exists")

	// Receive general options.
	flags.BoolVar(&version, []string{"v", "-version"}, false,
		"Print version information and quit")
	flags.BoolVar(&help, []string{"h", "-help"}, false,
		"Print this message and quit")
	flags.BoolVar(&debug, []string{"-debug"}, false,
		"Run as DEBUG mode")

	// Parse all the flags
	if err := flags.Parse(args[1:]); err != nil {
		return ExitCodeParseFlagsError
	}

	// Version
	if version {
		fmt.Fprintf(cli.errStream, "ghr version %s, build %s \n", Version, GitCommit)
		return ExitCodeOK
	}

	// Help
	if help {
		fmt.Fprintf(cli.errStream, helpText)
		return ExitCodeOK
	}

	// Run as DEBUG mode
	if debug {
		os.Setenv("DEBUG", "1")
	}

	// Get the parsed arguments
	parsedArgs := flags.Args()
	if len(parsedArgs) != 2 {
		fmt.Fprintf(cli.errStream, "must specify two arguments - tag, path\n")
		fmt.Fprintf(cli.errStream, helpText)
		return ExitCodeBadArgs
	}

	// Get the tag of release and path
	tag, path := parsedArgs[0], parsedArgs[1]
	githubAPIOpts.TagName = tag

	// Set Token
	err = setToken(&githubAPIOpts)
	if err != nil {
		fmt.Fprintf(cli.errStream, err.Error())
		return ExitCodeTokenNotFound
	}

	// Set repository owner name.
	err = setOwner(&githubAPIOpts)
	if err != nil {
		fmt.Fprintf(cli.errStream, err.Error())
		return ExitCodeOwnerNotFound
	}

	// Set repository owner name.
	err = setRepo(&githubAPIOpts)
	if err != nil {
		fmt.Fprintf(cli.errStream, err.Error())
		return ExitCodeRepoNotFound
	}

	// Get the asset to upload.
	assets, err := GetLocalAssets(path)
	if err != nil {
		fmt.Fprintf(cli.errStream, err.Error())
		return ExitCodeError
	}

	// Create release.
	err = CreateRelease(&ghrOpts, &githubAPIOpts)
	if err != nil {
		fmt.Fprintf(cli.errStream, err.Error())
		return ExitCodeError
	}

	// Fetch All assets ID which is already on Github
	if ghrOpts.Replace {
		err = FetchAssetID(assets, &githubAPIOpts)
		if err != nil {
			fmt.Fprintf(cli.errStream, err.Error())
			return ExitCodeError
		}
	}

	// Use CPU efficiently
	cpu := runtime.NumCPU()
	runtime.GOMAXPROCS(cpu)

	// Limit amount of parallelism by number of logic CPU
	if ghrOpts.Parallel <= 0 {
		ghrOpts.Parallel = runtime.NumCPU()
	}

	// Start releasing
	errors := UploadAssets(assets, &ghrOpts, &githubAPIOpts)
	if len(errors) > 0 {
		fmt.Fprintf(cli.errStream, "%d errors occurred:\n", len(errors))
		for _, err := range errors {
			fmt.Fprintf(cli.errStream, "--> %s\n", err)
		}
		return ExitCodeRleaseError
	}

	return ExitCodeOK
}

// setToken sets GitHub API Token.
func setToken(githubOpts *GitHubAPIOpts) (err error) {
	// Use flag value.
	if githubOpts.Token != "" {
		return nil
	}

	// Use Environmental value.
	if os.Getenv("GITHUB_TOKEN") != "" {
		githubOpts.Token = os.Getenv("GITHUB_TOKEN")
		return nil
	}

	// Use .gitconfig value.
	githubOpts.Token, err = gitconfig.GithubToken()
	if err != nil {
		return err
	}

	// Confirm value is not blank.
	if githubOpts.Token == "" {
		return fmt.Errorf("Please set your Github API Token in the GITHUB_TOKEN env var\n")
	}

	return nil
}

// setOwner sets repository owner name.
func setOwner(githubOpts *GitHubAPIOpts) (err error) {
	// Use flag value.
	if githubOpts.OwnerName != "" {
		return nil
	}

	// Use .gitconfig value.
	githubOpts.OwnerName, err = gitconfig.Username()
	if err != nil {
		return err
	}

	// Confirm value is not blank.
	if githubOpts.OwnerName == "" {
		return fmt.Errorf("Cound not retrieve git user name\n")
	}

	return nil
}

// setRepo sets repository name.
func setRepo(githubOpts *GitHubAPIOpts) (err error) {
	// Use flag value.
	if githubOpts.RepoName != "" {
		return nil
	}

	// Use .gitconfig value.
	githubOpts.RepoName, err = gitconfig.Repository()
	if err != nil {
		return err
	}

	// Confirm value is not blank.
	if githubOpts.RepoName == "" {
		return fmt.Errorf("cound not retrieve repository name\n")
	}

	return nil
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
