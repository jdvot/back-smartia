package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestValidateTestToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "Valid token",
			token:    createTestToken("test-user-123", time.Now().Add(time.Hour).Unix()),
			expected: "test-user-123",
		},
		{
			name:     "Expired token",
			token:    createTestToken("test-user-123", time.Now().Add(-time.Hour).Unix()),
			expected: "",
		},
		{
			name:     "Token without user_id",
			token:    createTestTokenWithoutUserID(time.Now().Add(time.Hour).Unix()),
			expected: "",
		},
		{
			name:     "Invalid base64",
			token:    "invalid-base64!@#",
			expected: "",
		},
		{
			name:     "Invalid JSON",
			token:    base64.StdEncoding.EncodeToString([]byte("invalid json")),
			expected: "",
		},
		{
			name:     "No token",
			token:    "",
			expected: "",
		},
		{
			name:     "Token without Bearer prefix",
			token:    "not-bearer-token",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			result := validateTestToken(req)
			if result != tt.expected {
				t.Errorf("validateTestToken() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMiddleware_DevelopmentMode(t *testing.T) {
	// Set development environment
	os.Setenv("ENV", "development")
	os.Setenv("STORAGE_TYPE", "local")
	defer os.Unsetenv("ENV")
	defer os.Unsetenv("STORAGE_TYPE")

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := w.Write([]byte("User: " + userID)); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	// Create middleware
	middleware := Middleware(testHandler)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid test token",
			token:          createTestToken("test-user-123", time.Now().Add(time.Hour).Unix()),
			expectedStatus: http.StatusOK,
			expectedBody:   "User: test-user-123",
		},
		{
			name:           "No token",
			token:          "",
			expectedStatus: http.StatusOK,
			expectedBody:   "User: ",
		},
		{
			name:           "Invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusOK,
			expectedBody:   "User: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			middleware.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if w.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestMiddleware_HealthEndpoint(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("OK")); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	// Create middleware
	middleware := Middleware(testHandler)

	// Test health endpoint (should bypass auth)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got %q", w.Body.String())
	}
}

func TestMiddleware_SwaggerEndpoint(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("Swagger")); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	// Create middleware
	middleware := Middleware(testHandler)

	// Test swagger endpoint (should bypass auth)
	req := httptest.NewRequest("GET", "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "Swagger" {
		t.Errorf("Expected body 'Swagger', got %q", w.Body.String())
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		expected    string
		expectError bool
	}{
		{
			name:        "Valid user ID",
			ctx:         context.WithValue(context.Background(), UserIDKey, "test-user-123"),
			expected:    "test-user-123",
			expectError: false,
		},
		{
			name:        "No user ID in context",
			ctx:         context.Background(),
			expected:    "",
			expectError: true,
		},
		{
			name:        "Wrong type in context",
			ctx:         context.WithValue(context.Background(), UserIDKey, 123),
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetUserIDFromContext(tt.ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

// Helper functions
func createTestToken(userID string, exp int64) string {
	token := map[string]interface{}{
		"user_id": userID,
		"exp":     exp,
	}
	tokenBytes, _ := json.Marshal(token)
	return base64.StdEncoding.EncodeToString(tokenBytes)
}

func createTestTokenWithoutUserID(exp int64) string {
	token := map[string]interface{}{
		"exp": exp,
	}
	tokenBytes, _ := json.Marshal(token)
	return base64.StdEncoding.EncodeToString(tokenBytes)
} 