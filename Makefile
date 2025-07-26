# Makefile for the Tragedy Looper project

# Go parameters
GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor
GOFILES := $(wildcard *.go)

# Binary name
BINARY_NAME=tragedylooper

.PHONY: all build run test clean lint proto clean-proto install-tools

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

# Install protobuf tools
install-tools:
	@echo "Installing protobuf tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install github.com/chrusty/protoc-gen-jsonschema/cmd/protoc-gen-jsonschema@latest

# Protobuf generation
PROTO_FILES := $(wildcard proto/game/*.proto)

proto: install-tools
	@echo "Generating Go code and JSON schema from protobuf..."
	@protoc --go_out=. --go_opt=paths=source_relative $(PROTO_FILES)
	@protoc --jsonschema_out=./proto/game $(PROTO_FILES)

# Clean generated protobuf files
clean-proto:
	@echo "Cleaning generated protobuf files..."
	@rm -f proto/game/*.pb.go
	@rm -f proto/game/*.json
