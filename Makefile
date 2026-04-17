BIN     := convertr
MODULE  := git.mark1708.ru/me/convertr
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

.PHONY: build install test lint clean

build:
	go build $(LDFLAGS) -o $(BIN) ./cmd/convertr

install:
	go install $(LDFLAGS) ./cmd/convertr

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BIN)
