VERSION ?= $$(grep 'Version string' version.go | sed -E 's/.*"(.+)"$$/\1/')
COMMIT ?= $$(git describe --always)

default: test

clean:
	rm $(GOPATH)/bin/ghr
	rm bin/ghr

deps:
	go get -d -t -v .

# build generate binary on './bin' directory.
build: deps
	go build -ldflags "-X main.GitCommit=$(COMMIT)" -o bin/ghr

# package runs compile.sh to run gox and zip them.
# Artifacts will be generated in './pkg' directory
package: deps
	@sh -c "'$(CURDIR)/scripts/package.sh'"

brew: deps package
	go run release/main.go $(VERSION) pkg/dist/$(VERSION)/ghr_$(VERSION)_darwin_amd64.zip > ../homebrew-ghr/ghr.rb

ghr: brew build
	bin/ghr -v
	bin/ghr $(VERSION) pkg/dist/$(VERSION)


install: deps
	go install -ldflags "-X main.GitCommit=$(COMMIT)"

test-all: vet test

test: 
	go test -v -parallel=4 .

test-race: 
	@go test -race .

vet:
	go vet *.go

lint:
	@go get github.com/golang/lint/golint
	golint ./...

# cover shows test coverages
cover:
	@go get golang.org/x/tools/cmd/cover		
	godep go test -coverprofile=cover.out
	go tool cover -html cover.out
	rm cover.out

