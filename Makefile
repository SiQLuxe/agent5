BINARY := build/agent
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

.PHONY: build run test clean

build:
	go build -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o $(BINARY) ./cmd/agent
	@echo "Built: $(BINARY)  (version: $(VERSION))"

run: build
	$(BINARY)

test:
	go test -count=1 -short ./...
	go test -race -count=1 -short ./...

clean:
	rm -rf build/
