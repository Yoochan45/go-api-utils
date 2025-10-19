package request

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// ParseJSON decodes JSON request body into provided struct
// Use this to parse POST/PUT request body
// Example:
//
//	var product Product
//	if err := request.ParseJSON(r, &product); err != nil {
//	    response.BadRequest(w, "Invalid JSON")
//	    return
//	}
func ParseJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Reject unknown fields
	return decoder.Decode(v)
}

// GetIDFromURL extracts ID from URL path
// Assumes URL format: /resource/123
// Use this to get resource ID from URL
// Example:
//
//	id, err := request.GetIDFromURL(r)  // from /products/123 -> returns 123
func GetIDFromURL(r *http.Request) (int, error) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(parts) == 0 {
		return 0, strconv.ErrSyntax
	}

	idStr := parts[len(parts)-1]
	return strconv.Atoi(idStr)
}

// GetQueryParam retrieves query parameter from URL
// Use this to get query string values
// Example:
//
//	search := request.GetQueryParam(r, "search")  // from /products?search=laptop
func GetQueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// GetQueryParamInt retrieves integer query parameter from URL
// Use this for pagination, limits, etc.
// Example:
//
//	page := request.GetQueryParamInt(r, "page", 1)  // from /products?page=2
func GetQueryParamInt(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// GetPathSegment extracts specific segment from URL path
// Use this to extract path parameters
// Example:
//
//	category := request.GetPathSegment(r, 1)  // from /products/electronics/123
func GetPathSegment(r *http.Request, index int) string {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if index >= 0 && index < len(parts) {
		return parts[index]
	}
	return ""
}
