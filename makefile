BINARY_NAME ?= biathlon
CMD_PATH    := ./cmd/biathlon

.DEFAULT_GOAL := build

.PHONY: fmt build run clean help

fmt:
	go fmt ./...

build: fmt
	go build -o $(BINARY_NAME) $(CMD_PATH)

run: build
	go run $(CMD_PATH)

clean:
	go clean -cache
	rm -f $(BINARY_NAME)

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-10s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Variables (can be overridden, e.g., make build BINARY_NAME=my_app):"
	@echo "  BINARY_NAME   : Name of the output binary (default: $(BINARY_NAME))"