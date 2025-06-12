VERSION=$(shell awk -F'"' '/"version":/ {print $$4}' version.json)
GOFLAGS=-ldflags="-X github.com/storacha/piri/pkg/build.version=$(VERSION)"
TAGS?=

.PHONY: all build piri install test clean calibnet mockgen

all: build

build: piri

piri:
	go build $(GOFLAGS) $(TAGS) -o ./piri ./cmd/main.go

install:
	go install ./cmd/storage

test:
	go test ./...

clean:
	rm -f ./piri

mockgen:
	mockgen -source=./pkg/pdp/aggregator/interface.go -destination=./internal/mocks/aggregator.go -package=mocks
	mockgen -source=./pkg/pdp/curio/client.go -destination=./internal/mocks/curio_client.go -package=mocks
	mockgen -source=./internal/ipldstore/ipldstore.go -destination=./internal/mocks/ipldstore.go -package=mocks
	mockgen -source=./pkg/pdp/aggregator/steps.go -destination=./internal/mocks/steps.go -package=mocks
	mockgen -destination=./internal/mocks/sender_eth_client.go -package=mocks github.com/storacha/piri/pkg/pdp/tasks SenderETHClient
	mockgen -destination=./internal/mocks/message_watcher_eth_client.go -package=mocks github.com/storacha/piri/pkg/pdp/tasks MessageWatcherEthClient
	mockgen -destination=./internal/mocks/contract_backend.go -package=mocks github.com/ethereum/go-ethereum/accounts/abi/bind ContractBackend
	mockgen -destination=./internal/mocks/pdp.go -package=mocks github.com/storacha/piri/pkg/pdp/service/contract PDP
	mockgen -destination=./internal/mocks/pdp_proving_schedule.go -package=mocks github.com/storacha/piri/pkg/pdp/service/contract PDPProvingSchedule
	mockgen -destination=./internal/mocks/pdp_verifier.go -package=mocks github.com/storacha/piri/pkg/pdp/service/contract PDPVerifier


# special target that sets the calibnet tag and invokes build
calibnet: TAGS=-tags calibnet
calibnet: build
