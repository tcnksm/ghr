COMMIT = $$(git describe --always)

default: test

clean:
	rm $(GOPATH)/bin/ghr
	rm bin/ghr

deps:
	go get -d -t -v .

# build generate binary on './bin' directory.
build: deps
	go build -ldflags "-X main.GitCommit=\"$(COMMIT)\"" -o bin/ghr

# package runs compile.sh to run gox and zip them.
# Artifacts will be generated in './pkg' directory
package: deps
	@sh -c "'$(CURDIR)/scripts/package.sh'"

install: deps
	go install -ldflags "-X main.GitCommit=\"$(COMMIT)\""

test-all: test test-race vet lint

test: 
	go test -v -timeout=30s -parallel=4 .

test-race: 
	@go test -race .

vet:
	@go get golang.org/x/tools/cmd/vet
	go tool vet *.go

lint:
	@go get github.com/golang/lint/golint
	golint ./...

# cover shows test coverages
cover:
	@go get golang.org/x/tools/cmd/cover		
	godep go test -coverprofile=cover.out
	go tool cover -html cover.out
	rm cover.out

