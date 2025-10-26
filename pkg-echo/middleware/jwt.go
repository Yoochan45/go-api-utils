package middleware

import (
	"net/http"
	"strings"

	"github.com/Yoochan45/go-api-utils/pkg-echo/auth"
	"github.com/labstack/echo/v4"
)

// JWTConfig holds JWT middleware configuration
type JWTConfig struct {
	SecretKey   string
	SkipperFunc func(c echo.Context) bool // optional: skip auth for certain routes
}

// JWTMiddleware validates JWT token from Authorization header
// Use this to protect routes that require authentication
// Token should be in format: "Bearer <token>"
// Example:
//
//	e.Use(middleware.JWTMiddleware(middleware.JWTConfig{SecretKey: "your-secret"}))
//	// or for specific route group:
//	api := e.Group("/api", middleware.JWTMiddleware(cfg))
func JWTMiddleware(config JWTConfig) echo.MiddlewareFunc {
	if config.SecretKey == "" {
		panic("JWT secret key cannot be empty")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// skip if skipper func provided and returns true
			if config.SkipperFunc != nil && config.SkipperFunc(c) {
				return next(c)
			}

			// extract token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "missing authorization header",
				})
			}

			// check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid authorization header format",
				})
			}

			tokenString := parts[1]

			// validate token
			claims, err := auth.ValidateToken(tokenString, config.SecretKey)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid or expired token",
				})
			}

			// store claims in context for handler access
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)
			c.Set("claims", claims)

			return next(c)
		}
	}
}

// GetUserID retrieves user ID from Echo context (set by JWTMiddleware)
// Use this in your handlers to get authenticated user ID
// Example:
//
//	userID := middleware.GetUserID(c)
func GetUserID(c echo.Context) int {
	if id, ok := c.Get("user_id").(int); ok {
		return id
	}
	return 0
}

// GetUserEmail retrieves email from Echo context
func GetUserEmail(c echo.Context) string {
	if email, ok := c.Get("email").(string); ok {
		return email
	}
	return ""
}

// GetUserRole retrieves role from Echo context
func GetUserRole(c echo.Context) string {
	if role, ok := c.Get("role").(string); ok {
		return role
	}
	return ""
}

// GetClaims retrieves full claims from Echo context
func GetClaims(c echo.Context) *auth.Claims {
	if claims, ok := c.Get("claims").(*auth.Claims); ok {
		return claims
	}
	return nil
}
