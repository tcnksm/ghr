COMMIT = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X main.GitCommit=$(COMMIT)"
u := $(if $(update),-u)

.PHONY: default
default: test

.PHONY: deps
deps:
	go get ${u}
	go mod tidy

# install external tools for this project
.PHONY: devel-deps
devel-deps: deps
	go install github.com/Songmu/gocredits/cmd/gocredits@v0.5.0

# build generate binary on './bin' directory.
.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) -o bin/ghr

.PHONY: prepare-release
prepare-release: devel-deps
	go mod tidy
	gocredits . > CREDITS
	git update-index --add --remove -- go.mod go.sum CREDITS

# install installs binary on $GOPATH/bin directory.
.PHONY: install
install:
	go install -ldflags=$(BUILD_LDFLAGS)

.PHONY: upload
upload: build
	bin/ghr -v
	bin/ghr $(VERSION) pkg/dist/$(VERSION)

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
