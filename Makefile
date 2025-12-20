# Makefile for NexFlux Virtual Lab Backend

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run
SWAG=~/go/bin/swag

# Binary name
BINARY_NAME=nexflux-backend

# --- Main Commands ---

# Default command
all: help

# Run the application
run:
	@echo "🚀 Starting NexFlux Virtual Lab API..."
	$(GORUN) main.go

# Build the application
build:
	@echo "📦 Building NexFlux backend..."
	$(GOBUILD) -o $(BINARY_NAME) main.go

# Run tests
test:
	@echo "🧪 Running tests..."
	$(GOTEST) -v ./...

# Generate Swagger documentation
swagger:
	@echo "📚 Generating Swagger docs..."
	$(SWAG) init --parseDependency --parseInternal
	@echo "✅ Swagger docs generated at docs/"

# Run with hot reload using reflex
dev:
	@echo "🔥 Starting with hot reload..."
	reflex -r '\.go$$' -s -- sh -c '$(GORUN) main.go'

# Clean up build artifacts
clean:
	@echo "🧹 Cleaning up..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Install dependencies
deps:
	@echo "📥 Installing dependencies..."
	$(GOCMD) mod download
	$(GOCMD) install github.com/swaggo/swag/cmd/swag@latest

# --- Helper Commands ---

# Display help
help:
	@echo "NexFlux Virtual Lab Backend - Available commands:"
	@echo ""
	@echo "  make run      - Start the application"
	@echo "  make build    - Build the application binary"
	@echo "  make test     - Run all tests"
	@echo "  make swagger  - Generate Swagger API documentation"
	@echo "  make dev      - Run with hot reload (requires reflex)"
	@echo "  make deps     - Install dependencies"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make help     - Display this help message"

.PHONY: all run build test swagger dev clean deps help

