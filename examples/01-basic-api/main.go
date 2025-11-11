package main

import (
	"log"
	"net/http"

	"github.com/yoockh/go-api-utils/pkg/middleware"
	"github.com/yoockh/go-api-utils/pkg/response"
)

func main() {
	mux := http.NewServeMux()

	// Simple health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, "API is running", map[string]string{
			"status": "healthy",
		})
	})

	// Example error endpoint
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		response.InternalServerError(w, "This is a test error")
	})

	// Apply middleware
	handler := middleware.Logger(middleware.CORS(mux))

	// Start server
	port := "8080"
	log.Printf("🚀 Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
