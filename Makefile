VERSION = $(shell godzil show-version)
COMMIT = $(shell git rev-parse --short HEAD)
EXTERNAL_TOOLS = \
    golang.org/x/lint/golint            \
    github.com/Songmu/godzil/cmd/godzil \
    github.com/mattn/goveralls                \
    github.com/Songmu/gocredits/cmd/gocredits \
    golang.org/x/tools/cmd/cover

ifdef update
  u=-u
endif

export GO111MODULE=on

.PHONY: default
default: test

.PHONY: deps
deps:
	go get ${u} -d

# install external tools for this project
.PHONY: devel-deps
devel-deps: deps
	@for tool in $(EXTERNAL_TOOLS) ; do \
      echo "Installing $$tool" ; \
      GO111MODULE=off go get $$tool; \
    done

# build generate binary on './bin' directory.
.PHONY: build
build:
	go build -ldflags "-X main.GitCommit=$(COMMIT)" -o bin/ghr

.PHONY: bump
bump: devel-deps
	godzil release

.PHONY: crossbuild
crossbuild: devel-deps
	goxz -pv=v${VERSION} -build-ldflags="-X main.GitCommit=${COMMIT}" \
        -arch=386,amd64 -d=./pkg/dist/v${VERSION}

# install installs binary on $GOPATH/bin directory.
.PHONY: install
install:
	go install -ldflags "-X main.GitCommit=$(COMMIT)"

# package runs compile.sh to run gox and zip them.
# Artifacts will be generated in './pkg' directory
.PHONY: package
package: devel-deps
	@sh -c "'$(CURDIR)/scripts/package.sh'"

.PHONY: brew
brew: package
	go run release/main.go v$(VERSION) pkg/dist/v$(VERSION)/ghr_v$(VERSION)_darwin_amd64.zip > ../homebrew-ghr/ghr.rb

.PHONY: upload
upload: build devel-deps
	bin/ghr -v
	bin/ghr v$(VERSION) pkg/dist/v$(VERSION)

.PHONY: test-all
test-all: lint test

.PHONY: test
test:
	go test -v -parallel=4 ./...

.PHONY: test-race
test-race:
	go test -v -race ./...

.PHONY: lint
lint: vet
	go vet ./...
	golint -set_exit_status ./...

.PHONY: cover
cover:
	go test -coverprofile=cover.out
	go tool cover -html cover.out
	rm cover.out

.PHONY: release
release: bump package upload
