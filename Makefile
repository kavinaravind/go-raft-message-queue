NAME ?= queue

.PHONY: build test clean

## build: Build the binary
build: 
	go build -o $(NAME)

## test: Run tests
test:
	go test -v ./...

## clean: Clean build files
clean:
	go clean
	rm -f $(NAME)