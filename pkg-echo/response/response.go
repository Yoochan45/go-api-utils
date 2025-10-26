package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Response represents standard API response structure
type Response struct {
	Success bool        `json:"success,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success sends a standardized 200 OK JSON response with message and data.
// Example:
//
//	return response.Success(c, "books retrieved", books)
func Success(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// SuccessData sends a 200 OK JSON response with raw data (no wrapper).
// Useful for list/detail endpoints where message is unnecessary.
// Example:
//
//	return response.SuccessData(c, books)
func SuccessData(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, data)
}

// Paginated sends a standardized 200 OK response with pagination metadata.
// "meta" can be any struct/map with fields like page, per_page, total, total_pages.
// Example:
//
//	meta := map[string]interface{}{"page": 1, "per_page": 10, "total": 42, "total_pages": 5}
//	return response.Paginated(c, "books retrieved", books, meta)
func Paginated(c echo.Context, message string, data interface{}, meta interface{}) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": message,
		"data":    data,
		"meta":    meta,
	})
}

// Created sends 201 Created
func Created(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// NoContent sends 204 No Content
func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// Error sends error response with custom status code
func Error(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, Response{
		Success: false,
		Error:   message,
	})
}

// BadRequest sends 400
func BadRequest(c echo.Context, message string) error {
	return Error(c, http.StatusBadRequest, message)
}

// Unauthorized sends 401
func Unauthorized(c echo.Context, message string) error {
	return Error(c, http.StatusUnauthorized, message)
}

// Forbidden sends 403
func Forbidden(c echo.Context, message string) error {
	return Error(c, http.StatusForbidden, message)
}

// NotFound sends 404
func NotFound(c echo.Context, message string) error {
	return Error(c, http.StatusNotFound, message)
}

// InternalServerError sends 500
func InternalServerError(c echo.Context, message string) error {
	return Error(c, http.StatusInternalServerError, message)
}
