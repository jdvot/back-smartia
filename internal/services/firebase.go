package services

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var (
	FirebaseApp    *firebase.App
	FirebaseAuth   *auth.Client
	FirestoreClient *firestore.Client
	StorageClient   *storage.Client
)

// InitFirebase initializes all Firebase services (Auth, Firestore, Storage)
func InitFirebase() error {
	if FirebaseApp != nil {
		return nil
	}

	// Initialize Firebase App
	ctx := context.Background()
	
	// Use service account key if available, otherwise use default credentials
	var opts []option.ClientOption
	if os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY") != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY"))))
	}

	config := &firebase.Config{
		ProjectID: os.Getenv("FIREBASE_PROJECT_ID"),
	}

	app, err := firebase.NewApp(ctx, config, opts...)
	if err != nil {
		return fmt.Errorf("error initializing firebase app: %w", err)
	}
	FirebaseApp = app

	// Initialize Firebase Auth
	authClient, err := app.Auth(ctx)
	if err != nil {
		return fmt.Errorf("error initializing firebase auth: %w", err)
	}
	FirebaseAuth = authClient

	// Initialize Firestore
	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("error initializing firestore: %w", err)
	}
	FirestoreClient = firestoreClient

	// Initialize Firebase Storage
	storageClient, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("error initializing storage client: %w", err)
	}
	StorageClient = storageClient

	log.Println("Firebase services initialized successfully")
	return nil
}

// CloseFirebase closes all Firebase connections
func CloseFirebase() {
	if FirestoreClient != nil {
		FirestoreClient.Close()
	}
	if StorageClient != nil {
		StorageClient.Close()
	}
} 