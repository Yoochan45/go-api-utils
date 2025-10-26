package auth

import (
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

const defaultCost = bcrypt.DefaultCost

// HashPassword hashes plain text password using bcrypt.
// BCRYPT_COST can override default cost (env var).
// Example:
//
//	hashed, err := auth.HashPassword("secret")
func HashPassword(password string) (string, error) {
	cost := defaultCost
	if v := os.Getenv("BCRYPT_COST"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= bcrypt.MinCost && n <= bcrypt.MaxCost {
			cost = n
		}
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

// ComparePassword compares bcrypt hashed password with plain text.
// Example:
//
//	ok := auth.ComparePassword(user.Password, "secret")
func ComparePassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}
