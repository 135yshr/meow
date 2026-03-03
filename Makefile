.PHONY: build test lint vet generate clean install wasm help

## build: Build the meow compiler
build:
	go build -o meow ./cmd/meow

## install: Install meow to $GOPATH/bin
install:
	go install ./cmd/meow

## test: Run all tests
test:
	go test ./...

## test-v: Run all tests with verbose output
test-v:
	go test ./... -v

## test-update: Update golden files
test-update:
	go test ./compiler/ -update

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## vet: Run go vet
vet:
	go vet ./...

## generate: Run go generate (requires stringer)
generate:
	go install golang.org/x/tools/cmd/stringer@latest
	go generate ./...

## wasm: Build WASM for Playground
wasm:
	GOOS=js GOARCH=wasm go build -o playground/meow.wasm ./cmd/playground/

## cover: Run tests with coverage report
cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Open coverage.html to view the report"

## clean: Remove build artifacts
clean:
	rm -f meow playground-server playground/meow.wasm
	rm -f coverage.out coverage.html

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
