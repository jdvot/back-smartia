package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// DocumentsCollection is the Firestore collection name for documents.
const DocumentsCollection = "documents"

// StorageBucket is the default Firebase Storage bucket name.
const StorageBucket = "smartdoc-uploads" // Will be overridden by env var

// Document represents a document in Firestore
type Document struct {
	ID            string     `firestore:"id"`
	UserID        string     `firestore:"userId"`
	Filename      string     `firestore:"filename"`
	Size          int64      `firestore:"size"`
	MimeType      string     `firestore:"mimeType"`
	UploadDate    time.Time  `firestore:"uploadDate"`
	Status        string     `firestore:"status"`
	OcrText       *string    `firestore:"ocrText,omitempty"`
	Summary       *string    `firestore:"summary,omitempty"`
	OcrStatus     string     `firestore:"ocrStatus"`
	SummaryStatus string     `firestore:"summaryStatus"`
	StoragePath   string     `firestore:"storagePath"`
}

// StorageService handles document storage operations
type StorageService struct {
	firestore *firestore.Client
	storage   *storage.Client
	bucket    string
	local     *LocalStorageService
	useLocal  bool
}

// NewStorageService creates a new storage service
func NewStorageService() *StorageService {
	storageType := os.Getenv("STORAGE_TYPE")
	useLocal := storageType == "local"
	
	bucket := StorageBucket
	if envBucket := os.Getenv("FIREBASE_STORAGE_BUCKET"); envBucket != "" {
		bucket = envBucket
	}
	
	service := &StorageService{
		firestore: FirestoreClient,
		storage:   StorageClient,
		bucket:    bucket,
		useLocal:  useLocal,
	}
	
	if useLocal {
		service.local = NewLocalStorageService()
	}
	
	return service
}

// UploadDocument uploads a file to storage and saves metadata
func (s *StorageService) UploadDocument(ctx context.Context, userID string, file *multipart.FileHeader) (*Document, error) {
	if s.useLocal {
		return s.local.UploadDocument(ctx, userID, file)
	}
	
	// Firebase Storage implementation
	// Generate unique ID and storage path
	docID := generateID()
	storagePath := fmt.Sprintf("users/%s/documents/%s%s", userID, docID, filepath.Ext(file.Filename))
	
	// Upload file to Firebase Storage
	fileReader, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer fileReader.Close()

	bucket := s.storage.Bucket(s.bucket)
	obj := bucket.Object(storagePath)
	writer := obj.NewWriter(ctx)
	writer.ContentType = file.Header.Get("Content-Type")

	if _, err := io.Copy(writer, fileReader); err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
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

	// Save to Firestore
	_, err = s.firestore.Collection(DocumentsCollection).Doc(docID).Set(ctx, doc)
	if err != nil {
		// Clean up storage if Firestore fails
		if deleteErr := obj.Delete(ctx); deleteErr != nil {
			// Log the delete error but return the original error
			fmt.Printf("Failed to delete object after Firestore error: %v", deleteErr)
		}
		return nil, fmt.Errorf("failed to save document metadata: %w", err)
	}

	return doc, nil
}

// GetDocument retrieves a document by ID
func (s *StorageService) GetDocument(ctx context.Context, docID, userID string) (*Document, error) {
	if s.useLocal {
		return s.local.GetDocument(ctx, docID, userID)
	}
	
	// Firebase implementation
	docRef := s.firestore.Collection(DocumentsCollection).Doc(docID)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	var doc Document
	if err := docSnap.DataTo(&doc); err != nil {
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	// Check if user owns this document
	if doc.UserID != userID {
		return nil, fmt.Errorf("document not found")
	}

	return &doc, nil
}

// ListDocuments retrieves all documents for a user
func (s *StorageService) ListDocuments(ctx context.Context, userID string, limit int) ([]*Document, error) {
	if s.useLocal {
		return s.local.ListDocuments(ctx, userID, limit)
	}
	
	// Firebase implementation
	query := s.firestore.Collection(DocumentsCollection).
		Where("userId", "==", userID).
		OrderBy("uploadDate", firestore.Desc).
		Limit(limit)

	iter := query.Documents(ctx)
	var documents []*Document

	for {
		docSnap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate documents: %w", err)
		}

		var doc Document
		if err := docSnap.DataTo(&doc); err != nil {
			return nil, fmt.Errorf("failed to parse document: %w", err)
		}
		documents = append(documents, &doc)
	}

	return documents, nil
}

// UpdateDocument updates document metadata
func (s *StorageService) UpdateDocument(ctx context.Context, doc *Document) error {
	if s.useLocal {
		return s.local.UpdateDocument(ctx, doc)
	}
	
	// Firebase implementation
	_, err := s.firestore.Collection(DocumentsCollection).Doc(doc.ID).Set(ctx, doc)
	return err
}

// DeleteDocument deletes a document and its file
func (s *StorageService) DeleteDocument(ctx context.Context, docID, userID string) error {
	if s.useLocal {
		return s.local.DeleteDocument(ctx, docID, userID)
	}
	
	// Firebase implementation
	// Get document first to check ownership and get storage path
	doc, err := s.GetDocument(ctx, docID, userID)
	if err != nil {
		return err
	}

	// Delete from Firebase Storage
	bucket := s.storage.Bucket(s.bucket)
	obj := bucket.Object(doc.StoragePath)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	// Delete from Firestore
	_, err = s.firestore.Collection(DocumentsCollection).Doc(docID).Delete(ctx)
	return err
}

// GetFileReader returns a reader for the document file
func (s *StorageService) GetFileReader(ctx context.Context, doc *Document) (io.ReadCloser, error) {
	if s.useLocal {
		return s.local.GetFileReader(ctx, doc)
	}
	
	// Firebase implementation
	bucket := s.storage.Bucket(s.bucket)
	obj := bucket.Object(doc.StoragePath)
	return obj.NewReader(ctx)
}

// Helper function to generate unique IDs
func generateID() string {
	return fmt.Sprintf("doc_%d", time.Now().UnixNano())
} 