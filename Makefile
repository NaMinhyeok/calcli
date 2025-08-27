# Makefile for calcli

.PHONY: build test test-race lint fmt vet clean install dev help

# Default target
help:
	@echo "Available targets:"
	@echo "  build     - Build the binary"
	@echo "  test      - Run tests"
	@echo "  fmt       - Format code"
	@echo "  vet       - Run go vet"
	@echo "  clean     - Clean build artifacts"
	@echo "  install   - Install binary to GOPATH/bin"
	@echo "  dev       - Run fmt, vet, and test (development cycle)"

# Build the binary
build:
	go build -o bin/calcli cmd/calcli/main.go

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf bin/

install:
	go install cmd/calcli/main.go

dev: fmt vet test
	@echo "Development checks passed!"
