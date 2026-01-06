# Makefile for Library Management System

.PHONY: help build run test clean dev deps lint fmt

# Default target
help:
	@echo "Available commands:"
	@echo "  build    - Build the application"
	@echo "  run      - Run the application"
	@echo "  test     - Run all tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  deps     - Download dependencies"
	@echo "  lint     - Run linter"
	@echo "  fmt      - Format code"
	@echo "  help     - Show this help message"

# Build the application
build:
	@echo "Building application..."
	go build -o bin/librarymanagementsystem ./cmd/server

# Run the application
run:
	@echo "Running application..."
	@go run .

# Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...


