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

// ValidateRequired checks if all fields are not empty
func ValidateRequired(fields map[string]string) (bool, string) {
	for name, value := range fields {
		if IsEmpty(value) {
			return false, name + " is required"
		}
	}
	return true, ""
}
