package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// Test root endpoint
func TestRootHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	rootHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response Response
	json.NewDecoder(w.Body).Decode(&response)

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}

	if response.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}
}

// Test 404 on root handler
func TestRootHandler404(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/invalid", nil)
	w := httptest.NewRecorder()

	rootHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response Response
	json.NewDecoder(w.Body).Decode(&response)

	if response.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", response.Status)
	}
}

// Test /me endpoint
func TestMeHandler(t *testing.T) {
	// Set test env vars
	os.Setenv("USER_EMAIL", "test@example.com")
	os.Setenv("USER_NAME", "Test User")
	os.Setenv("USER_STACK", "Go")

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	meHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response Response
	json.NewDecoder(w.Body).Decode(&response)

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}

	if response.User == nil {
		t.Fatal("User should not be nil")
	}

	if response.User.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", response.User.Email)
	}

	if response.Fact == "" {
		t.Error("Fact should not be empty")
	}

	if response.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}
}

// Test /me endpoint - wrong method
func TestMeHandlerWrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/me", nil)
	w := httptest.NewRecorder()

	meHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// Test timestamp format
func TestTimestampFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	meHandler(w, req)

	var response Response
	json.NewDecoder(w.Body).Decode(&response)

	// Parse timestamp to verify ISO 8601 format
	_, err := time.Parse(time.RFC3339, response.Timestamp)
	if err != nil {
		t.Errorf("Timestamp not in ISO 8601 format: %v", err)
	}
}

// Test getCatFact function
func TestGetCatFact(t *testing.T) {
	fact, err := getCatFact()

	// Allow failure since it's external API
	if err != nil {
		t.Logf("Cat API failed (expected): %v", err)
		return
	}

	if fact == "" {
		t.Error("Fact should not be empty when API succeeds")
	}
}

// Test getEnvOrDefault
func TestGetEnvOrDefault(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")

	result := getEnvOrDefault("TEST_KEY", "default")
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}

	result = getEnvOrDefault("MISSING_KEY", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
}

// Test CORS headers
func TestCORSHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	meHandler(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS header not set correctly")
	}
}
