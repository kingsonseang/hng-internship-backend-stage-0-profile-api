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

// Test rate limiting
func TestRateLimiting(t *testing.T) {
	// Create limiter: 1 req/sec, burst of 1
	testLimiter := NewIPRateLimiter(1, 1)
	limiter = testLimiter

	testIP := "192.168.1.1"

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/me", nil)
	req1.RemoteAddr = testIP
	w1 := httptest.NewRecorder()
	rateLimitMiddleware(meHandler)(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First request: Expected status 200, got %d", w1.Code)
	}

	// Second request immediately - should fail
	req2 := httptest.NewRequest(http.MethodGet, "/me", nil)
	req2.RemoteAddr = testIP
	w2 := httptest.NewRecorder()
	rateLimitMiddleware(meHandler)(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 (rate limited), got %d", w2.Code)
	}

	var response Response
	json.NewDecoder(w2.Body).Decode(&response)

	if response.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", response.Status)
	}

	if response.Message != "rate limit exceeded" {
		t.Errorf("Expected message 'rate limit exceeded', got '%s'", response.Message)
	}
}

// Test different IPs have separate limits
func TestRateLimitingPerIP(t *testing.T) {
	// Create NEW limiter for this test
	testLimiter := NewIPRateLimiter(1, 1)
	limiter = testLimiter

	// IP 1 makes request
	req1 := httptest.NewRequest(http.MethodGet, "/me", nil)
	req1.RemoteAddr = "192.168.1.1"
	w1 := httptest.NewRecorder()
	rateLimitMiddleware(meHandler)(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("IP1 first request: Expected 200, got %d", w1.Code)
	}

	// IP 2 makes request (should succeed - different IP, different limiter)
	req2 := httptest.NewRequest(http.MethodGet, "/me", nil)
	req2.RemoteAddr = "192.168.1.2"
	w2 := httptest.NewRecorder()
	rateLimitMiddleware(meHandler)(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("IP2 first request: Expected 200, got %d", w2.Code)
	}

	// IP 1 makes second request immediately (should fail)
	req3 := httptest.NewRequest(http.MethodGet, "/me", nil)
	req3.RemoteAddr = "192.168.1.1"
	w3 := httptest.NewRecorder()
	rateLimitMiddleware(meHandler)(w3, req3)

	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("IP1 second request: Expected 429, got %d", w3.Code)
	}
}

// Test X-Forwarded-For header
func TestRateLimitingXForwardedFor(t *testing.T) {
	// Create NEW limiter for this test
	testLimiter := NewIPRateLimiter(1, 1)
	limiter = testLimiter

	// First request with X-Forwarded-For
	req1 := httptest.NewRequest(http.MethodGet, "/me", nil)
	req1.RemoteAddr = "proxy.ip"
	req1.Header.Set("X-Forwarded-For", "192.168.1.100")
	w1 := httptest.NewRecorder()
	rateLimitMiddleware(meHandler)(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First request: Expected 200, got %d", w1.Code)
	}

	// Second request from same X-Forwarded-For IP
	req2 := httptest.NewRequest(http.MethodGet, "/me", nil)
	req2.RemoteAddr = "proxy.ip"
	req2.Header.Set("X-Forwarded-For", "192.168.1.100")
	w2 := httptest.NewRecorder()
	rateLimitMiddleware(meHandler)(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Second request: Expected 429, got %d", w2.Code)
	}
}
