package main

import (
	"fmt"
	"io"
	"net/url"
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
	ExitCodeInvalidURL
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
	var githubAPIOpts GitHubAPIOpts
	var ghrOpts GhrOpts
	var stat bool
	var err error

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
	flags.BoolVar(&stat, []string{"-stat"}, false,
		"Show statical infomation")

	// Receive general options.
	version := flags.Bool([]string{"v", "-version"}, false,
		"Print version information and quit")
	help := flags.Bool([]string{"h", "-help"}, false,
		"Print this message and quit")
	debug := flags.Bool([]string{"-debug"}, false,
		"Run as DEBUG mode")

	// Parse all the flags
	if err := flags.Parse(args[1:]); err != nil {
		return ExitCodeParseFlagsError
	}

	// Version
	if *version {
		fmt.Fprintf(cli.errStream, "ghr version %s, build %s \n", Version, GitCommit)
		return ExitCodeOK
	}

	// Help
	if *help {
		fmt.Fprintf(cli.errStream, helpText)
		return ExitCodeOK
	}

	// Run as DEBUG mode
	if *debug {
		os.Setenv("DEBUG", "1")
	}

	// Set BaseURL
	_ = setBaseURL(&githubAPIOpts)

	// Set Token
	err = setToken(&githubAPIOpts)
	if err != nil {
		errMsg := fmt.Sprintf("Could not retrieve GitHub API Token.\n")
		errMsg += "Please set your Github API Token in the GITHUB_TOKEN env var.\n"
		errMsg += "Or set one via `-t` option.\n"
		errMsg += "See about GitHub API Token on https://github.com/blog/1509-personal-api-tokens\n"
		fmt.Fprint(cli.errStream, ColoredError(errMsg))
		return ExitCodeTokenNotFound
	}

	// Set repository owner name.
	err = setOwner(&githubAPIOpts)
	if err != nil {
		errMsg := fmt.Sprintf("Could not retrieve repository user name: %s\n", err)
		errMsg += "ghr try to retrieve git user name from `~/.gitcofig` file.\n"
		errMsg += "Please set one via -u option or `~/.gitconfig` file.\n"
		fmt.Fprintf(cli.errStream, ColoredError(errMsg))
		return ExitCodeOwnerNotFound
	}

	// Set repository owner name.
	err = setRepo(&githubAPIOpts)
	if err != nil {
		errMsg := fmt.Sprintf("Could not retrieve repository name: %s\n", err)
		errMsg += "ghr try to retrieve github repository name from `.git/cofig` file.\n"
		errMsg += "Please be sure you're in github repository. Or set one via `-r` options.\n"
		fmt.Fprintf(cli.errStream, ColoredError(errMsg))
		return ExitCodeRepoNotFound
	}

	// Display statical information.
	if stat {
		err = ShowStat(cli.outStream, &githubAPIOpts)
		if err != nil {
			fmt.Fprintf(cli.errStream, ColoredError(err.Error()))
			return ExitCodeError
		}
		return ExitCodeOK
	}

	// Get the parsed arguments
	parsedArgs := flags.Args()
	if len(parsedArgs) != 2 {
		fmt.Fprintf(cli.errStream, ColoredError("Argument error: must specify two arguments - tag, path\n\n"))
		fmt.Fprintf(cli.errStream, helpText)
		return ExitCodeBadArgs
	}

	// Get the tag of release and path
	tag, path := parsedArgs[0], parsedArgs[1]
	githubAPIOpts.TagName = tag

	// Get the asset to upload.
	assets, err := GetLocalAssets(path)
	if err != nil {
		fmt.Fprintf(cli.errStream, ColoredError(err.Error()))
		return ExitCodeError
	}

	// Create release.
	err = CreateRelease(&ghrOpts, &githubAPIOpts)
	if err != nil {
		fmt.Fprintf(cli.errStream, ColoredError(err.Error()))
		return ExitCodeError
	}

	// Fetch All assets ID which is already on Github.
	// This is invorked when `--replace` option is used.
	if ghrOpts.Replace {
		err = FetchAssetID(assets, &githubAPIOpts)
		if err != nil {
			fmt.Fprintf(cli.errStream, ColoredError(err.Error()))
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
		errMsg := fmt.Sprintf("%d errors occurred:\n", len(errors))
		fmt.Fprintf(cli.errStream, ColoredError(errMsg))
		for _, err := range errors {
			fmt.Fprintf(cli.errStream, ColoredError(fmt.Sprintf("--> %s\n", err)))
		}
		return ExitCodeRleaseError
	}

	return ExitCodeOK
}

// setBaseURL sets Base GitHub API URL
// Default is https://api.github.com
func setBaseURL(githubOpts *GitHubAPIOpts) (err error) {
	if os.Getenv("GITHUB_API") == "" {
		return nil
	}

	// Use Environmental value.
	u := os.Getenv("GITHUB_API")

	// Pase it as url.URL
	baseURL, err := url.Parse(u)
	if err != nil {
		return fmt.Errorf("failed to parse url %s", u)
	}

	Debug("BaseURL:", baseURL)

	// Set it
	githubOpts.BaseURL = baseURL

	return nil
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
	githubOpts.Token, _ = gitconfig.GithubToken()

	// Confirm value is not blank.
	if githubOpts.Token == "" {
		return fmt.Errorf("GitHub API token is not found.")
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
		return fmt.Errorf("key `user.name` is not found in `~/.gitconfig`")
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
		return fmt.Errorf("key `remote.origin.url` is not found in `.git/config`")
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
  --stat             Show how many tool donwloaded
　--replace          Replace asset if target already exists
　--delete           Delete release and its git tag if same version exists
  --draft            Create unpublised release
  --prerelease       Create prerelease
  -h, --help         Print this message and quit
  -v, --version      Print version information and quit
  --debug=false      Run as DEBUG mode

Example:
  $ ghr v1.0.0 dist/
  $ ghr v1.0.2 dist/tool.zip
  $ ghr --stat
`
