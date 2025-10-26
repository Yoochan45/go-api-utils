package request

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/Yoochan45/go-api-utils/pkg-echo/response"
	"github.com/Yoochan45/go-api-utils/pkg-echo/validator"
	"github.com/labstack/echo/v4"
)

// BindAndValidate binds request body and validates required fields.
// Deprecated: This approach is error-prone because the "requiredFields" map
// is evaluated BEFORE c.Bind fills the struct, often resulting in zero-values
// being validated. Use BindAndRequireFields instead.
//
// Example (NOT recommended):
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

// BindAndRequireFields binds JSON request body into v and validates required JSON fields
// by their json tag names (e.g., "email", "password"). This avoids the zero-value pitfall
// of passing a map before binding.
// Example:
//
//	var req LoginRequest
//	if !request.BindAndRequireFields(c, &req, "email", "password") {
//	    return nil // error response already sent
//	}
func BindAndRequireFields(c echo.Context, v interface{}, requiredJSON ...string) bool {
	if err := c.Bind(v); err != nil {
		response.BadRequest(c, "invalid request body")
		return false
	}

	if ok, msg := RequireFields(v, requiredJSON...); !ok {
		response.BadRequest(c, msg)
		return false
	}
	return true
}

// RequireFields validates required JSON fields on an already-bound struct.
// It does not write any HTTP response, only returns (ok, message) so you can decide
// how to handle the error in higher layers.
// Example:
//
//	var req LoginRequest
//	if err := c.Bind(&req); err != nil { return response.BadRequest(c, "invalid body") }
//	if ok, msg := request.RequireFields(&req, "email", "password"); !ok {
//	    return response.BadRequest(c, msg)
//	}
func RequireFields(v interface{}, requiredJSON ...string) (bool, string) {
	if len(requiredJSON) == 0 {
		return true, ""
	}

	// Collect string fields by json tag name
	values := map[string]string{}
	rv := reflect.Indirect(reflect.ValueOf(v))
	if rv.IsValid() && rv.Kind() == reflect.Struct {
		rt := rv.Type()
		for i := 0; i < rt.NumField(); i++ {
			sf := rt.Field(i)
			tag := sf.Tag.Get("json")
			if tag == "" || tag == "-" {
				continue
			}
			name := strings.Split(tag, ",")[0]
			fv := rv.Field(i)
			// Only validate string fields. Extend as needed.
			if fv.IsValid() && fv.Kind() == reflect.String {
				values[name] = fv.String()
			}
		}
	}

	fieldsToCheck := map[string]string{}
	for _, key := range requiredJSON {
		fieldsToCheck[key] = values[key]
	}

	return validator.ValidateRequired(fieldsToCheck)
}

// ValidateEmail validates email and sends error response if invalid
// Example:
//
//	if !request.ValidateEmail(c, req.Email) {
//	    return nil // error response already sent
//	}
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
// Example:
//
//	isAdmin := request.GetBool(data, "is_admin")
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
// Example:
//
//	price := request.GetFloat(data, "price")
func GetFloat(data map[string]interface{}, key string) float64 {
	if data == nil {
		return 0.0
	}
	if val, ok := data[key].(float64); ok {
		return val
	}
	return 0.0
}

// GetUint safely converts interface{} to uint from token data (or 0 on failure).
// Useful when your IDs are unsigned in DB/models.
// Example:
//
//	uid := request.GetUint(data, "user_id")
func GetUint(data map[string]interface{}, key string) uint {
	if data == nil {
		return 0
	}
	// JWT numbers are float64 by default
	if v, ok := data[key].(float64); ok && v >= 0 {
		return uint(v)
	}
	if v, ok := data[key].(int); ok && v >= 0 {
		return uint(v)
	}
	if v, ok := data[key].(uint); ok {
		return v
	}
	return 0
}

// QueryString returns query param as string with default fallback.
// Example:
//
//	q := request.QueryString(c, "search", "")
func QueryString(c echo.Context, key, def string) string {
	v := c.QueryParam(key)
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

// QueryInt returns query param as int with default fallback.
// Example:
//
//	page := request.QueryInt(c, "page", 1)
func QueryInt(c echo.Context, key string, def int) int {
	v := strings.TrimSpace(c.QueryParam(key))
	if v == "" {
		return def
	}
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	return def
}

// PathParamUint parses a path param (e.g., :id) into uint, 0 if invalid.
// Example:
//
//	id := request.PathParamUint(c, "id")
func PathParamUint(c echo.Context, key string) uint {
	v := strings.TrimSpace(c.Param(key))
	if v == "" {
		return 0
	}
	if n, err := strconv.Atoi(v); err == nil && n >= 0 {
		return uint(n)
	}
	return 0
}
