DEBUG_FLAG = $(if $(DEBUG),-debug)
COMMIT = $$(git describe --always)

deps:
	go get -d -t ./...

test: deps
	go test -v ./...

build: deps
	go build -ldflags "-X main.GitCommit \"$(COMMIT)\"" -o ghr

install: deps
	go install -ldflags "-X main.GitCommit \"$(COMMIT)\""
