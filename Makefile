NAME ?= queue

.PHONY: build test clean

## build: Build the binary
build: 
	go build -o $(NAME)

## test: Run tests (Package tests are not run concurrently due to hardcoded network ports)
test:
	go test -p 1 ./...

## clean: Clean build files, tmp files
clean:
	go clean -testcache
	rm -f $(NAME)
	rm -rf ./tmp/* 