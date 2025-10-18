package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

// IP-based rate limiter
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// Create new rate limiter
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

// Get limiter for IP
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// Global rate limiter - 10 requests per minute per IP
var limiter *IPRateLimiter

func init() {
	limiter = NewIPRateLimiter(rate.Every(time.Minute/10), 10)
}

// Rate limit middleware - use the global limiter pointer
func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}

		// Check rate limit - use global limiter
		ipLimiter := limiter.GetLimiter(ip)
		if !ipLimiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(Response{
				Status:    "error",
				Message:   "rate limit exceeded",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			})
			return
		}

		next(w, r)
	}
}

func main() {
	// Load environment variables from .env file
	// Falls back to system env vars if .env doesn't exist
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env vars")
	}

	// Register route handlers
	http.HandleFunc("/", rateLimitMiddleware(rootHandler))
	http.HandleFunc("/me", rateLimitMiddleware(meHandler))

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start HTTP server
	log.Println("Server starting on:", port)
	http.ListenAndServe(":"+port, nil)
}

// Response represents the API response structure
type Response struct {
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	User      *User  `json:"user,omitempty"` // Pointer allows omitting when nil
	Timestamp string `json:"timestamp"`
	Fact      string `json:"fact,omitempty"`
}

// User represents user profile information
type User struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Stack string `json:"stack"`
}

// rootHandler handles requests to the root endpoint
func rootHandler(w http.ResponseWriter, r *http.Request) {
	// Return 404 for any path other than exactly "/"
	if r.URL.Path != "/" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		response := Response{
			Status:    "error",
			Message:   "resource not found",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Set response header for JSON
	w.Header().Set("Content-Type", "application/json")

	// Build and send welcome response
	response := Response{
		Status:    "success",
		Message:   "HNG Internship Backend Track Stage 0 Task",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// meHandler handles GET requests to /me endpoint
// Returns user profile with dynamic cat fact
func meHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")

	// Fetch cat fact from external API
	catFact, err := getCatFact()
	if err != nil {
		log.Printf("Cat API error: %v", err)
		catFact = "Cat fact unavailable" // Fallback message
	}

	// Build response with user data and cat fact
	response := Response{
		Status: "success",
		User: &User{
			Email: getEnvOrDefault("USER_EMAIL", "default@example.com"),
			Name:  getEnvOrDefault("USER_NAME", "Unknown"),
			Stack: getEnvOrDefault("USER_STACK", "Go"),
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339), // Current UTC time
		Fact:      catFact,
	}

	json.NewEncoder(w).Encode(response)
}

// getCatFact fetches a random cat fact from external API
func getCatFact() (string, error) {
	// Create HTTP client with 5 second timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make GET request to cat facts API
	resp, err := client.Get("https://catfact.ninja/fact")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() // Ensure connection is closed

	// Check for successful status code
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("cat API returned status %d", resp.StatusCode)
	}

	// Parse JSON response
	var result struct {
		Fact string `json:"fact"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	return result.Fact, nil
}

// getEnvOrDefault retrieves environment variable or returns default value
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
