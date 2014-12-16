package main

import (
	"fmt"
	"io"
)

type Stat struct {
	// TagName is relase tag
	TagName string

	// DownloadCount is download number of its release
	DownloadCount int

	// Best is marked when it's most downloaded release
	Best bool

	// MaxLength
	MaxLength int
}

// ShowStat displays dowload counts of latest releases.
func ShowStat(outStream io.Writer, apiOpts *GitHubAPIOpts) (err error) {
	// Get Statical infomation
	stats, err := GetStat(apiOpts)
	if err != nil {
		return err
	}

	// Find most donwloaded release
	MarkBest(stats)

	// Display all stats
	for _, s := range stats {
		msg := fmt.Sprintf("%-20s: %4d downloads\n", s.TagName, s.DownloadCount)
		if s.Best {
			fmt.Fprintf(outStream, Hot(msg))
			continue
		}

		fmt.Fprintf(outStream, msg)
	}

	return
}

// GetStat gets download counts of all releases
func GetStat(apiOpts *GitHubAPIOpts) (stats []*Stat, err error) {

	// Create client
	client := NewOAuthedClient(apiOpts)

	// Get All releases
	releases, res, err := client.Repositories.ListReleases(apiOpts.OwnerName, apiOpts.RepoName, nil)
	if err != nil {
		return stats, err
	}

	err = CheckStatusOK(res)
	if err != nil {
		return stats, err
	}

	for _, r := range releases {

		var count int
		for _, a := range r.Assets {
			count += *a.DownloadCount
		}
		stats = append(stats, &Stat{TagName: *r.TagName, DownloadCount: count})
	}

	return stats, nil
}

// MarkBest marks most donloaded release
func MarkBest(stats []*Stat) {

	// Find most max download num
	var max int
	for _, s := range stats {
		if s.DownloadCount > max {
			max = s.DownloadCount
		}
	}

	// Mark best
	for _, s := range stats {
		if s.DownloadCount >= max {
			s.Best = true
		}
	}

}
