.PHONY: build install clean lint

BINARY  := arpmap
GOOS    ?= $(shell go env GOOS)
GOARCH  ?= $(shell go env GOARCH)

build:
	go build -o $(BINARY) .

install:
	go install .

# Cross-compile for common targets
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)-linux-amd64 .

build-darwin:
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY)-darwin-arm64 .

lint:
	go vet ./...

clean:
	rm -f $(BINARY) $(BINARY)-linux-* $(BINARY)-darwin-*
