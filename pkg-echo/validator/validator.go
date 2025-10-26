package validator

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// IsValidEmail checks if email format is valid
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsEmpty checks if string is empty or whitespace only
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// MinLength checks if string meets minimum length
func MinLength(s string, min int) bool {
	return len(strings.TrimSpace(s)) >= min
}

// ValidateRequired checks if all provided fields (key -> value) are non-empty after trimming spaces.
// Returns (true, "") if all valid, otherwise (false, "<field> is required") for the first missing field.
// Example:
//
//	ok, msg := validator.ValidateRequired(map[string]string{"email": req.Email, "password": req.Password})
func ValidateRequired(fields map[string]string) (bool, string) {
	for k, v := range fields {
		if strings.TrimSpace(v) == "" {
			return false, k + " is required"
		}
	}
	return true, ""
}
