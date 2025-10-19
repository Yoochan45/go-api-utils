package middleware

import (
	"log"
	"net/http"
	"time"
)

// CORS adds Cross-Origin Resource Sharing headers
// Use this to allow frontend to access your API
// Example:
//
//	handler := middleware.CORS(mux)
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logger logs HTTP requests with method, path, and duration
// Use this to monitor API requests
// Example:
//
//	handler := middleware.Logger(mux)
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log request
		log.Printf("==> [%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Call next handler
		next.ServeHTTP(w, r)

		// Log completion
		duration := time.Since(start)
		log.Printf("Completed in %v", duration)
	})
}
