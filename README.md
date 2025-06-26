# SmartDoc AI Backend

🚀 **RESTful API for document processing with OCR and AI summarization**

[![CI/CD](https://github.com/your-username/back-smartia/workflows/CI/CD%20Pipeline/badge.svg)](https://github.com/your-username/back-smartia/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/your-username/back-smartia)](https://goreportcard.com/report/github.com/your-username/back-smartia)
[![Coverage](https://codecov.io/gh/your-username/back-smartia/branch/main/graph/badge.svg)](https://codecov.io/gh/your-username/back-smartia)

## 📋 Table of Contents

- [Features](#-features)
- [Quick Start](#-quick-start)
- [API Documentation](#-api-documentation)
- [Security](#-security)
- [Testing](#-testing)
- [Development](#-development)
- [CI/CD](#-cicd)
- [Flutter Client](#-flutter-client)
- [Deployment](#-deployment)
- [Contributing](#-contributing)

## ✨ Features

- 🔐 **Firebase Authentication** - Secure token-based authentication
- 📄 **Document Upload** - Support for multiple file formats
- 🔍 **OCR Processing** - Text extraction from documents
- 🤖 **AI Summarization** - Intelligent document summarization
- 📊 **Document Management** - Full CRUD operations
- 📱 **Flutter Client Generation** - Auto-generated API client
- 🐳 **Docker Support** - Containerized deployment
- 🧪 **Comprehensive Testing** - Unit tests with coverage
- 🔒 **Security Scanning** - Vulnerability detection
- 📚 **Swagger Documentation** - Interactive API docs

## 🚀 Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Node.js 18+ (for client generation)

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/your-username/back-smartia.git
cd back-smartia

# Start the application
docker-compose up --build

# The API will be available at http://localhost:8080
# Swagger documentation at http://localhost:8080/swagger/index.html
```

### Local Development

```bash
# Install dependencies
go mod download

# Set up environment
cp env.example .env
# Edit .env with your configuration

# Run the server
make run-dev
```

## 📚 API Documentation

### Interactive Documentation

Visit [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) for interactive API documentation.

### Authentication

All API endpoints (except `/health` and `/swagger/*`) require Firebase authentication:

```bash
# Include your Firebase ID token in the Authorization header
curl -H "Authorization: Bearer YOUR_FIREBASE_TOKEN" \
     http://localhost:8080/docs/history
```

### Development Mode

In development mode, you can use test tokens:

```bash
# Create a test token (base64 encoded JSON)
echo '{"user_id":"test-user-123","exp":'$(($(date +%s) + 3600))'}' | base64

# Use the test token
curl -H "Authorization: Bearer eyJ1c2VyX2lkIjoidGVzdC11c2VyLTEyMyIsImV4cCI6MTYzNTY3ODkwMH0=" \
     http://localhost:8080/docs/history
```

## 🔒 Security

### Token Validation

- **Production**: Firebase ID tokens are validated against Firebase Auth
- **Development**: Test tokens are validated locally with expiration checks
- **Bypass**: Health check and Swagger endpoints bypass authentication

### Security Features

- ✅ Token expiration validation
- ✅ User context isolation
- ✅ Secure file upload validation
- ✅ Input sanitization
- ✅ CORS configuration
- ✅ Rate limiting (configurable)

## 🧪 Testing

### Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests in Docker
docker-compose -f docker-compose.dev.yml --profile dev up --build
docker-compose -f docker-compose.dev.yml exec smartdoc-api-dev go test -v ./...
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
open coverage.html
```

### Test Structure

```
├── internal/auth/
│   └── firebase_test.go      # Authentication tests
├── cmd/server/
│   └── server_test.go        # API endpoint tests
└── internal/services/
    └── *_test.go             # Service layer tests
```

## 🛠️ Development

### Available Commands

```bash
# Development
make run-dev              # Run development server
make test                 # Run tests
make test-coverage        # Run tests with coverage
make lint                 # Run linter

# Build
make build                # Build application
make clean                # Clean artifacts

# Docker
make docker-build         # Build Docker image
make docker-run           # Run Docker container
make docker-stop          # Stop Docker container

# Documentation
make generate-swagger     # Generate Swagger docs
make generate-client      # Generate Flutter client

# CI/CD
make ci-test              # Run CI tests
make ci-build             # Run CI build
make ci-deploy            # Run CI deployment
```

### Project Structure

```
├── cmd/server/           # Application entry point
├── internal/             # Private application code
│   ├── auth/            # Authentication logic
│   ├── models/          # Data models
│   ├── services/        # Business logic
│   └── storage/         # Storage implementations
├── docs/                # Generated Swagger docs
├── generated/           # Generated code
├── scripts/             # Utility scripts
├── .github/workflows/   # CI/CD pipelines
├── openapi.yaml         # API specification
└── docker-compose.yml   # Docker configuration
```

## 🔄 CI/CD

### GitHub Actions Pipeline

The CI/CD pipeline includes:

1. **Tests** - Unit tests with coverage
2. **Security** - Vulnerability scanning with Trivy
3. **Build** - Docker image building
4. **Client Generation** - Flutter API client generation
5. **Deployment** - Staging and production deployment

### Pipeline Triggers

- **Push to `main`** - Full pipeline with production deployment
- **Push to `develop`** - Full pipeline with staging deployment
- **Pull Request** - Tests and security scanning only

### Environment Variables

Set these in your GitHub repository secrets:

```bash
FIREBASE_PROJECT_ID=your-project-id
FIREBASE_SERVICE_ACCOUNT_KEY=your-service-account-json
```

## 📱 Flutter Client

### Auto-Generated Client

The CI/CD pipeline automatically generates a Flutter API client from the OpenAPI specification.

### Manual Generation

```bash
# Generate Flutter client locally
make generate-client

# The client will be available in:
# - ./generated/flutter-client/
# - ./flutter-client.tar.gz
```

### Usage in Flutter

```dart
import 'package:smartdoc_api/api.dart';

void main() async {
  final api = SmartDocApi();
  
  // Set base URL
  api.setBasePath('https://your-api-url.com');
  
  // Set authentication token
  api.setBearerAuth('your-firebase-token');
  
  // Upload document
  final file = await MultipartFile.fromFile('path/to/document.pdf');
  final response = await api.uploadDocument(file);
  
  print('Document uploaded: ${response.data?.id}');
}
```

## 🚀 Deployment

### Docker Deployment

```bash
# Build and run
docker-compose up --build -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

### Environment Configuration

```bash
# Copy example environment
cp env.example .env

# Configure your environment
ENV=production
FIREBASE_PROJECT_ID=your-project-id
FIREBASE_SERVICE_ACCOUNT_KEY=your-service-account-json
STORAGE_TYPE=firebase  # or 'local' for development
PORT=8080
```

### Production Considerations

- ✅ Use HTTPS in production
- ✅ Configure proper CORS settings
- ✅ Set up monitoring and logging
- ✅ Use environment-specific configurations
- ✅ Implement rate limiting
- ✅ Set up backup strategies

## 🤝 Contributing

### Development Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Add tests for new functionality
5. Run tests: `make test`
6. Commit your changes: `git commit -m 'Add amazing feature'`
7. Push to the branch: `git push origin feature/amazing-feature`
8. Open a Pull Request

### Code Standards

- Follow Go coding standards
- Write comprehensive tests
- Update documentation
- Use conventional commit messages
- Ensure all tests pass

### Testing Guidelines

- Write unit tests for all new functionality
- Maintain test coverage above 80%
- Use descriptive test names
- Mock external dependencies
- Test both success and error cases

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📧 Email: support@smartdoc.ai
- 📖 Documentation: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
- 🐛 Issues: [GitHub Issues](https://github.com/your-username/back-smartia/issues)

## 🙏 Acknowledgments

- Firebase for authentication
- Google Cloud Vision for OCR
- OpenAI for AI summarization
- Swagger for API documentation
- OpenAPI Generator for client generation 