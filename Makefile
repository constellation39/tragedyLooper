# Makefile for the Tragedy Looper project

# Binary name
BINARY_NAME=tragedylooper
# Go command
GO := go

.PHONY: all build run test clean lint proto clean-proto install-tools format

all: build

# Format the code
format:
	@echo "Formatting..."
	@$(GO) run github.com/bufbuild/buf/cmd/buf@latest format -w
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
	@$(GO) test ./...

# Clean the binary
clean:
	@echo "Cleaning..."
	@if exist bin ( rmdir /S /Q bin )

# Lint the code
lint: format
	@echo "Linting..."
	@$(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run
	@$(GO) run github.com/bufbuild/buf/cmd/buf@latest lint

# Install tools
install-tools:
	@echo "Installing tools..."
	@$(GO) install golang.org/x/tools/cmd/goimports@latest
	@$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@$(GO) install cuelang.org/go/cmd/cue@latest
	@$(GO) install github.com/bufbuild/buf/cmd/buf@latest
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Protobuf generation
proto:
	@echo "Generating Go code from protobuf..."
	@$(GO) run github.com/bufbuild/buf/cmd/buf@latest generate
	@$(GO) run cuelang.org/go/cmd/cue@latest get go github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1

# Clean generated protobuf files
clean-proto:
	@echo "Cleaning generated protobuf files..."
	@if exist pkg\proto ( rmdir /S /Q pkg\proto )
