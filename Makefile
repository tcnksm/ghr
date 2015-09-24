DEBUG_FLAG = $(if $(DEBUG),-debug)
COMMIT = $$(git describe --always)

ORG := github.com/tcnksm
REPO := $(ORG)/ghr
REPO_PATH :=$(GOPATH)/src/$(REPO)

clean:
	rm $(GOPATH)/bin/ghr

deps:
	go get -d -t ./...

test: deps
	go test -v ./...

build: deps
	go build -ldflags "-X main.GitCommit=\"$(COMMIT)\"" -o bin/ghr

build-docker:
	/usr/local/bin/docker run --rm -v $(REPO_PATH):/gopath/src/$(REPO) -w /gopath/src/$(REPO) tcnksm/gox:1.5.1 sh -c "make build"

install: deps
	go install -ldflags "-X main.GitCommit=\"$(COMMIT)\""

