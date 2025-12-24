.PHONY: build build-linux build-windows clean run serve dump test tidy help

# Build variables
BINARY_NAME := hazelmere
BUILD_DIR := bin
CMD_PATH := ./src/cmd/hazelmere

# Add .exe extension on Windows
ifeq ($(OS),Windows_NT)
	EXT := .exe
else
	EXT :=
endif

# Build info (injected at build time)
COMMIT := $(or $(RAILWAY_GIT_COMMIT_SHA),$(shell git rev-parse --short HEAD 2>/dev/null))
BUILD_TIME := $(shell powershell -Command "Get-Date -UFormat '%Y-%m-%dT%H:%M:%SZ'" 2>/dev/null || date -u '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null)

# Go build flags
VERSION_PKG := github.com/ctfloyd/hazelmere-api/src/internal/version
LDFLAGS := -ldflags "-X $(VERSION_PKG).Commit=$(COMMIT) -X $(VERSION_PKG).BuildTime=$(BUILD_TIME)"

# Default target
all: build

build:
	$(info Building $(BINARY_NAME)...)
	$(info   Commit:     $(COMMIT))
	$(info   Build Time: $(BUILD_TIME))
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)$(EXT) $(CMD_PATH)
	$(info Built: $(BUILD_DIR)/$(BINARY_NAME)$(EXT))

build-linux:
	$(info Building $(BINARY_NAME) for Linux...)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	$(info Built: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64)

build-windows:
	$(info Building $(BINARY_NAME) for Windows...)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)
	$(info Built: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe)

build-all: build build-linux build-windows

clean:
	$(info Cleaning...)
	go clean
	-rm -rf $(BUILD_DIR)
	-rmdir /s /q $(BUILD_DIR)
	$(info Done)

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)$(EXT) serve

serve: run

dump: build
	./$(BUILD_DIR)/$(BINARY_NAME)$(EXT) dump

test:
	go test ./...

tidy:
	go mod tidy

help:
	$(info Hazelmere API Makefile)
	$(info )
	$(info Usage:)
	$(info   make build          Build for current platform)
	$(info   make build-linux    Build for Linux amd64)
	$(info   make build-windows  Build for Windows amd64)
	$(info   make build-all      Build for all platforms)
	$(info   make clean          Remove build artifacts)
	$(info   make run            Build and run the API server)
	$(info   make serve          Alias for make run)
	$(info   make dump           Build and run database dump)
	$(info   make test           Run tests)
	$(info   make tidy           Run go mod tidy)
	$(info   make help           Show this help)
	$(info )
	$(info Variables:)
	$(info   COMMIT=$(COMMIT))
	$(info   BUILD_TIME=$(BUILD_TIME))
