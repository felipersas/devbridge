VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build install test clean

build:
	go build $(LDFLAGS) -o devbridge ./cmd/devbridge

install: build
	cp devbridge /usr/local/bin/

test:
	go test -race ./...

clean:
	rm -f devbridge
