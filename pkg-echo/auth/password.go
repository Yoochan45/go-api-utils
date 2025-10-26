package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes plain password using bcrypt
// Use this before saving password to database
// Example:
//
//	hashed, err := auth.HashPassword("mypassword123")
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// ComparePassword checks if plain password matches hashed password
// Use this during login validation
// Example:
//
//	valid := auth.ComparePassword(hashedFromDB, inputPassword)
func ComparePassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}
