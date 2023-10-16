VERSION = $(shell godzil show-version)
COMMIT = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X main.GitCommit=$(COMMIT)"
ifdef update
  u=-u
endif

.PHONY: default
default: test

.PHONY: deps
deps:
	go get ${u} -d
	go mod tidy

# install external tools for this project
.PHONY: devel-deps
devel-deps: deps
	go install github.com/Songmu/godzil/cmd/godzil@latest

# build generate binary on './bin' directory.
.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) -o bin/ghr

CREDITS: go.sum devel-deps
	godzil credits -w

.PHONY: crossbuild
crossbuild: CREDITS
	CGO_ENABLED=0 godzil crossbuild -pv=v${VERSION} -build-ldflags=$(BUILD_LDFLAGS) \
        -arch=amd64,arm64 -os=windows,darwin,linux,freebsd -d=./pkg/dist/v${VERSION}
	cd pkg/dist/v${VERSION} && shasum -a 256 * > ./v${VERSION}_SHASUMS

# install installs binary on $GOPATH/bin directory.
.PHONY: install
install:
	go install -ldflags=$(BUILD_LDFLAGS)

.PHONY: upload
upload: build devel-deps
	bin/ghr -v
	bin/ghr v$(VERSION) pkg/dist/v$(VERSION)

.PHONY: test
test: deps
	go test -v -parallel=4 ./...

.PHONY: test-race
test-race:
	go test -v -race ./...

.PHONY: cover
cover:
	go test -coverprofile=cover.out
	go tool cover -html cover.out

.PHONY: release
release: crossbuild upload
