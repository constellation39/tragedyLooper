# Makefile for the Tragedy Looper project

# Go parameters
GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor
GOFILES := $(wildcard *.go)

# Binary name
BINARY_NAME=tragedylooper

.PHONY: all build run test clean lint

all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./cmd/tragedylooper

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	@go run ./cmd/tragedylooper

# Test the application
test:
	@echo "Running tests..."
	@go test ./...

# Clean the binary
clean:
	@echo "Cleaning..."
	@if [ -f bin/$(BINARY_NAME) ]; then rm bin/$(BINARY_NAME); fi

# Lint the code
lint:
	@echo "Linting..."
	@golangci-lint run
