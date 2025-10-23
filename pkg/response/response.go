package response

import (
    "encoding/json"
    "log"
    "net/http"
)

// Response represents standard API response structure
type Response struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

// writeJSON writes JSON response and logs encode error server-side.
// It ensures Content-Type and status are set before encoding, and logs
// any encoding failure for server-side debugging without exposing details
// to clients.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(v); err != nil {
        // Log encode error for server-side debugging; do NOT expose details to client
        log.Printf("response encode error: %v", err)
    }
}

// Success sends a successful JSON response (200 OK)
// Use this for GET requests and successful operations
// Example:
//
//	response.Success(w, "Data retrieved", products)
func Success(w http.ResponseWriter, message string, data interface{}) {
    writeJSON(w, http.StatusOK, Response{
        Success: true,
        Message: message,
        Data:    data,
    })
}

// Created sends a resource created response (201 Created)
// Use this after successful POST/CREATE operations
// Example:
//
//	response.Created(w, "Product created", product)
func Created(w http.ResponseWriter, message string, data interface{}) {
    writeJSON(w, http.StatusCreated, Response{
        Success: true,
        Message: message,
        Data:    data,
    })
}

// NoContent sends a no content response (204 No Content)
// Use this after successful DELETE operations
// Example:
//
//	response.NoContent(w)
func NoContent(w http.ResponseWriter) {
    w.WriteHeader(http.StatusNoContent)
}

// Error sends an error response with custom status code
// Use this for general errors
// Example:
//
//	response.Error(w, http.StatusInternalServerError, "Database error")
func Error(w http.ResponseWriter, statusCode int, message string) {
    writeJSON(w, statusCode, Response{
        Success: false,
        Error:   message,
    })
}

// BadRequest sends a bad request error (400 Bad Request)
// Use this for validation errors or invalid input
// Example:
//
//	response.BadRequest(w, "Invalid product ID")
func BadRequest(w http.ResponseWriter, message string) {
    Error(w, http.StatusBadRequest, message)
}

// NotFound sends a not found error (404 Not Found)
// Use this when resource doesn't exist
// Example:
//
//	response.NotFound(w, "Product not found")
func NotFound(w http.ResponseWriter, message string) {
    Error(w, http.StatusNotFound, message)
}

// Unauthorized sends an unauthorized error (401 Unauthorized)
// Use this when authentication is required
// Example:
//
//	response.Unauthorized(w, "Invalid credentials")
func Unauthorized(w http.ResponseWriter, message string) {
    Error(w, http.StatusUnauthorized, message)
}

// Forbidden sends a forbidden error (403 Forbidden)
// Use this when user doesn't have permission
// Example:
//
//	response.Forbidden(w, "Access denied")
func Forbidden(w http.ResponseWriter, message string) {
    Error(w, http.StatusForbidden, message)
}

// InternalServerError sends internal server error (500 Internal Server Error)
// Use this for unexpected server errors
// Example:
//
//	response.InternalServerError(w, "Something went wrong")
func InternalServerError(w http.ResponseWriter, message string) {
    Error(w, http.StatusInternalServerError, message)
}
