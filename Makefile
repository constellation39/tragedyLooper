# Makefile for the Tragedy Looper project

# Binary name
BINARY_NAME=tragedylooper
PROTO_FILES := $(wildcard proto/model/*.proto)

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
	@go run ./tools/rmrf bin

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
GOGEN_OUT_DIR := internal/game
JSONSCHEMA_OUT_DIR := data/jsonschema

proto: install-tools
	@echo "Generating Go code and JSON schema from protobuf..."
	@go run ./tools/mkdir $(GOGEN_OUT_DIR)
	@go run ./tools/mkdir $(JSONSCHEMA_OUT_DIR)
	@protoc --proto_path=. --go_out=$(GOGEN_OUT_DIR) --go_opt=paths=source_relative $(PROTO_FILES)
	@protoc --proto_path=. --jsonschema_out=$(JSONSCHEMA_OUT_DIR) $(PROTO_FILES)

# Clean generated protobuf files
clean-proto:
	@echo "Cleaning generated protobuf files..."
	@go run ./tools/rmrf $(GOGEN_OUT_DIR)
	@go run ./tools/rmrf $(JSONSCHEMA_OUT_DIR)