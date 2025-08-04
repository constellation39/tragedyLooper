# Makefile for the Tragedy Looper project

# Binary name
BINARY_NAME=tragedylooper
# Go command
GO := go
# Schema directory
SCHEMA_DIR=schemas

.PHONY: all build run test clean lint proto clean-proto install-tools format

all: build

# Format the code
format:
	@echo "Formatting..."
	@$(GO) fmt ./...
	@$(GO) run golang.org/x/tools/cmd/goimports@latest -w cmd internal pkg tools

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@$(GO) build -o bin/$(BINARY_NAME) ./cmd/tragedylooper

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	@$(GO) run ./cmd/tragedylooper

# Test the application
test:
	@echo "Running tests..."
	@$(GO) test ./... -timeout 1m

# Clean the binary
clean:
	@echo "Cleaning..."
	@$(GO) run ./tools/rmrf bin

# Lint the code
lint: format
	@echo "Linting..."
	@$(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

# Validate the project
validate:
	@echo "Validating..."
	@$(GO) run ./tools/autovalidator/main.go

# Install tools
install-tools:
	@echo "Installing tools..."
	@$(GO) install golang.org/x/tools/cmd/goimports@latest
	@$(GO) install github.com/bufbuild/buf-cli/cmd/buf@latest
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install github.com/chrusty/protoc-gen-jsonschema/cmd/protoc-gen-jsonschema@latest

# Protobuf generation
gen:
	@echo "Generating Go code and JSON schema from protobuf..."
	@buf generate

# Clean generated protobuf files
clean-proto:
	@echo "Cleaning generated protobuf files..."
	@$(GO) run ./tools/rmrf pkg/proto
	@$(GO) run ./tools/rmrf $(SCHEMA_DIR)
