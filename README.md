# SmartDoc AI - RESTful API Backend

A contract-first RESTful API backend in Go (Golang 1.22+) for "SmartDoc AI", following OpenAPI 3 (Swagger) specification with Firebase Auth integration, document processing, OCR, and AI summarization.

## 🏗️ Architecture Overview

This project follows a **contract-first approach** where the OpenAPI specification (`openapi.yaml`) is the single source of truth for the API design. The Go server is generated from this contract using `oapi-codegen`.

### Key Features

- ✅ **Contract-First Design**: OpenAPI 3.0 specification drives the entire API
- ✅ **Firebase Auth Integration**: Secure authentication with Firebase ID tokens
- ✅ **Document Processing**: Upload, store, and manage documents
- ✅ **OCR Processing**: Google Vision API or OCR.space integration
- ✅ **AI Summarization**: OpenAI GPT or Google Gemini integration
- ✅ **Firebase Storage**: Scalable file storage
- ✅ **Firestore Database**: NoSQL document database
- ✅ **CORS Support**: Cross-origin resource sharing
- ✅ **Swagger UI**: Live API documentation
- ✅ **Railway/Render Ready**: Easy deployment to cloud platforms

## 📋 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/docs/upload` | Upload a document (multipart) |
| `POST` | `/docs/{docId}/ocr` | Trigger OCR processing |
| `POST` | `/docs/{docId}/summary` | Trigger AI summary generation |
| `GET` | `/docs/history` | List all documents for user |
| `GET` | `/docs/{docId}` | Get document details |
| `DELETE` | `/docs/{docId}` | Delete a document |
| `GET` | `/swagger/` | Swagger UI documentation |

All endpoints require Firebase Auth ID token in the `Authorization: Bearer <token>` header.

## 🚀 Quick Start

### Prerequisites

- Go 1.22+
- Firebase project with Auth, Firestore, and Storage enabled
- (Optional) Google Vision API, OCR.space, OpenAI, or Google Gemini API keys

### 1. Clone and Setup

```bash
git clone <repository-url>
cd smartdoc-ai
```

### 2. Install Dependencies

```bash
go mod download
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

### 3. Configure Environment

Copy the environment example and configure your settings:

```bash
cp env.example .env
```

Edit `.env` with your Firebase and API credentials:

```env
# Firebase Configuration
FIREBASE_PROJECT_ID=your-firebase-project-id
FIREBASE_SERVICE_ACCOUNT_KEY={"type":"service_account",...}

# Storage Configuration
FIREBASE_STORAGE_BUCKET=your-firebase-storage-bucket

# Optional: OCR Services
OCR_SERVICE_URL=https://api.ocr.space/parse/image
OCR_API_KEY=your-ocr-api-key

# Optional: AI Summary Services
AI_SERVICE_URL=https://api.openai.com/v1/chat/completions
AI_API_KEY=your-openai-api-key
```

### 4. Generate Go Code from OpenAPI Contract

```bash
oapi-codegen -config oapi-codegen.yaml openapi.yaml
```

This generates the Go types, handlers, and server stubs in `api/generated.go`.

### 5. Run Locally

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

### 6. Access Swagger UI

Visit `http://localhost:8080/swagger/` to see the live API documentation.

## 🔧 Contract-First Development

### Editing the API Contract

1. **Modify `openapi.yaml`**: Add new endpoints, modify schemas, or update responses
2. **Regenerate Go code**: Run `oapi-codegen -config oapi-codegen.yaml openapi.yaml`
3. **Update handlers**: Modify `cmd/server/server.go` to implement new endpoints
4. **Test**: Use Swagger UI to test your changes

### Example: Adding a New Endpoint

1. Add to `openapi.yaml`:
```yaml
  /docs/{docId}/analyze:
    post:
      summary: Analyze document
      operationId: analyzeDocument
      # ... rest of specification
```

2. Regenerate code:
```bash
oapi-codegen -config oapi-codegen.yaml openapi.yaml
```

3. Implement in `cmd/server/server.go`:
```go
func (s *ServerImpl) AnalyzeDocument(ctx *gin.Context, docId string) {
    // Your implementation here
}
```

## 🚀 Deployment

### Railway.app

1. **Connect Repository**: Link your GitHub repository to Railway
2. **Set Environment Variables**: Add all variables from `.env` in Railway dashboard
3. **Deploy**: Railway will automatically detect Go and deploy

```bash
# Railway CLI (optional)
railway login
railway link
railway up
```

### Render.com

1. **Create New Web Service**: Connect your GitHub repository
2. **Build Command**: `go build -o main cmd/server/main.go`
3. **Start Command**: `./main`
4. **Environment Variables**: Add all variables from `.env`

### Environment Variables for Production

```env
PORT=8080
ENV=production
FIREBASE_PROJECT_ID=your-production-project
FIREBASE_SERVICE_ACCOUNT_KEY={"type":"service_account",...}
FIREBASE_STORAGE_BUCKET=your-production-bucket
CORS_ALLOWED_ORIGINS=https://your-flutter-app.web.app
```

## 🔌 Flutter Integration

### Setup Firebase in Flutter

1. **Add Firebase to Flutter project**:
```bash
flutterfire configure
```

2. **Install dependencies**:
```yaml
dependencies:
  firebase_auth: ^4.15.3
  http: ^1.1.0
```

### API Client Example

