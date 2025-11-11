.PHONY: build run clean install test help

# Binary name
BINARY_NAME=inventory-manager
BINARY_PATH=bin/$(BINARY_NAME)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build the application
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p bin
	@$(GOBUILD) -o $(BINARY_PATH) ./cmd/inventory-manager/
	@echo "✅ Build complete: $(BINARY_PATH)"

# Run the application
run: build
	@echo "🚀 Running $(BINARY_NAME)..."
	@./$(BINARY_PATH)

# Clean build artifacts
clean:
	@echo "🗑️  Cleaning..."
	@$(GOCLEAN)
	@rm -rf bin/
	@echo "✅ Clean complete"

# Install to system
install: build
	@echo "📦 Installing to /usr/local/bin..."
	@sudo cp $(BINARY_PATH) /usr/local/bin/$(BINARY_NAME)
	@echo "✅ Installed: /usr/local/bin/$(BINARY_NAME)"

# Run tests
test:
	@echo "🧪 Running tests..."
	@$(GOTEST) -v ./...

# Update dependencies
deps:
	@echo "📦 Updating dependencies..."
	@$(GOGET) -u ./...
	@$(GOMOD) tidy
	@echo "✅ Dependencies updated"

# Format code
fmt:
	@echo "🎨 Formatting code..."
	@gofmt -s -w .
	@echo "✅ Code formatted"

# Lint code
lint:
	@echo "🔍 Linting code..."
	@golangci-lint run ./... 2>/dev/null || echo "⚠️  golangci-lint not installed"

# Show help
help:
	@echo "Available commands:"
	@echo "  make build    - Build the application"
	@echo "  make run      - Build and run the application"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make install  - Install to /usr/local/bin"
	@echo "  make test     - Run tests"
	@echo "  make deps     - Update dependencies"
	@echo "  make fmt      - Format code"
	@echo "  make lint     - Lint code"
	@echo "  make help     - Show this help message"

# Default target
.DEFAULT_GOAL := build
