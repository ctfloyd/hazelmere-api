.PHONY: build build-linux build-windows clean run serve dump help

# Build variables
BINARY_NAME := hazelmere
BUILD_DIR := bin
CMD_PATH := ./src/cmd/hazelmere

# Version info (injected at build time)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Go build flags
VERSION_PKG := github.com/ctfloyd/hazelmere-api/src/internal/version
LDFLAGS := -ldflags "-X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).Commit=$(COMMIT) -X $(VERSION_PKG).BuildTime=$(BUILD_TIME)"

# Default target
all: build

# Build for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@echo "  Version:    $(VERSION)"
	@echo "  Commit:     $(COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for Linux
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

# Build for Windows
build-windows:
	@echo "Building $(BINARY_NAME) for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe"

# Build all platforms
build-all: build build-linux build-windows

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	@echo "Done"

# Run the API server (development)
run: build
	./$(BUILD_DIR)/$(BINARY_NAME) serve

serve: run

# Run database dump
dump: build
	./$(BUILD_DIR)/$(BINARY_NAME) dump

# Run tests
test:
	go test ./...

# Run go mod tidy
tidy:
	go mod tidy

# Show help
help:
	@echo "Hazelmere API Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build          Build for current platform"
	@echo "  make build-linux    Build for Linux (amd64)"
	@echo "  make build-windows  Build for Windows (amd64)"
	@echo "  make build-all      Build for all platforms"
	@echo "  make clean          Remove build artifacts"
	@echo "  make run            Build and run the API server"
	@echo "  make serve          Alias for 'make run'"
	@echo "  make dump           Build and run database dump"
	@echo "  make test           Run tests"
	@echo "  make tidy           Run go mod tidy"
	@echo "  make help           Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  COMMIT=$(COMMIT)"
	@echo "  BUILD_TIME=$(BUILD_TIME)"