```dart
class SmartDocAPI {
  static const String baseUrl = 'https://your-api.railway.app';
  
  static Future<String?> _getAuthToken() async {
    final user = FirebaseAuth.instance.currentUser;
    return await user?.getIdToken();
  }
  
  static Future<Map<String, dynamic>> uploadDocument(File file) async {
    final token = await _getAuthToken();
    if (token == null) throw Exception('Not authenticated');
    
    final request = http.MultipartRequest(
      'POST',
      Uri.parse('$baseUrl/docs/upload'),
    );
    
    request.headers['Authorization'] = 'Bearer $token';
    request.files.add(await http.MultipartFile.fromPath('file', file.path));
    
    final response = await request.send();
    final responseData = await response.stream.bytesToString();
    
    return json.decode(responseData);
  }
  
  static Future<Map<String, dynamic>> triggerOCR(String docId) async {
    final token = await _getAuthToken();
    if (token == null) throw Exception('Not authenticated');
    
    final response = await http.post(
      Uri.parse('$baseUrl/docs/$docId/ocr'),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
    );
    
    return json.decode(response.body);
  }
}
```

### Usage in Flutter

```dart
// Upload document
final result = await SmartDocAPI.uploadDocument(file);
final docId = result['data']['id'];

// Trigger OCR
await SmartDocAPI.triggerOCR(docId);

// Get document history
final history = await SmartDocAPI.getDocumentHistory();
```

## 🔐 Authentication

### Firebase Setup

1. **Create Firebase Project**: Go to [Firebase Console](https://console.firebase.google.com/)
2. **Enable Authentication**: Add Email/Password or Google Sign-in
3. **Enable Firestore**: Create database in test mode
4. **Enable Storage**: Create storage bucket
5. **Generate Service Account Key**: 
   - Go to Project Settings > Service Accounts
   - Click "Generate new private key"
   - Save the JSON file
   - Add content to `FIREBASE_SERVICE_ACCOUNT_KEY` environment variable

### Token Validation

The API validates Firebase ID tokens on every request:

```bash
curl -X POST http://localhost:8080/docs/upload \
  -H "Authorization: Bearer YOUR_FIREBASE_ID_TOKEN" \
  -F "file=@document.pdf"
```

## 🧪 Testing

### Using Swagger UI

1. Start the server: `go run cmd/server/main.go`
2. Visit: `http://localhost:8080/swagger/`
3. Click "Authorize" and enter your Firebase ID token
4. Test endpoints directly from the UI

### Using curl

```bash
# Get Firebase ID token (from your Flutter app or Firebase console)
TOKEN="your-firebase-id-token"

# Upload document
curl -X POST http://localhost:8080/docs/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@test-document.pdf"

# Trigger OCR
curl -X POST http://localhost:8080/docs/DOC_ID/ocr \
  -H "Authorization: Bearer $TOKEN"

# Get document history
curl -X GET http://localhost:8080/docs/history \
  -H "Authorization: Bearer $TOKEN"
```

## 📁 Project Structure

```
smartdoc-ai/
├── openapi.yaml              # OpenAPI 3.0 specification
├── oapi-codegen.yaml         # Code generation config
├── go.mod                    # Go module file
├── env.example               # Environment variables template
├── README.md                 # This file
├── api/
│   └── generated.go          # Generated Go types and handlers
├── cmd/
│   └── server/
│       ├── main.go           # Server entry point
│       └── server.go         # Handler implementations
└── internal/
    ├── auth/
    │   └── firebase.go       # Firebase authentication
    └── services/
        ├── firebase.go       # Firebase initialization
        ├── storage.go        # Document storage service
        ├── ocr.go           # OCR processing service
        └── summary.go       # AI summary service
```

## 🔧 Configuration Options

### OCR Services

The system supports multiple OCR providers:

1. **Google Vision API** (Recommended): Set `GOOGLE_APPLICATION_CREDENTIALS` or use Firebase service account
2. **OCR.space**: Set `OCR_SERVICE_URL` and `OCR_API_KEY`
3. **Mock OCR**: Fallback when no real service is configured

### AI Summary Services

The system supports multiple AI providers:

1. **OpenAI GPT**: Set `AI_SERVICE_URL` and `AI_API_KEY`
2. **Google Gemini**: Set `GEMINI_API_URL` and `GEMINI_API_KEY`
3. **Mock Summary**: Fallback when no real service is configured

### Storage Options

1. **Firebase Storage** (Recommended): Set `FIREBASE_STORAGE_BUCKET`
2. **Local Storage**: Set `STORAGE_TYPE=local` (for development)

## 🐛 Troubleshooting

### Common Issues

1. **Firebase Connection Error**:
   - Verify `FIREBASE_PROJECT_ID` is correct
   - Check `FIREBASE_SERVICE_ACCOUNT_KEY` format
   - Ensure Firebase services are enabled

2. **CORS Errors**:
   - Update `CORS_ALLOWED_ORIGINS` in environment
   - Check Flutter app origin matches

3. **File Upload Issues**:
   - Verify Firebase Storage bucket exists
   - Check storage permissions
   - Ensure file size is within limits

4. **OCR/Summary Failures**:
   - Check API keys are valid
   - Verify service endpoints are accessible
   - Review error logs for specific issues

### Logs

Enable debug logging by setting:
```env
LOG_LEVEL=debug
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Update tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:
- Create an issue on GitHub
- Check the Swagger UI documentation
- Review Firebase console for authentication issues

---

**SmartDoc AI** - Making document processing intelligent and accessible. 