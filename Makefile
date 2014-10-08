DEBUG_FLAG = $(if $(DEBUG),-debug)
COMMIT = $$(git describe --always)

deps:
	go get -d -t ./...

test: deps
	go test -v ./...

install: deps
	go install -ldflags "-X main.GitCommit \"$(COMMIT)\""
