# GitHub Reporting Tool Makefile

.PHONY: build test clean install run help

# Default target
help:
	@echo "GitHub Reporting Tool"
	@echo ""
	@echo "Available targets:"
	@echo "  build    - Build the application"
	@echo "  test     - Run all tests"
	@echo "  clean    - Remove build artifacts"
	@echo "  install  - Install dependencies"
	@echo "  run      - Run with example parameters"
	@echo "  help     - Show this help message"

# Build the application
build:
	@echo "Building ghreporting..."
	go mod tidy
	go build -o ghreporting .

# Run all tests
test:
	@echo "Running tests..."
	go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f ghreporting
	go clean

# Install/update dependencies
install:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Example run (requires GITHUB_TOKEN env var)
run:
	@echo "Running example (requires GITHUB_TOKEN environment variable)..."
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo "Warning: GITHUB_TOKEN not set. Using unauthenticated requests (limited rate)."; \
	fi
	./ghreporting -target octocat -since 2024-10-01 -until 2024-10-28

# Build and test
all: install build test

# Install for production use
prod-install: build
	@echo "Installing to /usr/local/bin (requires sudo)..."
	sudo cp ghreporting /usr/local/bin/
	@echo "ghreporting installed successfully!"
	@echo "You can now run 'ghreporting' from anywhere."