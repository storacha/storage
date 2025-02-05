storage:
	go build -o ./storage ./cmd/storage

.PHONY: install

install:
	go install ./cmd/storage

.PHONY: clean

clean:
	rm -f ./storage
