VERSION = $(shell gobump show -r)
COMMIT = $(shell git describe --always)
EXTERNAL_TOOLS = \
    github.com/Songmu/goxz/cmd/goxz \
    github.com/motemen/gobump \
    github.com/Songmu/ghch/cmd/ghch

default: test

# install external tools for this project
bootstrap:
	@for tool in $(EXTERNAL_TOOLS) ; do \
      echo "Installing $$tool" ; \
      go get $$tool; \
    done

# build generate binary on './bin' directory.
build:
	go build -ldflags "-X main.GitCommit=$(COMMIT)" -o bin/ghr

bump: bootstrap
	@sh -c "'$(CURDIR)/scripts/bump.sh'"

crossbuild: bootstrap
	goxz -pv=v${VERSION} -build-ldflags="-X main.GitCommit=${COMMIT}" \
        -arch=386,amd64 -d=./pkg/dist/v${VERSION}

# install installs binary on $GOPATH/bin directory.
install:
	go install -ldflags "-X main.GitCommit=$(COMMIT)"

# package runs compile.sh to run gox and zip them.
# Artifacts will be generated in './pkg' directory
package: bootstrap
	@sh -c "'$(CURDIR)/scripts/package.sh'"

brew: package
	go run release/main.go v$(VERSION) pkg/dist/v$(VERSION)/ghr_v$(VERSION)_darwin_amd64.zip > ../homebrew-ghr/ghr.rb

upload: build bootstrap
	bin/ghr -v
	bin/ghr v$(VERSION) pkg/dist/v$(VERSION)

test-all: vet lint test

test:
	go test -v -parallel=4 ./...

test-race:
	go test -v -race ./...

vet:
	go vet ./...

lint:
	@go get github.com/golang/lint/golint
	go list ./... | grep -v vendor | xargs -n1 golint -set_exit_status

cover:
	@go get golang.org/x/tools/cmd/cover
	go test -coverprofile=cover.out
	go tool cover -html cover.out
	rm cover.out

release: bump package upload

.PHONY: bootstrap bump crossbuild build install package brew test test-race test-all vet lint cover release
