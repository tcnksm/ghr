VERSION = $(shell grep 'Version string' version.go | sed -E 's/.*"(.+)"$$/\1/')
COMMIT = $(shell git describe --always)
PACKAGES = $(shell go list ./... | grep -v '/vendor/')
EXTERNAL_TOOLS = github.com/mitchellh/gox	

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

# install installs binary on $GOPATH/bin directory.
install: 
	go install -ldflags "-X main.GitCommit=$(COMMIT)"

# package runs compile.sh to run gox and zip them.
# Artifacts will be generated in './pkg' directory
package: 
	@sh -c "'$(CURDIR)/scripts/package.sh'"

brew: package
	go run release/main.go $(VERSION) pkg/dist/$(VERSION)/ghr_$(VERSION)_darwin_amd64.zip > ../homebrew-ghr/ghr.rb

upload: build
	bin/ghr -v
	bin/ghr $(VERSION) pkg/dist/$(VERSION)

test-all: vet lint test

test: 
	go test -v -parallel=4 ${PACKAGES}

test-race:
	go test -v -race ${PACKAGES}

vet:
	go vet ${PACKAGES}

lint:
	@go get github.com/golang/lint/golint
	go list ./... | grep -v vendor | xargs -n1 golint 

cover:
	@go get golang.org/x/tools/cmd/cover		
	go test -coverprofile=cover.out
	go tool cover -html cover.out
	rm cover.out

.PHONY: bootstrap build install package brew test test-race test-all vet lint cover  