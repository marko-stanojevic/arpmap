.PHONY: build install clean lint build-linux build-darwin build-windows

BINARY  := arpmap
GOOS    ?= $(shell go env GOOS)
GOARCH  ?= $(shell go env GOARCH)
DIST    := ../dist

ifeq ($(GOOS),windows)
BIN_EXT := .exe
else
BIN_EXT :=
endif

build:
	mkdir -p $(DIST)/debug
	go build -o $(DIST)/debug/$(BINARY)$(BIN_EXT) .

install:
	go install .

# Cross-compile for common targets
build-linux:
	mkdir -p $(DIST)
	GOOS=linux GOARCH=amd64 go build -o $(DIST)/$(BINARY)-linux-amd64 .

build-darwin:
	mkdir -p $(DIST)
	GOOS=darwin GOARCH=arm64 go build -o $(DIST)/$(BINARY)-darwin-arm64 .

build-windows:
	mkdir -p $(DIST)
	GOOS=windows GOARCH=amd64 go build -o $(DIST)/$(BINARY)-windows-amd64.exe .

lint:
	go vet ./...

clean:
	rm -rf $(DIST)
