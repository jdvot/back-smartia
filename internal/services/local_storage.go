package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// LocalStorageService handles local file storage for development/testing
type LocalStorageService struct {
	basePath string
}

// NewLocalStorageService creates a new local storage service
func NewLocalStorageService() *LocalStorageService {
	basePath := os.Getenv("LOCAL_STORAGE_PATH")
	if basePath == "" {
		basePath = "/app/data"
	}
	
	return &LocalStorageService{
		basePath: basePath,
	}
}

// UploadDocument saves a file locally and returns document metadata
func (s *LocalStorageService) UploadDocument(ctx context.Context, userID string, file *multipart.FileHeader) (*Document, error) {
	// Generate unique ID and storage path
	docID := generateID()
	storagePath := filepath.Join(s.basePath, "users", userID, "documents", docID+filepath.Ext(file.Filename))
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(storagePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Open source file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()
	
	// Create destination file
	dst, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()
	
	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}
	
	// Create document metadata
	doc := &Document{
		ID:            docID,
		UserID:        userID,
		Filename:      file.Filename,
		Size:          file.Size,
		MimeType:      file.Header.Get("Content-Type"),
		UploadDate:    time.Now(),
		Status:        "uploaded",
		OcrStatus:     "pending",
		SummaryStatus: "pending",
		StoragePath:   storagePath,
	}
	
	return doc, nil
}

// GetDocument retrieves a document by ID (mock implementation for local storage)
func (s *LocalStorageService) GetDocument(ctx context.Context, docID, userID string) (*Document, error) {
	// For local storage, we'll return a mock document
	// In a real implementation, you'd store metadata in a local database or file
	doc := &Document{
		ID:            docID,
		UserID:        userID,
		Filename:      "mock-document.pdf",
		Size:          1024,
		MimeType:      "application/pdf",
		UploadDate:    time.Now(),
		Status:        "uploaded",
		OcrStatus:     "pending",
		SummaryStatus: "pending",
		StoragePath:   filepath.Join(s.basePath, "users", userID, "documents", docID+".pdf"),
	}
	
	return doc, nil
}

// ListDocuments retrieves all documents for a user (mock implementation)
func (s *LocalStorageService) ListDocuments(ctx context.Context, userID string, limit int) ([]*Document, error) {
	// For local storage, return mock documents
	var documents []*Document
	for i := 0; i < 3; i++ { // Return 3 mock documents
		doc := &Document{
			ID:            fmt.Sprintf("doc_%d", i),
			UserID:        userID,
			Filename:      fmt.Sprintf("document_%d.pdf", i),
			Size:          int64(1024 * (i + 1)),
			MimeType:      "application/pdf",
			UploadDate:    time.Now().Add(-time.Duration(i) * time.Hour),
			Status:        "uploaded",
			OcrStatus:     "pending",
			SummaryStatus: "pending",
			StoragePath:   filepath.Join(s.basePath, "users", userID, "documents", fmt.Sprintf("doc_%d.pdf", i)),
		}
		documents = append(documents, doc)
	}
	
	return documents, nil
}

// UpdateDocument updates document metadata (mock implementation)
func (s *LocalStorageService) UpdateDocument(ctx context.Context, doc *Document) error {
	// For local storage, just log the update
	fmt.Printf("Updated document: %s\n", doc.ID)
	return nil
}

// DeleteDocument deletes a document and its file
func (s *LocalStorageService) DeleteDocument(ctx context.Context, docID, userID string) error {
	// Get document first to get storage path
	doc, err := s.GetDocument(ctx, docID, userID)
	if err != nil {
		return err
	}
	
	// Delete file
	if err := os.Remove(doc.StoragePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	
	return nil
}

// GetFileReader returns a reader for the document file
func (s *LocalStorageService) GetFileReader(ctx context.Context, doc *Document) (io.ReadCloser, error) {
	file, err := os.Open(doc.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
} 