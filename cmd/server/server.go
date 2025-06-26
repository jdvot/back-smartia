package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"smartdoc-ai/api"
	"smartdoc-ai/internal/auth"
	"smartdoc-ai/internal/services"
)

// ServerImpl implements the generated ServerInterface
type ServerImpl struct {
	storageService  *services.StorageService
	ocrService      *services.OCRService
	summaryService  *services.SummaryService
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
	doc, err := s.storageService.UploadDocument(r.Context(), userID, header)
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
		Status:        doc.Status,
		OcrText:       doc.OcrText,
		Summary:       doc.Summary,
		OcrStatus:     &doc.OcrStatus,
		SummaryStatus: &doc.SummaryStatus,
	}

	resp := api.UploadResponse{
		Success: api.Ptr(true),
		Message: api.Ptr("Document uploaded successfully"),
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
	doc, err := s.storageService.GetDocument(r.Context(), docId, userID)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Update status to processing
	doc.OcrStatus = "processing"
	doc.Status = "processing"
	if err := s.storageService.UpdateDocument(r.Context(), doc); err != nil {
		http.Error(w, "Failed to update document status", http.StatusInternalServerError)
		return
	}

	// Process OCR
	fileReader, err := s.storageService.GetFileReader(r.Context(), doc)
	if err != nil {
		http.Error(w, "Failed to read document file", http.StatusInternalServerError)
		return
	}
	defer fileReader.Close()

	ocrText, err := s.ocrService.ProcessOCR(r.Context(), fileReader)
	if err != nil {
		// Update status to failed
		doc.OcrStatus = "failed"
		doc.Status = "failed"
		s.storageService.UpdateDocument(r.Context(), doc)
		
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

	if err := s.storageService.UpdateDocument(r.Context(), doc); err != nil {
		http.Error(w, "Failed to update document with OCR results", http.StatusInternalServerError)
		return
	}

	resp := api.OCRResponse{
		Success: api.Ptr(true),
		Message: api.Ptr("OCR processing completed successfully"),
		Data: api.OCRResponseData{
			DocId:    doc.ID,
			OcrText:  &ocrText,
			Status:   doc.OcrStatus,
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
	doc, err := s.storageService.GetDocument(r.Context(), docId, userID)
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
	if err := s.storageService.UpdateDocument(r.Context(), doc); err != nil {
		http.Error(w, "Failed to update document status", http.StatusInternalServerError)
		return
	}

	// Generate summary
	summary, err := s.summaryService.GenerateSummary(r.Context(), *doc.OcrText)
	if err != nil {
		// Update status to failed
		doc.SummaryStatus = "failed"
		doc.Status = "failed"
		s.storageService.UpdateDocument(r.Context(), doc)
		
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

	if err := s.storageService.UpdateDocument(r.Context(), doc); err != nil {
		http.Error(w, "Failed to update document with summary results", http.StatusInternalServerError)
		return
	}

	resp := api.SummaryResponse{
		Success: api.Ptr(true),
		Message: api.Ptr("Summary generation completed successfully"),
		Data: api.SummaryResponseData{
			DocId:   doc.ID,
			Summary: &summary,
			Status:  doc.SummaryStatus,
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
	documents, err := s.storageService.ListDocuments(r.Context(), userID, limit)
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
			Status:        doc.Status,
			OcrText:       doc.OcrText,
			Summary:       doc.Summary,
			OcrStatus:     doc.OcrStatus,
			SummaryStatus: doc.SummaryStatus,
		}
		apiDocs = append(apiDocs, apiDoc)
	}

	resp := api.DocumentHistoryResponse{
		Success: api.Ptr(true),
		Message: api.Ptr("Document history retrieved successfully"),
		Data: api.DocumentHistoryResponseData{
			Documents: apiDocs,
			Pagination: api.Pagination{
				Page:       page,
				Limit:      limit,
				Total:      len(apiDocs),
				TotalPages: 1, // Simplified for now
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
	doc, err := s.storageService.GetDocument(r.Context(), docId, userID)
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
		Status:        doc.Status,
		OcrText:       doc.OcrText,
		Summary:       doc.Summary,
		OcrStatus:     doc.OcrStatus,
		SummaryStatus: doc.SummaryStatus,
	}

	resp := api.DocumentResponse{
		Success: api.Ptr(true),
		Message: api.Ptr("Document details retrieved successfully"),
		Data:    apiDoc,
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
	err = s.storageService.DeleteDocument(r.Context(), docId, userID)
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