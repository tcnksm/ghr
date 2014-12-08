package main

import (
	. "github.com/onsi/gomega"
	"os"
	"testing"
)

func TestSetToken(t *testing.T) {
	RegisterTestingT(t)

	var (
		err error
	)

	*token = "1005fbo4311bfa51ypioerqyp081y54p1rq"
	err = setToken(token)
	Expect(err).NotTo(HaveOccurred())
	Expect(*token).To(Equal("1005fbo4311bfa51ypioerqyp081y54p1rq"))

	*token = ""
	reset := withEnv("751-84751-85471-095471323495435452-87")
	defer reset()

	err = setToken(token)
	Expect(err).NotTo(HaveOccurred())
	Expect(*token).To(Equal("751-84751-85471-095471323495435452-87"))

	*token = ""
	reset = withEnv("")
	defer reset()

	err = setToken(token)
	Expect(err).To(HaveOccurred())
}

func withEnv(value string) func() {
	preEnv := os.Getenv("GITHUB_TOKEN")
	os.Setenv("GITHUB_TOKEN", value)
	return func() {
		os.Setenv("GITHUB_TOKEN", preEnv)
	}
}
