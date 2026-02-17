BINARY  := kko
VERSION := $(shell git describe --tags --always 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build install test lint clean release

build:
	go build $(LDFLAGS) -o $(BINARY) .

install:
	go install $(LDFLAGS) .

test:
	go test ./... -race -count=1

lint:
	golangci-lint run

clean:
	rm -f $(BINARY)

release:
	goreleaser release --clean
