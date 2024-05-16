NAME ?= queue

.PHONY: build test clean

## build: Build the binary
build: 
	go build -o $(NAME)

## test: Run tests
test:
	go test -p 1 -v ./...

## clean: Clean build files, tmp files
clean:
	go clean
	rm -f $(NAME)
	rm -rf ./tmp/* 