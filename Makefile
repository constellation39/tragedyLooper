# Makefile for the Tragedy Looper project

# Paths
GO_BIN_PATH := C:/Users/const/scoop/apps/go/current/bin
GIT_BIN_PATH := C:/Users/const/scoop/apps/git/current/usr/bin
SHELL_PREFIX := export PATH="$(GIT_BIN_PATH):$(GO_BIN_PATH):$$PATH" &&

# Binary name
BINARY_NAME=tragedylooper

.PHONY: all build run test clean lint proto clean-proto install-tools format

all: build

# Format the code
format:
	@echo "Formatting..."
	@$(SHELL_PREFIX) go run github.com/bufbuild/buf/cmd/buf@latest format -w
	@$(SHELL_PREFIX) goimports -w .

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@$(SHELL_PREFIX) go build -o bin/$(BINARY_NAME) ./cmd/tragedylooper

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	@$(SHELL_PREFIX) go run ./cmd/tragedylooper

# Build the test client
build-test-client:
	@echo "Building test client..."
	@$(SHELL_PREFIX) go build -o bin/testclient ./cmd/testclient

# Run the test client
run-test-client:
	@echo "Running test client..."
	@$(SHELL_PREFIX) go run ./cmd/testclient

# Test the application
test:
	@echo "Running tests..."
	@$(SHELL_PREFIX) go test ./...

# Clean the binary
clean:
	@echo "Cleaning..."
	@$(SHELL_PREFIX) go run ./tools/rmrf bin

# Validate data files
validate-data:
	@echo "Validating data files..."
	@$(SHELL_PREFIX) go run ./tools/autovalidator/main.go

# Lint the code
lint: format
	@echo "Linting..."
	@$(SHELL_PREFIX) golangci-lint run
	@$(SHELL_PREFIX) go run github.com/bufbuild/buf/cmd/buf@latest lint

# Install protobuf tools
install-tools:
	@echo "Installing protobuf tools..."
	@$(SHELL_PREFIX) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@$(SHELL_PREFIX) go install github.com/chrusty/protoc-gen-jsonschema/cmd/protoc-gen-jsonschema@latest

# Protobuf generation
proto:
	@echo "Generating Go code and JSON schema from protobuf..."
	@$(SHELL_PREFIX) buf generate

# Clean generated protobuf files
clean-proto:
	@echo "Cleaning generated protobuf files..."
	@$(SHELL_PREFIX) go run ./tools/rmrf internal/game/model
	@$(SHELL_PREFIX) go run ./tools/rmrf data/jsonschema
