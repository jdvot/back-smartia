package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// FirebaseApp holds the Firebase application instance.
var FirebaseApp *firebase.App

// AuthClient holds the Firebase Auth client instance.
var AuthClient *auth.Client

// userIDKeyType est un type pour la clé de contexte utilisateur.
type userIDKeyType struct{}

// UserIDKey est la clé typée pour stocker l'ID utilisateur dans le contexte.
var UserIDKey = userIDKeyType{}

// InitializeFirebase initializes Firebase Admin SDK
func InitializeFirebase() error {
	if os.Getenv("ENV") == "development" && os.Getenv("STORAGE_TYPE") == "local" {
		return nil
	}
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		return fmt.Errorf("FIREBASE_PROJECT_ID environment variable is required")
	}
	serviceAccountKey := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY")
	if serviceAccountKey == "" {
		return fmt.Errorf("FIREBASE_SERVICE_ACCOUNT_KEY environment variable is required")
	}
	opt := option.WithCredentialsJSON([]byte(serviceAccountKey))
	app, err := firebase.NewApp(context.Background(), &firebase.Config{
		ProjectID: projectID,
	}, opt)
	if err != nil {
		return fmt.Errorf("failed to initialize Firebase app: %w", err)
	}
	FirebaseApp = app
	authClient, err := app.Auth(context.Background())
	if err != nil {
		return fmt.Errorf("failed to initialize Firebase Auth client: %w", err)
	}
	AuthClient = authClient
	return nil
}

// Middleware validates Firebase ID tokens
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || strings.HasPrefix(r.URL.Path, "/swagger/") {
			next.ServeHTTP(w, r)
			return
		}
		if os.Getenv("ENV") == "development" && os.Getenv("STORAGE_TYPE") == "local" {
			userID := validateTestToken(r)
			if userID != "" {
				ctx := context.WithValue(r.Context(), UserIDKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		if AuthClient == nil {
			http.Error(w, "Authentication service not available", http.StatusInternalServerError)
			return
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return
		}
		token, err := AuthClient.VerifyIDToken(r.Context(), tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, token.UID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateTestToken validates a test token for development
func validateTestToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return ""
	}
	tokenBytes, err := base64.StdEncoding.DecodeString(tokenString)
	if err != nil {
		return ""
	}
	var token map[string]interface{}
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		return ""
	}
	if exp, ok := token["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return ""
		}
	}
	if userID, ok := token["user_id"].(string); ok {
		return userID
	}
	return ""
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return "", fmt.Errorf("user ID not found in context")
	}
	return userID, nil
} 