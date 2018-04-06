package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
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

	client := testGithubClient(t)
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

	client := testGithubClient(t)
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
