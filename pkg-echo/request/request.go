package request

import (
	"github.com/Yoochan45/go-api-utils/pkg-echo/response"
	"github.com/Yoochan45/go-api-utils/pkg-echo/validator"
	"github.com/labstack/echo/v4"
)

// BindAndValidate binds request body and validates required fields
// Example:
//
//	var req LoginRequest
//	if !request.BindAndValidate(c, &req, map[string]string{"email": req.Email, "password": req.Password}) {
//	    return nil // error response already sent
//	}
func BindAndValidate(c echo.Context, v interface{}, requiredFields map[string]string) bool {
	if err := c.Bind(v); err != nil {
		response.BadRequest(c, "invalid request body")
		return false
	}

	if valid, msg := validator.ValidateRequired(requiredFields); !valid {
		response.BadRequest(c, msg)
		return false
	}

	return true
}

// ValidateEmail validates email and sends error response if invalid
func ValidateEmail(c echo.Context, email string) bool {
	if !validator.IsValidEmail(email) {
		response.BadRequest(c, "invalid email format")
		return false
	}
	return true
}

// GetInt safely converts interface{} to int from token data
// Returns 0 if key not found or conversion fails
// Use this when getting numeric fields from custom token
// Example:
//
//	data := middleware.GetTokenData(c)
//	userID := request.GetInt(data, "user_id")
func GetInt(data map[string]interface{}, key string) int {
	if data == nil {
		return 0
	}
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	if val, ok := data[key].(int); ok {
		return val
	}
	return 0
}

// GetString safely converts interface{} to string from token data
// Returns empty string if key not found or conversion fails
// Example:
//
//	email := request.GetString(data, "email")
func GetString(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

// GetBool safely converts interface{} to bool from token data
// Returns false if key not found or conversion fails
func GetBool(data map[string]interface{}, key string) bool {
	if data == nil {
		return false
	}
	if val, ok := data[key].(bool); ok {
		return val
	}
	return false
}

// GetFloat safely converts interface{} to float64 from token data
// Returns 0.0 if key not found or conversion fails
func GetFloat(data map[string]interface{}, key string) float64 {
	if data == nil {
		return 0.0
	}
	if val, ok := data[key].(float64); ok {
		return val
	}
	return 0.0
}
