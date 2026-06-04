BINARY := /tmp/agent-tui/agent
SANDBOX := /tmp/agent-tui/sandbox

.PHONY: build run test clean

build: clean-root-binary
	@mkdir -p /tmp/agent-tui 2>/dev/null || true
	go build -o $(BINARY) ./cmd/agent
	@echo "Built: $(BINARY)  (sandbox: $(SANDBOX))"

clean-root-binary:
	@rm -f ./agent

run: build
	SANDBOX_DIR=$(SANDBOX) $(BINARY)

test:
	go test -count=1 -short ./...
	go test -race -count=1 -short ./...

clean:
	rm -rf /tmp/agent-tui
