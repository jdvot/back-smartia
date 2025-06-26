package main

import (
	// "context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	// "io"
	"net/http"
	// "os"
	// "strconv"
	// "strings"
	"time"

	// "github.com/gin-gonic/gin"

	"smartdoc-ai/api"
	"smartdoc-ai/internal/auth"
	"smartdoc-ai/internal/services"
)

// Helper function to create pointers
func Ptr[T any](v T) *T {
	return &v
}

// Helper function to convert string to DocumentStatus
func toDocumentStatus(s string) api.DocumentStatus {
	switch s {
	case "completed":
		return api.DocumentStatusCompleted
	case "failed":
		return api.DocumentStatusFailed
	case "processing":
		return api.DocumentStatusProcessing
	case "uploaded":
		return api.DocumentStatusUploaded
	default:
		return api.DocumentStatusUploaded
	}
}

// Helper function to convert string to DocumentOcrStatus
func toDocumentOcrStatus(s string) api.DocumentOcrStatus {
	switch s {
	case "completed":
		return api.DocumentOcrStatusCompleted
	case "failed":
		return api.DocumentOcrStatusFailed
	case "pending":
		return api.DocumentOcrStatusPending
	case "processing":
		return api.DocumentOcrStatusProcessing
	default:
		return api.DocumentOcrStatusPending
	}
}

// Helper function to convert string to DocumentSummaryStatus
func toDocumentSummaryStatus(s string) api.DocumentSummaryStatus {
	switch s {
	case "completed":
		return api.DocumentSummaryStatusCompleted
	case "failed":
		return api.DocumentSummaryStatusFailed
	case "pending":
		return api.DocumentSummaryStatusPending
	case "processing":
		return api.DocumentSummaryStatusProcessing
	default:
		return api.DocumentSummaryStatusPending
	}
}

// ServerImpl implements the generated ServerInterface
type ServerImpl struct {
	StorageService  *services.StorageService
	OCRService      *services.OCRService
	SummaryService  *services.SummaryService
}

// Helper to extract userID from context
func getUserID(r *http.Request) (string, error) {
	return auth.GetUserIDFromContext(r.Context())
}

