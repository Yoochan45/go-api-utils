package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT payload structure (basic fields)
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// CustomClaims allows flexible claims with any additional fields
type CustomClaims struct {
	Data map[string]interface{} `json:"data"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// GenerateToken creates JWT token for user (basic version)
// Use this after successful login
// Example:
//
//	token, err := auth.GenerateToken(1, "user@example.com", "admin", secretKey, 24*time.Hour)
func GenerateToken(userID int, email, role, secretKey string, expiry time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// GenerateCustomToken creates JWT token with flexible data.
// Use this when you need to include custom fields (first_name, last_name, etc)
// Example:
//
//	data := map[string]interface{}{
//	    "user_id": 1,
//	    "email": "user@example.com",
//	    "first_name": "John",
//	    "last_name": "Doe",
//	    "role": "admin",
//	}
//	token, err := auth.GenerateCustomToken(data, secretKey, 24*time.Hour)
func GenerateCustomToken(data map[string]interface{}, secretKey string, expiry time.Duration) (string, error) {
	claims := &CustomClaims{
		Data: data,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ValidateToken validates JWT token and returns claims
// Use this in middleware to check token validity
// Example:
//
//	claims, err := auth.ValidateToken(tokenString, secretKey)
func ValidateToken(tokenString, secretKey string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

// ValidateCustomToken validates JWT token with custom claims
// Use this when you generated token with GenerateCustomToken
// Example:
//
//	data, err := auth.ValidateCustomToken(tokenString, secretKey)
//	userID := int(data["user_id"].(float64))
//	email := data["email"].(string)
func ValidateCustomToken(tokenString, secretKey string) (map[string]interface{}, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims.Data, nil
}

// ParseClaims extracts claims from token without validation (use with caution)
// Use this only when you already validated token in middleware
// Example:
//
//	claims := auth.ParseClaims(tokenString)
func ParseClaims(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
