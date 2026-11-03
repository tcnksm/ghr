package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRun_commitish(t *testing.T) {
	for _, tc := range []struct {
		name      string
		commitish string
	}{
		{name: "omitted"},
		{name: "branch", commitish: "main"},
		{name: "SHA", commitish: "0123456789abcdef0123456789abcdef01234567"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := make(chan map[string]interface{}, 1)
			mux := http.NewServeMux()
			mux.HandleFunc("GET /repos/owner/repo/releases", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "[]")
			})
			mux.HandleFunc("GET /repos/owner/repo/releases/tags/run", func(w http.ResponseWriter, r *http.Request) {
				http.NotFound(w, r)
			})
			mux.HandleFunc("POST /repos/owner/repo/releases", func(w http.ResponseWriter, r *http.Request) {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("Decode release request: %s", err)
					http.Error(w, "invalid JSON", http.StatusBadRequest)
					return
				}
				requests <- req
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, `{"id":1,"upload_url":"http://%s/repos/owner/repo/releases/1/assets"}`, r.Host)
			})
			mux.HandleFunc("PATCH /repos/owner/repo/releases/1", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"id":1}`)
			})
			server := httptest.NewServer(mux)
			defer server.Close()
			t.Setenv(EnvGitHubAPI, server.URL+"/")

			outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
			cli := &CLI{outStream: outStream, errStream: errStream}
			args := []string{"ghr", "-username", "owner", "-repository", "repo", "-token", "test-token"}
			if tc.commitish != "" {
				args = append(args, "--commitish", tc.commitish)
			}
			args = append(args, "run")
			if got := cli.Run(args); got != ExitCodeOK {
				t.Fatalf("Run exits %d, want %d\n\n%s", got, ExitCodeOK, errStream.String())
			}

			select {
			case req := <-requests:
				got, present := req["target_commitish"]
				if tc.commitish == "" {
					if present {
						t.Errorf("target_commitish = %v, want absent", got)
					}
				} else if !present || got != tc.commitish {
					t.Errorf("target_commitish = %v, want %q", got, tc.commitish)
				}
			default:
				t.Fatal("No release-creation request received")
			}
		})
	}
}

func TestRun(t *testing.T) {
	client := testGithubClient(t)

	t.Parallel()

	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &CLI{outStream: outStream, errStream: errStream}

	tag := "run"
	command := fmt.Sprintf(
		"ghr -username %s -repository %s %s %s", TestOwner, TestRepo, tag, TestDir)

	args := strings.Split(command, " ")
	if got, want := cli.Run(args), ExitCodeOK; got != want {
		t.Fatalf("%q exits %d, want %d\n\n%s", command, got, want, errStream.String())
	}

	release, err := client.GetRelease(context.TODO(), tag)
	if err != nil {
		t.Fatalf("GetRelease failed: %s\n\n%s", err, outStream.String())
	}
	defer func() {
		if err := client.DeleteRelease(context.TODO(), *release.ID); err != nil {
			t.Fatal("DeleteRelease failed:", err)
		}

		if err := client.DeleteTag(context.TODO(), tag); err != nil {
			t.Fatal("DeleteTag failed:", err)
		}
	}()

	want := "==> Create a new release"
	if got := outStream.String(); !strings.Contains(got, want) {
		t.Fatalf("Run outputs %q, want %q", got, want)
	}

	if got, want := outStream.String(), "--> Uploading:"; !strings.Contains(got, want) {
		t.Fatalf("Run outputs %q, want %q", got, want)
	}

	assets, err := client.ListAssets(context.TODO(), *release.ID)
	if err != nil {
		t.Fatal("ListAssets failed:", err)
	}

	if got, want := len(assets), 4; got != want {
		t.Fatalf("ListAssets number = %d, want %d", got, want)
	}
}

func TestRun_recreate(t *testing.T) {
	client := testGithubClient(t)

	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &CLI{outStream: outStream, errStream: errStream}

	tag := "run-recreate"
	command := fmt.Sprintf(
		"ghr -username %s -repository %s %s %s", TestOwner, TestRepo, tag, TestDir)

	args := strings.Split(command, " ")
	if got, want := cli.Run(args), ExitCodeOK; got != want {
		t.Fatalf("%q exits %d, want %d\n\n%s", command, got, want, errStream.String())
	}

	// Prevent sending requests too much at same time
	time.Sleep(3 * time.Second)

	command = fmt.Sprintf(
		"ghr -username %s -repository %s -recreate %s %s", TestOwner, TestRepo, tag, TestDir)

	args = strings.Split(command, " ")
	if got, want := cli.Run(args), ExitCodeOK; got != want {
		t.Fatalf("%q exits %d, want %d\n\n%s", command, got, want, errStream.String())
	}

	release, err := client.GetRelease(context.TODO(), tag)
	if err != nil {
		t.Fatalf("GetRelease failed: %s\n\n%s", err, outStream.String())
	}
	defer func() {
		time.Sleep(5 * time.Second)
		if err := client.DeleteRelease(context.TODO(), *release.ID); err != nil {
			t.Fatal("DeleteRelease failed:", err)
		}

		if err := client.DeleteTag(context.TODO(), tag); err != nil {
			t.Fatal("DeleteTag failed:", err)
		}
	}()

	want := "==> Recreate a release"
	if got := outStream.String(); !strings.Contains(got, want) {
		t.Fatalf("Run outputs %q, want %q", got, want)
	}
}

func TestRun_versionFlag(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &CLI{outStream: outStream, errStream: errStream}
	command := "ghr --version"
	args := strings.Split(command, " ")

	if got, want := cli.Run(args), ExitCodeOK; got != want {
		t.Fatalf("%q exits %d, want %d", command, got, want)
	}

	want := fmt.Sprintf("ghr version v%s", Version)
	got := outStream.String()
	if !strings.Contains(got, want) {
		t.Fatalf("%q output %q, want = %q", command, got, want)
	}
}

func TestRetrieveOwnerName(t *testing.T) {
	testCases := []struct {
		in, expected string
	}{
		{"git@github.com:tcnksm/ghr.git", "tcnksm"},
		{"https://github.com/motemen/ghq.git", "motemen"},
		{"git://github.com/Songmu/ghg.git", "Songmu"},
		{"hoge", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			out := retrieveOwnerName(tc.in)
			if out != tc.expected {
				t.Errorf("out: %s, expected: %s", out, tc.expected)
			}
		})
	}
}
