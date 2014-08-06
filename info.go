package main

type Info struct {
	ID              int
	Token           string
	TagName         string
	RepoName        string
	OwnerName       string
	TargetCommitish string
	Body            string
	Draft           bool
	Prerelease      bool
}
