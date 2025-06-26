#!/bin/bash

# Script to generate Flutter API client from OpenAPI specification
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Generating Flutter API Client...${NC}"

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo -e "${RED}❌ Node.js is not installed. Please install Node.js first.${NC}"
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo -e "${RED}❌ npm is not installed. Please install npm first.${NC}"
    exit 1
fi

# Check if OpenAPI specification exists
if [ ! -f "openapi.yaml" ]; then
    echo -e "${RED}❌ openapi.yaml not found in current directory${NC}"
    exit 1
fi

# Install OpenAPI Generator CLI globally if not already installed
echo -e "${YELLOW}📦 Installing OpenAPI Generator CLI...${NC}"
npm install @openapitools/openapi-generator-cli -g

# Create output directory
echo -e "${YELLOW}📁 Creating output directory...${NC}"
mkdir -p generated/flutter-client

# Generate Dart/Flutter client
echo -e "${YELLOW}🔧 Generating Dart/Flutter client...${NC}"
openapi-generator-cli generate \
    -i openapi.yaml \
    -g dart-dio \
    -o ./generated/flutter-client \
    --additional-properties=pubName=smartdoc_api,pubVersion=1.0.0,returnResponse=true,useEnumExtension=true

# Check if generation was successful
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Flutter API client generated successfully!${NC}"
    echo -e "${YELLOW}📂 Output location: ./generated/flutter-client${NC}"
    
    # Create a README for the generated client
    cat > generated/flutter-client/README.md << 'EOF'
# SmartDoc API Flutter Client

This is an auto-generated Flutter/Dart client for the SmartDoc AI API.

## Installation

Add this to your `pubspec.yaml`:

```yaml
dependencies:
  smartdoc_api:
    path: ./generated/flutter-client
```

## Usage

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

## Features

- ✅ Document upload
- ✅ OCR processing
- ✅ Summary generation
- ✅ Document history
- ✅ Document management
- ✅ Firebase authentication
- ✅ Error handling
- ✅ Type-safe API calls

## Generated from

- OpenAPI Specification: `openapi.yaml`
- Generator: OpenAPI Generator CLI
- Language: Dart with Dio HTTP client
EOF

    echo -e "${GREEN}📝 README created for the generated client${NC}"
    
    # Show generated files
    echo -e "${YELLOW}📋 Generated files:${NC}"
    ls -la generated/flutter-client/
    
    # Create archive for easy distribution
    echo -e "${YELLOW}📦 Creating archive...${NC}"
    cd generated/flutter-client
    tar -czf ../../flutter-client.tar.gz .
    cd ../..
    
    echo -e "${GREEN}✅ Archive created: flutter-client.tar.gz${NC}"
    
else
    echo -e "${RED}❌ Failed to generate Flutter API client${NC}"
    exit 1
fi

echo -e "${GREEN}🎉 Flutter API client generation completed!${NC}"
echo -e "${YELLOW}💡 Next steps:${NC}"
echo -e "   1. Copy the generated client to your Flutter project"
echo -e "   2. Add it to your pubspec.yaml dependencies"
echo -e "   3. Import and use the API client in your Flutter app" 