VERSION=$(shell awk -F'"' '/"version":/ {print $$4}' version.json)
GOFLAGS=-ldflags="-X github.com/storacha/storage/pkg/build.version=$(VERSION)"

.PHONY: all build storage install test clean

all: build

build: storage

storage:
	go build $(GOFLAGS) -tags calibnet -o ./storage ./cmd/storage

install:
	go install ./cmd/storage

test:
	go test ./...

clean:
	rm -f ./storage
