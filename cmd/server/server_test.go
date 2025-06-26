package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"smartdoc-ai/internal/auth"
)

func TestServerImpl_UploadDocument(t *testing.T) {
	// Set development environment
	os.Setenv("ENV", "development")
	os.Setenv("STORAGE_TYPE", "local")
	defer os.Unsetenv("ENV")
	defer os.Unsetenv("STORAGE_TYPE")

	server := &ServerImpl{
		StorageService:  &MockStorageService{},
		OCRService:      &MockOCRService{},
		SummaryService:  &MockSummaryService{},
	}

	tests := []struct {
		name           string
		fileContent    string
		fileName       string
		userID         string
		expectedStatus int
	}{
		{
			name:           "Valid file upload",
			fileContent:    "test document content",
			fileName:       "test.pdf",
			userID:         "test-user-123",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Empty file",
			fileContent:    "",
			fileName:       "empty.pdf",
			userID:         "test-user-123",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create multipart form
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			
			part, err := writer.CreateFormFile("file", tt.fileName)
			if err != nil {
				t.Fatal(err)
			}
			part.Write([]byte(tt.fileContent))
			writer.Close()

			// Create request
			req := httptest.NewRequest("POST", "/docs/upload", &buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			
			// Add user context
			ctx := context.WithValue(req.Context(), auth.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			server.UploadDocument(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				
				if success, ok := response["success"].(bool); !ok || !success {
					t.Errorf("Expected success: true, got %v", success)
				}
			}
		})
	}
}

func TestServerImpl_TriggerOCR(t *testing.T) {
	// Set development environment
	os.Setenv("ENV", "development")
	os.Setenv("STORAGE_TYPE", "local")
	defer os.Unsetenv("ENV")
	defer os.Unsetenv("STORAGE_TYPE")

	server := &ServerImpl{
		StorageService:  &MockStorageService{},
		OCRService:      &MockOCRService{},
		SummaryService:  &MockSummaryService{},
	}

	tests := []struct {
		name           string
		docID          string
		userID         string
		expectedStatus int
	}{
		{
			name:           "Valid OCR trigger",
			docID:          "test-doc-123",
			userID:         "test-user-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid document ID",
			docID:          "invalid-doc",
			userID:         "test-user-123",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/docs/"+tt.docID+"/ocr", nil)
			
			// Add user context
			ctx := context.WithValue(req.Context(), auth.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			server.TriggerOCR(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestServerImpl_TriggerSummary(t *testing.T) {
	// Set development environment
	os.Setenv("ENV", "development")
	os.Setenv("STORAGE_TYPE", "local")
	defer os.Unsetenv("ENV")
	defer os.Unsetenv("STORAGE_TYPE")

	server := &ServerImpl{
		StorageService:  &MockStorageService{},
		OCRService:      &MockOCRService{},
		SummaryService:  &MockSummaryService{},
	}

	tests := []struct {
		name           string
		docID          string
		userID         string
		expectedStatus int
	}{
		{
			name:           "Valid summary trigger",
			docID:          "test-doc-123",
			userID:         "test-user-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid document ID",
			docID:          "invalid-doc",
			userID:         "test-user-123",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/docs/"+tt.docID+"/summary", nil)
			
			// Add user context
			ctx := context.WithValue(req.Context(), auth.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			server.TriggerSummary(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestServerImpl_GetDocumentHistory(t *testing.T) {
	// Set development environment
	os.Setenv("ENV", "development")
	os.Setenv("STORAGE_TYPE", "local")
	defer os.Unsetenv("ENV")
	defer os.Unsetenv("STORAGE_TYPE")

	server := &ServerImpl{
		StorageService:  &MockStorageService{},
		OCRService:      &MockOCRService{},
		SummaryService:  &MockSummaryService{},
	}

	tests := []struct {
		name           string
		userID         string
		query          string
		expectedStatus int
	}{
		{
			name:           "Get document history",
			userID:         "test-user-123",
			query:          "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get document history with pagination",
			userID:         "test-user-123",
			query:          "?page=1&limit=10",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/docs/history"+tt.query, nil)
			
			// Add user context
			ctx := context.WithValue(req.Context(), auth.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			server.GetDocumentHistory(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				
				if success, ok := response["success"].(bool); !ok || !success {
					t.Errorf("Expected success: true, got %v", success)
				}
			}
		})
	}
}

func TestServerImpl_GetDocument(t *testing.T) {
	// Set development environment
	os.Setenv("ENV", "development")
	os.Setenv("STORAGE_TYPE", "local")
	defer os.Unsetenv("ENV")
	defer os.Unsetenv("STORAGE_TYPE")

	server := &ServerImpl{
		StorageService:  &MockStorageService{},
		OCRService:      &MockOCRService{},
		SummaryService:  &MockSummaryService{},
	}

	tests := []struct {
		name           string
		docID          string
		userID         string
		expectedStatus int
	}{
		{
			name:           "Get valid document",
			docID:          "test-doc-123",
			userID:         "test-user-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get invalid document",
			docID:          "invalid-doc",
			userID:         "test-user-123",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/docs/"+tt.docID, nil)
			
			// Add user context
			ctx := context.WithValue(req.Context(), auth.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			server.GetDocument(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestServerImpl_DeleteDocument(t *testing.T) {
	// Set development environment
	os.Setenv("ENV", "development")
	os.Setenv("STORAGE_TYPE", "local")
	defer os.Unsetenv("ENV")
	defer os.Unsetenv("STORAGE_TYPE")

	server := &ServerImpl{
		StorageService:  &MockStorageService{},
		OCRService:      &MockOCRService{},
		SummaryService:  &MockSummaryService{},
	}

	tests := []struct {
		name           string
		docID          string
		userID         string
		expectedStatus int
	}{
		{
			name:           "Delete valid document",
			docID:          "test-doc-123",
			userID:         "test-user-123",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Delete invalid document",
			docID:          "invalid-doc",
			userID:         "test-user-123",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/docs/"+tt.docID, nil)
			
			// Add user context
			ctx := context.WithValue(req.Context(), auth.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			server.DeleteDocument(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// Mock services for testing
type MockStorageService struct{}

func (m *MockStorageService) UploadDocument(file io.Reader, filename string, userID string) (*Document, error) {
	if filename == "" {
		return nil, fmt.Errorf("invalid filename")
	}
	
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	
	if len(content) == 0 {
		return nil, fmt.Errorf("empty file")
	}
	
	return &Document{
		ID:          "test-doc-123",
		Filename:    filename,
		Size:        int64(len(content)),
		MimeType:    "application/pdf",
		UploadDate:  time.Now().Format(time.RFC3339),
		UserID:      userID,
		Status:      "uploaded",
		OCRStatus:   "pending",
		SummaryStatus: "pending",
	}, nil
}

func (m *MockStorageService) GetDocument(docID string, userID string) (*Document, error) {
	if docID == "invalid-doc" {
		return nil, fmt.Errorf("document not found")
	}
	
	return &Document{
		ID:          docID,
		Filename:    "test.pdf",
		Size:        1024,
		MimeType:    "application/pdf",
		UploadDate:  time.Now().Format(time.RFC3339),
		UserID:      userID,
		Status:      "completed",
		OCRStatus:   "completed",
		SummaryStatus: "completed",
		OCRText:     "Test OCR text",
		Summary:     "Test summary",
	}, nil
}

func (m *MockStorageService) GetDocumentHistory(userID string, page, limit int) (*DocumentHistoryResponse, error) {
	return &DocumentHistoryResponse{
		Success: true,
		Message: "Document history retrieved successfully",
		Data: DocumentHistoryData{
			Documents: []Document{
				{
					ID:          "test-doc-123",
					Filename:    "test.pdf",
					Size:        1024,
					MimeType:    "application/pdf",
					UploadDate:  time.Now().Format(time.RFC3339),
					UserID:      userID,
					Status:      "completed",
					OCRStatus:   "completed",
					SummaryStatus: "completed",
				},
			},
			Pagination: Pagination{
				Page:       page,
				Limit:      limit,
				Total:      1,
				TotalPages: 1,
			},
		},
	}, nil
}

func (m *MockStorageService) DeleteDocument(docID string, userID string) error {
	if docID == "invalid-doc" {
		return fmt.Errorf("document not found")
	}
	return nil
}

type MockOCRService struct{}

func (m *MockOCRService) ProcessDocument(docID string) error {
	if docID == "invalid-doc" {
		return fmt.Errorf("document not found")
	}
	return nil
}

type MockSummaryService struct{}

func (m *MockSummaryService) GenerateSummary(docID string) error {
	if docID == "invalid-doc" {
		return fmt.Errorf("document not found")
	}
	return nil
} 