# Makefile for the Tragedy Looper project

# Binary name
BINARY_NAME=tragedylooper
# Go command
GO := go
# Protobuf files
PROTO_FILES := $(shell find proto -name *.proto)


.PHONY: all build run test clean lint proto clean-proto install-tools format validate-cue

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

# Validate CUE files
validate-cue:
	@echo "Validating CUE files..."
	@$(GO) run cuelang.org/go/cmd/cue@latest vet ./data/...

# Lint the code
lint: format validate-cue
	@echo "Linting..."
	@$(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

# Install tools
install-tools:
	@echo "Installing tools..."
	@$(GO) install golang.org/x/tools/cmd/goimports@latest
	@$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@$(GO) install cuelang.org/go/cmd/cue@latest
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Protobuf generation
gen: gen-go gen-cue

gen-go: clean-proto
	@echo "Generating Go code from protobuf..."
	@mkdir -p pkg/proto
	@protoc --proto_path=proto \
		--go_out=pkg/proto --go_opt=paths=source_relative \
		$(PROTO_FILES)

gen-cue:
	@echo "Generating CUE files from Go..."
	@$(GO) run cuelang.org/go/cmd/cue@latest get go github.com/constellation39/tragedyLooper/pkg/proto/tragedylooper/v1/...

# Clean generated protobuf files
clean-proto:
	@echo "Cleaning generated protobuf files..."
	@$(GO) run ./tools/rmrf pkg/proto cue.mod/gen
