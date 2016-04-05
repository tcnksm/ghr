package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"runtime"
	"time"

	flag "github.com/docker/docker/pkg/mflag"
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

// EnvDebug is environmental var to handle debug mode
const EnvDebug = "GHR_DEBUG"

// Debugf prints debug output when EnvDebug is given
func Debugf(format string, args ...interface{}) {
	if env := os.Getenv(EnvDebug); len(env) != 0 {
		fmt.Fprintf(os.Stdout, "[DEBUG] "+format+"\n", args...)
	}
}

// GhrOpts are the options for ghr related
type GhrOpts struct {
	// Parallel determines number of amount of parallelism.
	// Default is number of CPU.
	Parallel int

	// Replace determines repalce asset on GitHub if it exists.
	Replace bool

	// Detele determines delete release if it exists.
	Delete bool

	// OutCh receive info output from uploading process
	outCh chan string

	// ErrCH receive error output from uploading process
	errCh chan string
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
	flags.Usage = func() {
		fmt.Fprint(cli.errStream, helpText)
	}

	// Options for GitHub API.
	flags.StringVar(&githubAPIOpts.OwnerName, []string{"u", "-username"}, "", "")
	flags.StringVar(&githubAPIOpts.RepoName, []string{"r", "-repository"}, "", "")
	flags.StringVar(&githubAPIOpts.Token, []string{"t", "-token"}, "", "")
	flags.StringVar(&githubAPIOpts.TargetCommitish, []string{"c", "-commitish"}, "", "")
	flags.BoolVar(&githubAPIOpts.Draft, []string{"-draft"}, false, "")
	flags.BoolVar(&githubAPIOpts.Prerelease, []string{"-prerelease"}, false, "")

	// Options to change ghr work.
	flags.IntVar(&ghrOpts.Parallel, []string{"p", "-parallel"}, -1, "")
	flags.BoolVar(&ghrOpts.Replace, []string{"-replace"}, false, "")
	flags.BoolVar(&ghrOpts.Delete, []string{"-delete"}, false, "")
	flags.BoolVar(&stat, []string{"-stat"}, false, "")

	// General options
	version := flags.Bool([]string{"v", "-version"}, false, "")
	debug := flags.Bool([]string{"-debug"}, false, "")

	// Parse all the flags
	if err := flags.Parse(args[1:]); err != nil {
		return ExitCodeParseFlagsError
	}

	// Show version. It also try to fetch latest version information from github
	if *version {
		fmt.Fprintf(cli.errStream, "ghr version %s, build %s \n", Version, GitCommit)

		select {
		case res := <-verCheckCh:
			if res != nil && res.Outdated {
				msg := fmt.Sprintf("Latest version of ghr is %s, please update it\n", res.Current)
				fmt.Fprint(cli.errStream, ColoredError(msg))
			}
		case <-time.After(CheckTimeout):
			// do nothing
		}

		return ExitCodeOK
	}

	// Run as DEBUG mode
	if *debug {
		os.Setenv("GHR_DEBUG", "1")
	}

	// Set BaseURL
	_ = setBaseURL(&githubAPIOpts)

	// Set Token
	err = setToken(&githubAPIOpts)
	if err != nil {
		errMsg := fmt.Sprintf("Could not retrieve GitHub API Token.\n" +
			"Please set your Github API Token in the GITHUB_TOKEN env var.\n" +
			"Or set one via `-t` option.\n" +
			"See about GitHub API Token on https://github.com/blog/1509-personal-api-tokens\n",
		)
		fmt.Fprint(cli.errStream, ColoredError(errMsg))
		return ExitCodeTokenNotFound
	}

	// Set repository owner name.
	err = setOwner(&githubAPIOpts)
	if err != nil {
		errMsg := fmt.Sprintf("Could not retrieve repository user name: %s\n"+
			"ghr try to retrieve git user name from `~/.gitconfig` file.\n"+
			"Please set one via -u option or `~/.gitconfig` file.\n",
			err)
		fmt.Fprintf(cli.errStream, ColoredError(errMsg))
		return ExitCodeOwnerNotFound
	}

	// Set repository owner name.
	err = setRepo(&githubAPIOpts)
	if err != nil {
		errMsg := fmt.Sprintf("Could not retrieve repository name: %s\n"+
			"ghr try to retrieve github repository name from `.git/config` file.\n"+
			"Please be sure you're in github repository. Or set one via `-r` options.\n",
			err)
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
		fmt.Fprintf(cli.errStream, ColoredError("Argument error: must specify two arguments - tag, path\n"))
		return ExitCodeBadArgs
	}

	// Get the tag of release and path
	tag, path := parsedArgs[0], parsedArgs[1]
	githubAPIOpts.TagName = tag

	// Get the asset to upload.
	assets, err := GetLocalAssets(path)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get assets from %s: %s\n"+
			"Path must be included more than one file.\n",
			path, err)
		fmt.Fprintf(cli.errStream, ColoredError(errMsg))
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
	doneCh, outCh, errCh := UploadAssets(assets, &ghrOpts, &githubAPIOpts)

	// Receive messages
	statusCh := make(chan bool)
	go func() {
		errOccurred := false
		for {
			select {
			case out := <-outCh:
				fmt.Fprintf(cli.outStream, out)
			case err := <-errCh:
				fmt.Fprintf(cli.errStream, ColoredError(err))
				errOccurred = true
			case <-doneCh:
				statusCh <- errOccurred
				break
			}
		}
	}()

	// If more than one error is occured, return non-zero value
	errOccurred := <-statusCh
	if errOccurred {
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
	Debugf("Base GitHub API URL: %s", baseURL)

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
	githubOpts.OwnerName, err = gitconfig.GithubUser()
	if err != nil {
		githubOpts.OwnerName, _ = gitconfig.Username()
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

var helpText = `
Usage: ghr [options] TAG PATH

  ghr is a tool to create Release on Github and upload your artifacts to
  it. ghr parallelizes upload of multiple artifacts.

  You can use ghr on GitHub Enterprise. Change URL by GITHUB_API env var.

Options:

  --username, -u        GitHub username. By default, ghr extracts user
                        name from global gitconfig value.

  --repository, -r      GitHub repository name. By default, ghr extracts
                        repository name from current directory's .git/config
                        value.

  --token, -t           GitHub API Token. To use ghr, you will first need
                        to create a GitHub API token with an account which
                        has enough permissions to be able to create releases.
                        You can set this value via GITHUB_TOKEN env var.

  --parallel=-1         Parallelization factor. This option limits amount
                        of parallelism of uploading. By default, ghr uses
                        number of logic CPU of your PC.

  --delete              Delete release if it already created. If you want
                        to recreate release itself from beginning, use
                        this. Just want to upload same artifacts to same
                        release again, use --replace option.

  --replace             Replace artifacts if it is already uploaded. Same
                        artifact means, same release and same artifact
                        name.

  --stat=false          Show number of downloads of each release and quit.
                        This is special command.

Examples:

  $ ghr v1.0 dist/     Upload all artifacts which are in dist directory
                       with version v1.0. 

  $ ghr --stat         Show download number of each release and quit.

`
