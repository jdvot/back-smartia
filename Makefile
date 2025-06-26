.PHONY: help generate build run test clean deploy-railway deploy-render

# Default target
help:
	@echo "SmartDoc AI - Available commands:"
	@echo "  generate      - Generate Go code from OpenAPI specification"
	@echo "  build         - Build the application"
	@echo "  run           - Run the server locally"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  deploy-railway - Deploy to Railway"
	@echo "  deploy-render - Deploy to Render"

# Generate Go code from OpenAPI specification
generate:
	@echo "Generating Go code from OpenAPI specification..."
	oapi-codegen -config oapi-codegen.yaml openapi.yaml
	@echo "Code generation complete!"

# Build the application
build:
	@echo "Building SmartDoc AI..."
	go build -o bin/smartdoc-ai cmd/server/main.go
	@echo "Build complete! Binary: bin/smartdoc-ai"

# Run the server locally
run:
	@echo "Starting SmartDoc AI server..."
	go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Deploy to Railway
deploy-railway:
	@echo "Deploying to Railway..."
	railway up

# Deploy to Render
deploy-render:
	@echo "Deploying to Render..."
	@echo "Please deploy manually through Render dashboard"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Generate Swagger docs
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/server/main.go

# Development setup
dev-setup: deps generate
	@echo "Development setup complete!"
	@echo "Next steps:"
	@echo "1. Copy env.example to .env and configure your settings"
	@echo "2. Run 'make run' to start the server"
	@echo "3. Visit http://localhost:8080/swagger/ for API documentation" 