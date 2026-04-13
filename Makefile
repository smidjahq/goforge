BINARY_NAME = goforge
MODULE = github.com/smidjahq/goforge

.PHONY: build run version test test-short test-full vet lint clean help

## build: compile the binary
build:
	go build -o $(BINARY_NAME) .

## run: start the interactive TUI
run:
	go run . create

## version: print version info
version:
	go run . version

## test: run the test suite (skips compilation tests that shell out to go build/go mod tidy)
test:
	go test ./... -short

## test-full: run the full test suite including compilation tests (requires network)
test-full:
	go test ./...

## test-verbose: run short tests with verbose output
test-verbose:
	go test ./... -short -v

## vet: run go vet
vet:
	go vet ./...

## lint: run vet and check formatting
lint: vet
	@gofmt -l . | grep -v vendor | tee /dev/stderr | (! read line)

## clean: remove build artifacts
clean:
	rm -f $(BINARY_NAME)

## tidy: tidy go modules
tidy:
	go mod tidy

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/^## //'