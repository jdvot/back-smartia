.PHONY: help build test test-coverage lint clean docker-build docker-run docker-stop generate-client generate-swagger run-dev

# Default target
help:
	@echo "SmartDoc AI Backend - Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  run-dev          - Run the server in development mode"
	@echo "  test             - Run all tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  lint             - Run linter"
	@echo ""
	@echo "Build:"
	@echo "  build            - Build the application"
	@echo "  clean            - Clean build artifacts"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run Docker container"
	@echo "  docker-stop      - Stop Docker container"
	@echo ""
	@echo "Documentation:"
	@echo "  generate-swagger - Generate Swagger documentation"
	@echo "  generate-client  - Generate Flutter API client"
	@echo ""
	@echo "CI/CD:"
	@echo "  ci-test          - Run CI tests"
	@echo "  ci-build         - Run CI build"
	@echo "  ci-deploy        - Run CI deployment"

# Development
run-dev:
	@echo "🚀 Starting development server..."
	@ENV=development STORAGE_TYPE=local go run ./cmd/server

# Testing
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

test-coverage:
	@echo "🧪 Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report generated: coverage.html"

lint:
	@echo "🔍 Running linter..."
	@go install golang.org/x/lint/golint@latest
	@golint ./...

# Build
build:
	@echo "🔨 Building application..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server ./cmd/server

clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -rf generated/
	@rm -f flutter-client.tar.gz

# Docker
docker-build:
	@echo "🐳 Building Docker image..."
	@docker-compose build

docker-run:
	@echo "🐳 Starting Docker container..."
	@docker-compose up -d

docker-stop:
	@echo "🐳 Stopping Docker container..."
	@docker-compose down

# Documentation
generate-swagger:
	@echo "📚 Generating Swagger documentation..."
	@if command -v swag > /dev/null; then \
		swag init -g cmd/server/main.go -o docs; \
		echo "✅ Swagger documentation generated"; \
	else \
		echo "❌ swag not found. Installing..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		swag init -g cmd/server/main.go -o docs; \
		echo "✅ Swagger documentation generated"; \
	fi

generate-client:
	@echo "📱 Generating Flutter API client..."
	@chmod +x scripts/generate_flutter_client.sh
	@./scripts/generate_flutter_client.sh

# CI/CD
ci-test:
	@echo "🔧 Running CI tests..."
	@go mod download
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

ci-build:
	@echo "🔧 Running CI build..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server ./cmd/server

ci-deploy:
	@echo "🚀 Running CI deployment..."
	@echo "Deployment steps would go here..."

# Security
security-scan:
	@echo "🔒 Running security scan..."
	@if command -v trivy > /dev/null; then \
		trivy fs .; \
	else \
		echo "❌ Trivy not found. Please install Trivy for security scanning."; \
	fi

# Database
db-migrate:
	@echo "🗄️ Running database migrations..."
	@echo "Database migration steps would go here..."

db-seed:
	@echo "🌱 Seeding database..."
	@echo "Database seeding steps would go here..."

# Monitoring
health-check:
	@echo "🏥 Running health check..."
	@curl -f http://localhost:8080/health || echo "❌ Health check failed"

# Performance
benchmark:
	@echo "⚡ Running benchmarks..."
	@go test -bench=. ./...

# Dependencies
deps-update:
	@echo "📦 Updating dependencies..."
	@go get -u ./...
	@go mod tidy

deps-check:
	@echo "🔍 Checking for security vulnerabilities..."
	@go list -json -deps ./... | nancy sleuth

# Development setup
setup-dev:
	@echo "🛠️ Setting up development environment..."
	@cp env.example .env
	@echo "✅ Development environment setup complete"
	@echo "📝 Please edit .env file with your configuration"

# Production
build-prod:
	@echo "🏭 Building for production..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/server ./cmd/server

# Backup
backup:
	@echo "💾 Creating backup..."
	@tar -czf backup-$(shell date +%Y%m%d-%H%M%S).tar.gz --exclude=node_modules --exclude=.git .

# Release
release:
	@echo "🎉 Creating release..."
	@echo "Current version: $(shell git describe --tags --abbrev=0 2>/dev/null || echo 'v1.0.0')"
	@echo "Please create a new tag: git tag v1.x.x && git push origin v1.x.x" 