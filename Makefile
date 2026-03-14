GO ?= go
GORELEASER ?= goreleaser
BINARY ?= envcrypt

.PHONY: build test release-snapshot clean

build:
	$(GO) build -o $(BINARY) .

test:
	$(GO) test ./...

release-snapshot:
	$(GORELEASER) release --snapshot --clean

clean:
	rm -rf dist $(BINARY)