// UploadDocument handles document upload
func (s *ServerImpl) UploadDocument(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Invalid multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Upload document using storage service
	doc, err := s.StorageService.UploadDocument(r.Context(), userID, header)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload document: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to API response format
	apiDoc := api.Document{
		Id:            doc.ID,
		Filename:      doc.Filename,
		Size:          int(doc.Size),
		MimeType:      doc.MimeType,
		UploadDate:    doc.UploadDate,
		UserId:        doc.UserID,
		Status:        toDocumentStatus(doc.Status),
		OcrText:       doc.OcrText,
		Summary:       doc.Summary,
		OcrStatus:     Ptr(toDocumentOcrStatus(doc.OcrStatus)),
		SummaryStatus: Ptr(toDocumentSummaryStatus(doc.SummaryStatus)),
	}

	resp := api.UploadResponse{
		Success: Ptr(true),
		Message: Ptr("Document uploaded successfully"),
		Data:    &apiDoc,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// TriggerOCR handles OCR processing
func (s *ServerImpl) TriggerOCR(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	docId := r.URL.Query().Get("docId")
	if docId == "" {
		http.Error(w, "docId is required", http.StatusBadRequest)
		return
	}

	// Get document
	doc, err := s.StorageService.GetDocument(r.Context(), docId, userID)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Update status to processing
	doc.OcrStatus = "processing"
	doc.Status = "processing"
	if err := s.StorageService.UpdateDocument(r.Context(), doc); err != nil {
		http.Error(w, "Failed to update document status", http.StatusInternalServerError)
		return
	}

	// Process OCR
	fileReader, err := s.StorageService.GetFileReader(r.Context(), doc)
	if err != nil {
		http.Error(w, "Failed to read document file", http.StatusInternalServerError)
		return
	}
	defer fileReader.Close()

	ocrText, err := s.OCRService.ProcessOCR(r.Context(), fileReader)
	if err != nil {
		// Update status to failed
		doc.OcrStatus = "failed"
		doc.Status = "failed"
		s.StorageService.UpdateDocument(r.Context(), doc)
		
		http.Error(w, fmt.Sprintf("OCR processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Update document with OCR results
	doc.OcrText = &ocrText
	doc.OcrStatus = "completed"
	if doc.SummaryStatus == "completed" {
		doc.Status = "completed"
	} else {
		doc.Status = "uploaded"
	}

	if err := s.StorageService.UpdateDocument(r.Context(), doc); err != nil {
		http.Error(w, "Failed to update document with OCR results", http.StatusInternalServerError)
		return
	}

	resp := api.OCRResponse{
		Success: Ptr(true),
		Message: Ptr("OCR processing completed successfully"),
		Data: &struct {
			DocId   *string                    `json:"docId,omitempty"`
			OcrText *string                    `json:"ocrText"`
			Status  *api.OCRResponseDataStatus `json:"status,omitempty"`
		}{
			DocId:   Ptr(doc.ID),
			OcrText: &ocrText,
			Status:  Ptr(api.OCRResponseDataStatusCompleted),
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// TriggerSummary handles summary generation
func (s *ServerImpl) TriggerSummary(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	docId := r.URL.Query().Get("docId")
	if docId == "" {
		http.Error(w, "docId is required", http.StatusBadRequest)
		return
	}

	// Get document
	doc, err := s.StorageService.GetDocument(r.Context(), docId, userID)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Check if OCR is completed
	if doc.OcrStatus != "completed" || doc.OcrText == nil {
		http.Error(w, "OCR must be completed before generating summary", http.StatusBadRequest)
		return
	}

	// Update status to processing
	doc.SummaryStatus = "processing"
	doc.Status = "processing"
	if err := s.StorageService.UpdateDocument(r.Context(), doc); err != nil {
		http.Error(w, "Failed to update document status", http.StatusInternalServerError)
		return
	}

	// Generate summary
	summary, err := s.SummaryService.GenerateSummary(r.Context(), *doc.OcrText)
	if err != nil {
		// Update status to failed
		doc.SummaryStatus = "failed"
		doc.Status = "failed"
		s.StorageService.UpdateDocument(r.Context(), doc)
		
		http.Error(w, fmt.Sprintf("Summary generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Update document with summary results
	doc.Summary = &summary
	doc.SummaryStatus = "completed"
	if doc.OcrStatus == "completed" {
		doc.Status = "completed"
	} else {
		doc.Status = "uploaded"
	}

	if err := s.StorageService.UpdateDocument(r.Context(), doc); err != nil {
		http.Error(w, "Failed to update document with summary results", http.StatusInternalServerError)
		return
	}

	resp := api.SummaryResponse{
		Success: Ptr(true),
		Message: Ptr("Summary generation completed successfully"),
		Data: &struct {
			DocId   *string                        `json:"docId,omitempty"`
			Status  *api.SummaryResponseDataStatus `json:"status,omitempty"`
			Summary *string                        `json:"summary"`
		}{
			DocId:   Ptr(doc.ID),
			Summary: &summary,
			Status:  Ptr(api.Completed),
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// GetDocumentHistory handles document history retrieval
func (s *ServerImpl) GetDocumentHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Set default values
	limit := 20
	page := 1

	// Get documents
	documents, err := s.StorageService.ListDocuments(r.Context(), userID, limit)
	if err != nil {
		http.Error(w, "Failed to retrieve document history", http.StatusInternalServerError)
		return
	}

	// Convert to API format
	var apiDocs []api.Document
	for _, doc := range documents {
		apiDoc := api.Document{
			Id:            doc.ID,
			Filename:      doc.Filename,
			Size:          int(doc.Size),
			MimeType:      doc.MimeType,
			UploadDate:    doc.UploadDate,
			UserId:        doc.UserID,
			Status:        toDocumentStatus(doc.Status),
			OcrText:       doc.OcrText,
			Summary:       doc.Summary,
			OcrStatus:     Ptr(toDocumentOcrStatus(doc.OcrStatus)),
			SummaryStatus: Ptr(toDocumentSummaryStatus(doc.SummaryStatus)),
		}
		apiDocs = append(apiDocs, apiDoc)
	}

	resp := api.DocumentHistoryResponse{
		Success: Ptr(true),
		Message: Ptr("Document history retrieved successfully"),
		Data: &struct {
			Documents  *[]api.Document `json:"documents,omitempty"`
			Pagination *struct {
				Limit      *int `json:"limit,omitempty"`
				Page       *int `json:"page,omitempty"`
				Total      *int `json:"total,omitempty"`
				TotalPages *int `json:"totalPages,omitempty"`
			} `json:"pagination,omitempty"`
		}{
			Documents: &apiDocs,
			Pagination: &struct {
				Limit      *int `json:"limit,omitempty"`
				Page       *int `json:"page,omitempty"`
				Total      *int `json:"total,omitempty"`
				TotalPages *int `json:"totalPages,omitempty"`
			}{
				Page:       Ptr(page),
				Limit:      Ptr(limit),
				Total:      Ptr(len(apiDocs)),
				TotalPages: Ptr(1), // Simplified for now
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// GetDocument handles single document retrieval
func (s *ServerImpl) GetDocument(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	docId := r.URL.Query().Get("docId")
	if docId == "" {
		http.Error(w, "docId is required", http.StatusBadRequest)
		return
	}

	// Get document
	doc, err := s.StorageService.GetDocument(r.Context(), docId, userID)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Convert to API format
	apiDoc := api.Document{
		Id:            doc.ID,
		Filename:      doc.Filename,
		Size:          int(doc.Size),
		MimeType:      doc.MimeType,
		UploadDate:    doc.UploadDate,
		UserId:        doc.UserID,
		Status:        toDocumentStatus(doc.Status),
		OcrText:       doc.OcrText,
		Summary:       doc.Summary,
		OcrStatus:     Ptr(toDocumentOcrStatus(doc.OcrStatus)),
		SummaryStatus: Ptr(toDocumentSummaryStatus(doc.SummaryStatus)),
	}

	resp := api.DocumentResponse{
		Success: Ptr(true),
		Message: Ptr("Document details retrieved successfully"),
		Data:    &apiDoc,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// DeleteDocument handles document deletion
func (s *ServerImpl) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	docId := r.URL.Query().Get("docId")
	if docId == "" {
		http.Error(w, "docId is required", http.StatusBadRequest)
		return
	}

	// Delete document
	err = s.StorageService.DeleteDocument(r.Context(), docId, userID)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// addTestTokenEndpoint adds a test token endpoint for development
func addTestTokenEndpoint(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/auth/test-token" {
			handleTestToken(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// handleTestToken generates a test token for development
func handleTestToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	// Create a simple test token (not for production use)
	token := map[string]interface{}{
		"user_id": req.UserID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"iss":     "test-issuer",
		"aud":     "test-audience",
	}

	// Encode as JWT-like string (simplified for testing)
	tokenBytes, _ := json.Marshal(token)
	tokenString := base64.StdEncoding.EncodeToString(tokenBytes)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
		"user_id": req.UserID,
	})
} 