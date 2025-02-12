VERSION=$(shell awk -F'"' '/"version":/ {print $$4}' version.json)
GOFLAGS=-ldflags="-X github.com/storacha/storage/pkg/build.version=$(VERSION)"

storage:
	go build $(GOFLAGS) -o ./storage ./cmd/storage

.PHONY: install

install:
	go install ./cmd/storage

.PHONY: clean

clean:
	rm -f ./storage
