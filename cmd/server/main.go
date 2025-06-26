package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"smartdoc-ai/api"
	"smartdoc-ai/internal/auth"
	"smartdoc-ai/internal/services"
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
	handler := api.NewStrictHandler(&ServerImpl{
		StorageService:  services.NewStorageService(),
		OCRService:      services.NewOCRService(),
		SummaryService:  services.NewSummaryService(),
	}, nil)

	// Setup HTTP multiplexer
	mux := http.NewServeMux()

	// Register OpenAPI handlers
	api.RegisterHandlers(mux, handler)

	// Add Swagger UI static files if needed (optional)
	// mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("./swagger-ui"))))

	// Add health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

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