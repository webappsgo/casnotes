# casnotes Build System per CLAUDE.md Build System

.PHONY: build test clean run install docker release help

# Build variables per CLAUDE.md
APP_NAME := casnotes
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "1.0.0")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +%Y%m%d-%H%M%S)

# Build flags per CLAUDE.md
LDFLAGS := -ldflags "-s -w -X 'main.Version=$(VERSION)' -X 'main.Commit=$(COMMIT)' -X 'main.BuildTime=$(BUILD_TIME)'"

# Platform targets per CLAUDE.md
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

# Directories
BUILD_DIR := build
DIST_DIR := dist

# Default target
help:
	@echo "casnotes Build System per CLAUDE.md"
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@echo "  build     Build all platforms + host binary"
	@echo "  test      Run all tests with coverage"  
	@echo "  clean     Remove build artifacts"
	@echo "  run       Run locally for development"
	@echo "  install   Install host binary to /usr/local/bin"
	@echo "  docker    Build and push to ghcr.io"
	@echo "  release   Create GitHub release"
	@echo ""
	@echo "Current settings:"
	@echo "  Version:    $(VERSION)"
	@echo "  Commit:     $(COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"

# Build host binary per CLAUDE.md
build-host:
	@echo "Building $(APP_NAME) for host platform..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build $(LDFLAGS) -a -installsuffix cgo -o $(BUILD_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)
	@echo "✅ Built: $(BUILD_DIR)/$(APP_NAME)"

# Build all platforms per CLAUDE.md
build: clean build-host
	@echo "Building $(APP_NAME) v$(VERSION) for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		output_name=$(APP_NAME)-$$GOOS-$$GOARCH; \
		if [ "$$GOOS" = "windows" ]; then \
			output_name=$$output_name.exe; \
		fi; \
		echo "Building $$platform..."; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH go build $(LDFLAGS) -a -installsuffix cgo -o $(BUILD_DIR)/$$output_name ./cmd/$(APP_NAME); \
		if [ $$? -eq 0 ]; then \
			echo "✅ Built: $(BUILD_DIR)/$$output_name"; \
		else \
			echo "❌ Failed to build $$platform"; \
		fi; \
	done
	
	@echo "🎉 Build complete! Binaries in $(BUILD_DIR)/"

# Run tests per CLAUDE.md
test:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Tests complete. Coverage report: coverage.html"

# Run locally for development per CLAUDE.md
run: build-host
	@echo "Starting $(APP_NAME) in development mode..."
	@DEBUG=true DATA_DIR=./data PORT=64123 ./$(BUILD_DIR)/$(APP_NAME) --debug

# Install to system per CLAUDE.md
install: build-host
	@echo "Installing $(APP_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(APP_NAME)
	@echo "✅ Installed: /usr/local/bin/$(APP_NAME)"

# Clean build artifacts per CLAUDE.md
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR) coverage.out coverage.html
	@go clean -cache -testcache
	@echo "✅ Clean complete"

# Docker build per CLAUDE.md
docker:
	@echo "Building Docker image per CLAUDE.md..."
	@docker build -t ghcr.io/casapps/$(APP_NAME):latest .
	@docker tag ghcr.io/casapps/$(APP_NAME):latest ghcr.io/casapps/$(APP_NAME):$(VERSION)
	@echo "✅ Docker image: ghcr.io/casapps/$(APP_NAME):$(VERSION)"

# GitHub release per CLAUDE.md
release: build
	@echo "Creating GitHub release $(VERSION)..."
	@mkdir -p $(DIST_DIR)
	
	@for file in $(BUILD_DIR)/$(APP_NAME)-*; do \
		if [ -f "$$file" ]; then \
			basename_file=$$(basename $$file); \
			echo "Creating archive for $$basename_file..."; \
			tar -czf $(DIST_DIR)/$$basename_file.tar.gz -C $(BUILD_DIR) $$basename_file; \
		fi \
	done
	
	@cd $(DIST_DIR) && sha256sum *.tar.gz > checksums.txt
	@echo "✅ Release artifacts in $(DIST_DIR)/"

# Development tools
dev:
	@echo "Development mode - watching for changes..."
	@DATA_DIR=./data PORT=64127 ./$(BUILD_DIR)/$(APP_NAME) --debug

format:
	@go fmt ./...
	@echo "✅ Code formatted"

mod-tidy:
	@go mod tidy
	@echo "✅ Modules tidied"