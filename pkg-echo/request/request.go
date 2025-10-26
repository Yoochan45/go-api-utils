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
