VERSION=$(shell awk -F'"' '/"version":/ {print $$4}' version.json)
GOFLAGS=-ldflags="-X github.com/storacha/storage/pkg/build.version=$(VERSION)"

.PHONY: all build storage install test clean mockgen

all: build

build: storage

storage:
	go build $(GOFLAGS) -o ./storage ./cmd/storage

install:
	go install ./cmd/storage

test:
	go test ./...

clean:
	rm -f ./storage

mockgen:
	mockgen -source=./pkg/pdp/aggregator/interface.go -destination=./internal/mocks/aggregator.go -package=mocks
	mockgen -source=./pkg/pdp/curio/client.go -destination=./internal/mocks/curio_client.go -package=mocks
	mockgen -source=./internal/ipldstore/ipldstore.go -destination=./internal/mocks/ipldstore.go -package=mocks
	mockgen -source=./pkg/pdp/aggregator/steps.go -destination=./internal/mocks/steps.go -package=mocks