package middleware

import (
	"strconv"

	"github.com/Yoochan45/go-api-utils/pkg-echo/request"
	"github.com/labstack/echo/v4"
)

// CurrentUserID returns user ID from context (custom token or basic claims), 0 if not found.
// Example:
//
//	uid := middleware.CurrentUserID(c)
func CurrentUserID(c echo.Context) uint {
	// Prefer explicit key from middleware
	if v := c.Get("user_id"); v != nil {
		switch t := v.(type) {
		case uint:
			return t
		case int:
			if t >= 0 {
				return uint(t)
			}
		case float64:
			if t >= 0 {
				return uint(t)
			}
		case string:
			if n, err := strconv.Atoi(t); err == nil && n >= 0 {
				return uint(n)
			}
		}
	}
	// Fallback to custom token data
	data := GetTokenData(c)
	return request.GetUint(data, "user_id")
}

// CurrentEmail returns email from context or empty string.
// Example:
//
//	email := middleware.CurrentEmail(c)
func CurrentEmail(c echo.Context) string {
	if v := c.Get("email"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	data := GetTokenData(c)
	return request.GetString(data, "email")
}

// CurrentRole returns role from context or empty string.
// Example:
//
//	role := middleware.CurrentRole(c)
func CurrentRole(c echo.Context) string {
	if v := c.Get("role"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	data := GetTokenData(c)
	return request.GetString(data, "role")
}
