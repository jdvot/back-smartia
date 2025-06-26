package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"smartdoc-ai/internal/auth"
	"smartdoc-ai/internal/services"
	
	httpSwagger "github.com/swaggo/http-swagger"
	_ "smartdoc-ai/docs" // This will be generated
)

// @title SmartDoc AI API
// @version 1.0
// @description RESTful API for document processing with OCR and AI summarization
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Firebase Auth ID token
func main() {
	// Initialize Firebase Auth
	if err := auth.InitializeFirebase(); err != nil {
		log.Printf("Warning: Firebase initialization failed: %v", err)
		log.Println("Continuing with local development mode...")
	}

	// Create and start server
	server := createServer()

	// Start server in a goroutine
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		log.Printf("Starting server on port %s", port)
		log.Printf("Swagger documentation available at: http://localhost:%s/swagger/index.html", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func createServer() *http.Server {
	// Create handler instance
	storageService := services.NewStorageService()
	ocrService, err := services.NewOCRService()
	if err != nil {
		log.Fatalf("Failed to create OCR service: %v", err)
	}
	summaryService := services.NewSummaryService()
	
	handler := &ServerImpl{
		StorageService:  storageService,
		OCRService:      ocrService,
		SummaryService:  summaryService,
	}

	// Setup HTTP multiplexer
	mux := http.NewServeMux()

	// Register routes manually
	mux.HandleFunc("POST /docs/upload", handler.UploadDocument)
	mux.HandleFunc("POST /docs/{docId}/ocr", handler.TriggerOCR)
	mux.HandleFunc("POST /docs/{docId}/summary", handler.TriggerSummary)
	mux.HandleFunc("GET /docs/history", handler.GetDocumentHistory)
	mux.HandleFunc("GET /docs/{docId}", handler.GetDocument)
	mux.HandleFunc("DELETE /docs/{docId}", handler.DeleteDocument)

	// Add health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Add Swagger documentation endpoint
	mux.HandleFunc("GET /swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	// Add authentication middleware
	finalHandler := auth.AuthMiddleware(mux)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &http.Server{
		Addr:         ":" + port,
		Handler:      finalHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
} 