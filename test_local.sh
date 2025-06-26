#!/bin/bash

# Test script for SmartDoc AI API with local storage

echo "🧪 Testing SmartDoc AI API with local storage..."

# Wait for server to start
echo "⏳ Waiting for server to start..."
sleep 5

# Test health endpoint
echo "🏥 Testing health endpoint..."
curl -s http://localhost:8080/health
echo -e "\n"

# Generate test token
echo "🔑 Generating test token..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/auth/test-token \
  -H "Content-Type: application/json" \
  -d '{"user_id": "test_user_123"}')

echo "Token response: $TOKEN_RESPONSE"

# Extract token (simple parsing)
TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "Token: $TOKEN"

if [ -z "$TOKEN" ]; then
    echo "❌ Failed to get test token"
    exit 1
fi

# Test document list
echo "📄 Testing document list..."
curl -s -X GET http://localhost:8080/documents \
  -H "Authorization: Bearer $TOKEN" | jq .
echo -e "\n"

# Test document upload (create a test file)
echo "📤 Testing document upload..."
echo "This is a test document content" > test_document.txt

UPLOAD_RESPONSE=$(curl -s -X POST http://localhost:8080/documents \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@test_document.txt")

echo "Upload response: $UPLOAD_RESPONSE"

# Extract document ID (simple parsing)
DOC_ID=$(echo $UPLOAD_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Document ID: $DOC_ID"

if [ -z "$DOC_ID" ]; then
    echo "❌ Failed to upload document"
    exit 1
fi

# Test get document
echo "📖 Testing get document..."
curl -s -X GET http://localhost:8080/documents/$DOC_ID \
  -H "Authorization: Bearer $TOKEN" | jq .
echo -e "\n"

# Test OCR trigger
echo "🔍 Testing OCR trigger..."
curl -s -X POST http://localhost:8080/documents/$DOC_ID/ocr \
  -H "Authorization: Bearer $TOKEN" | jq .
echo -e "\n"

# Test summary generation
echo "📝 Testing summary generation..."
curl -s -X POST http://localhost:8080/documents/$DOC_ID/summary \
  -H "Authorization: Bearer $TOKEN" | jq .
echo -e "\n"

# Test document list again
echo "📄 Testing document list after upload..."
curl -s -X GET http://localhost:8080/documents \
  -H "Authorization: Bearer $TOKEN" | jq .
echo -e "\n"

# Clean up
echo "🧹 Cleaning up..."
rm -f test_document.txt

echo "✅ All tests completed!"
echo "📁 Check the ./data/ directory for uploaded files"
echo "🌐 Swagger UI available at: http://localhost:8080/swagger/" 