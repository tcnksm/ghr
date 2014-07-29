package main

type Info struct {
	ID              int
	Token           string
	TagName         string
	RepoName        string
	OwnerName       string
	TargetCommitish string
	Draft           bool
	Prerelease      bool
}

func NewInfo() Info {
	return Info{
		TargetCommitish: "master",
		Draft:           false,
		Prerelease:      false,
	}
}
