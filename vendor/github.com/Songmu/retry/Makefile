BUILD_OS_TARGETS = "linux darwin freebsd windows"

test: deps
	go test ./...

deps:
	go get -d -v -t ./...
	go get github.com/golang/lint/golint
	go get golang.org/x/tools/cmd/cover
	go get github.com/axw/gocov/gocov
	go get github.com/mattn/goveralls

LINT_RET = .golint.txt
lint: deps
	go vet ./...
	rm -f $(LINT_RET)
	for os in "$(BUILD_OS_TARGETS)"; do \
		if [ $$os != "windows" ]; then \
			GOOS=$$os golint ./... | tee -a $(LINT_RET); \
		else \
			GOOS=$$os golint --min_confidence=0.9 ./... | tee -a $(LINT_RET); \
		fi \
	done
	test ! -s $(LINT_RET)

cover: deps
	goveralls

release:
	_tools/releng

.PHONY: test deps lint cover
