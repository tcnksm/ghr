package main

import (
	"net/url"
	"testing"
)

func TestNewClient(t *testing.T) {
	reset := setEnv(EnvGitHubAPI, "")
	defer reset()

	exp, _ := url.Parse("https://api.github.com/")

	apiOpts := &GitHubAPIOpts{}
	setBaseURL(apiOpts)

	client := NewOAuthedClient(apiOpts)
	if *client.BaseURL != *exp {
		t.Errorf("expected %q to eq %q", client.BaseURL, exp)
	}
}

func TestNewClient_enterprise(t *testing.T) {
	// Set API endpoint via environmental variable
	in := "http://github.company.com/api/v3/"
	reset := setEnv(EnvGitHubAPI, in)
	defer reset()

	exp, _ := url.Parse(in)

	apiOpts := &GitHubAPIOpts{}
	setBaseURL(apiOpts)

	client := NewOAuthedClient(apiOpts)
	if *client.BaseURL != *exp {
		t.Errorf("expected %q to eq %q", client.BaseURL, exp)
	}
}

func TestExtractUploadURL(t *testing.T) {

	in := &GitHubAPIOpts{
		UploadURL: "https://uploads.github.com/repos/tcnksm/ghr/releases/786857/assets{?name}\"",
	}
	exp, _ := url.Parse("https://uploads.github.com/")

	out := ExtractUploadURL(in)
	if *out != *exp {
		t.Errorf("expected %q to eq %q", out, exp)
	}
}

func TestExtractUploadURL_enterprise(t *testing.T) {

	in := &GitHubAPIOpts{
		UploadURL: "https://github.company.com/api/v3/repos/tcnksm/ghr/releases/786857/assets{?name}\"",
	}
	exp, _ := url.Parse("https://github.company.com/api/v3/")

	out := ExtractUploadURL(in)
	if *out != *exp {
		t.Errorf("expected %q to eq %q", out, exp)
	}
}
