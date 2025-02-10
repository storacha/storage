VERSION=$(shell awk -F'"' '/"version":/ {print $$4}' version.json)

storage:
	go build -ldflags="-s -w -X github.com/storacha/storage/pkg/build.version=$(VERSION)" -o ./storage ./cmd/storage

.PHONY: install

install:
	go install -ldflags="-s -w -X github.com/storacha/storage/pkg/build.version=$(VERSION)" ./cmd/storage

.PHONY: clean

clean:
	rm -f ./storage
