VERSION=$(shell awk -F'"' '/"version":/ {print $$4}' version.json)
GOFLAGS=-ldflags="-X github.com/storacha/storage/pkg/build.version=$(VERSION)"
TAGS?=

.PHONY: all build storage install test clean calibnet

all: build

build: storage

storage:
	go build $(GOFLAGS) $(TAGS) -o ./storage ./cmd/storage

install:
	go install ./cmd/storage

test:
	go test ./...

clean:
	rm -f ./storage

# special target that sets the calibnet tag and invokes build
calibnet: TAGS=-tags calibnet
calibnet: build
