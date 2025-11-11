package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/yoockh/go-api-utils/pkg-echo/auth"
	"github.com/yoockh/go-api-utils/pkg-echo/response"
)

// JWTConfig configures JWT middleware behavior.
// Example:
//
//	api := e.Group("/api")
//	api.Use(middleware.JWTMiddleware(middleware.JWTConfig{SecretKey: "secret", UseCustomToken: true}))
type JWTConfig struct {
	SecretKey      string
	UseCustomToken bool
	SkipperFunc    func(c echo.Context) bool
}

// JWTMiddleware validates Bearer token from Authorization header and injects claims into context.
// For custom token: stores map data under "token_data".
// For basic token: stores user_id, email, role, and "claims".
func JWTMiddleware(config JWTConfig) echo.MiddlewareFunc {
	if config.SecretKey == "" {
		panic("JWT secret key cannot be empty")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.SkipperFunc != nil && config.SkipperFunc(c) {
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return response.Unauthorized(c, "missing authorization header")
			}
			parts := strings.Fields(authHeader)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return response.Unauthorized(c, "invalid authorization header format")
			}
			tokenString := parts[1]

			if config.UseCustomToken {
				data, err := auth.ValidateCustomToken(tokenString, config.SecretKey)
				if err != nil {
					if err == auth.ErrExpiredToken {
						return response.Unauthorized(c, "token expired")
					}
					return response.Unauthorized(c, "invalid token")
				}
				c.Set("token_data", data)
				// Convenience extractions (if present)
				if v, ok := data["user_id"]; ok {
					c.Set("user_id", v)
				}
				if v, ok := data["email"]; ok {
					c.Set("email", v)
				}
				if v, ok := data["role"]; ok {
					c.Set("role", v)
				}
			} else {
				claims, err := auth.ValidateToken(tokenString, config.SecretKey)
				if err != nil {
					if err == auth.ErrExpiredToken {
						return response.Unauthorized(c, "token expired")
					}
					return response.Unauthorized(c, "invalid token")
				}
				c.Set("claims", claims)
				c.Set("user_id", claims.UserID)
				c.Set("email", claims.Email)
				if claims.Role != "" {
					c.Set("role", claims.Role)
				}
			}

			return next(c)
		}
	}
}

// GetTokenData returns custom token data from context or empty map if not present.
// Example:
//
//	data := middleware.GetTokenData(c)
//	userID := request.GetInt(data, "user_id")
func GetTokenData(c echo.Context) map[string]interface{} {
	v := c.Get("token_data")
	if v == nil {
		return map[string]interface{}{}
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}
