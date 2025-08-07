# PDU Parser Makefile for cross-platform builds
# Supports macOS, Windows, and Linux on x86-64 and ARM64

# Application info
APP_NAME := pdu-parser
MONTHLY_FILLER := monthly-filler
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build directory
BUILD_DIR := bin
DIST_DIR := dist

# Go build flags
LDFLAGS := -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)
BUILD_FLAGS := -ldflags "$(LDFLAGS)" -trimpath

# Platform targets
PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64 \
	linux/amd64 \
	linux/arm64

.PHONY: all build build-all clean help install deps test fmt vet lint

# Default target
all: build

# Help target
help:
	@echo "PDU Parser Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build        - Build for current platform"
	@echo "  build-all    - Cross-compile for all platforms"
	@echo "  clean        - Remove build artifacts"
	@echo "  install      - Install to GOPATH/bin"
	@echo "  deps         - Download dependencies"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format source code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run golangci-lint (if available)"
	@echo ""
	@echo "Platform-specific builds:"
	@echo "  build-darwin-amd64    - Build for macOS Intel"
	@echo "  build-darwin-arm64    - Build for macOS Apple Silicon"
	@echo "  build-windows-amd64   - Build for Windows x64"
	@echo "  build-windows-arm64   - Build for Windows ARM64"
	@echo "  build-linux-amd64     - Build for Linux x64"
	@echo "  build-linux-arm64     - Build for Linux ARM64"

# Build for current platform
build: deps $(BUILD_DIR)
	@echo "Building $(APP_NAME) for current platform..."
	go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) main.go
	@echo "Building $(MONTHLY_FILLER) for current platform..."
	go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(MONTHLY_FILLER) pkg/report/report.go
	@echo "Build complete: $(BUILD_DIR)/"

# Cross-compile for all platforms
build-all: deps $(DIST_DIR)
	@echo "Cross-compiling for all platforms..."
	@$(foreach platform,$(PLATFORMS), \
		$(call build_platform,$(platform)); \
	)
	@echo "All builds complete in $(DIST_DIR)/"

# Function to build for specific platform
define build_platform
	$(eval GOOS := $(word 1,$(subst /, ,$1)))
	$(eval GOARCH := $(word 2,$(subst /, ,$1)))
	$(eval EXT := $(if $(filter windows,$(GOOS)),.exe,))
	$(eval PLATFORM_DIR := $(DIST_DIR)/$(GOOS)-$(GOARCH))
	
	@echo "Building for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(PLATFORM_DIR)
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) \
		-o $(PLATFORM_DIR)/$(APP_NAME)$(EXT) main.go
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) \
		-o $(PLATFORM_DIR)/$(MONTHLY_FILLER)$(EXT) pkg/report/report.go
endef

# Platform-specific build targets
build-darwin-amd64: deps $(DIST_DIR)
	$(call build_platform,darwin/amd64)

build-darwin-arm64: deps $(DIST_DIR)
	$(call build_platform,darwin/arm64)

build-windows-amd64: deps $(DIST_DIR)
	$(call build_platform,windows/amd64)

build-windows-arm64: deps $(DIST_DIR)
	$(call build_platform,windows/arm64)

build-linux-amd64: deps $(DIST_DIR)
	$(call build_platform,linux/amd64)

build-linux-arm64: deps $(DIST_DIR)
	$(call build_platform,linux/arm64)

# Create build directories
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(DIST_DIR):
	mkdir -p $(DIST_DIR)

# Install to GOPATH/bin
install: deps
	@echo "Installing $(APP_NAME) to GOPATH/bin..."
	go install $(BUILD_FLAGS) .
	@echo "Installing $(MONTHLY_FILLER) to GOPATH/bin..."
	go install $(BUILD_FLAGS) ./pkg/report

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Run tests (when tests are added)
test:
	@echo "Running tests..."
	go test -v ./...

# Format source code
fmt:
	@echo "Formatting source code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run golangci-lint if available
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	go clean

# Development helpers
run-parser:
	@echo "Running PDU parser with C3.xlsx..."
	go run main.go C3.xlsx

run-filler:
	@echo "Running monthly filler with total_a1.csv..."
	go run pkg/report/report.go total_a1.csv

# Package releases
package: build-all
	@echo "Creating release packages..."
	@cd $(DIST_DIR) && \
	for dir in */; do \
		platform=$${dir%/}; \
		echo "Packaging $$platform..."; \
		if [[ "$$platform" == *"windows"* ]]; then \
			zip -r "../$${platform}.zip" "$$platform/"; \
		else \
			tar -czf "../$${platform}.tar.gz" "$$platform/"; \
		fi; \
	done
	@echo "Packages created in $(DIST_DIR)/"

# Show build info
info:
	@echo "Build Information:"
	@echo "  App Name: $(APP_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Go Version: $(shell go version)"
	@echo "  Platforms: $(PLATFORMS)"