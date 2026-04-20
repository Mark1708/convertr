BIN     := convertr
MODULE  := github.com/Mark1708/convertr
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

.PHONY: build install test lint clean snapshot release

build:
	go build $(LDFLAGS) -o $(BIN) ./cmd/convertr

install:
	go install $(LDFLAGS) ./cmd/convertr

test:
	go test ./...

lint:
	golangci-lint run ./...

snapshot:
	goreleaser release --snapshot --clean

release:
	goreleaser release --clean

clean:
	rm -f $(BIN)
