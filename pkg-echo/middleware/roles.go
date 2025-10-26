package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
)

// RequireRoles allows only requests whose role is included in the allowed list.
// It reads "role" from context keys set by JWTMiddleware (custom token or basic claims).
// Example:
//
//	api := e.Group("/api")
//	api.Use(middleware.JWTMiddleware(middleware.JWTConfig{SecretKey: "secret", UseCustomToken: true}))
//	api.GET("/admin/stats", adminHandler, middleware.RequireRoles("admin"))
func RequireRoles(allowed ...string) echo.MiddlewareFunc {
	set := map[string]struct{}{}
	for _, r := range allowed {
		set[strings.ToLower(strings.TrimSpace(r))] = struct{}{}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role := CurrentRole(c)
			if _, ok := set[strings.ToLower(role)]; !ok {
				return c.NoContent(403)
			}
			return next(c)
		}
	}
}
