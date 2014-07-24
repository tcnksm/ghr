DEBUG_FLAG = $(if $(DEBUG),-debug)

deps:
	go get -d -t ./...

test: deps
	go test -v ./...

install: deps
	go install
