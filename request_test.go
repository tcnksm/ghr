package main

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestCheckStatusOK(t *testing.T) {
	RegisterTestingT(t)

	err := checkStatusOK(200, "OK")
	Expect(err).ToNot(HaveOccurred())

	err = checkStatusOK(500, "Internal Server Error")
	Expect(err).To(HaveOccurred())
}

func TestCheckStatusCreated(t *testing.T) {
	RegisterTestingT(t)

	err := checkStatusCreated(201, "Created")
	Expect(err).ToNot(HaveOccurred())

	err = checkStatusOK(500, "Internal Server Error")
	Expect(err).To(HaveOccurred())
}
