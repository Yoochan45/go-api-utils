# Echo Framework Integration Guide

This guide demonstrates how to integrate `go-api-utils/pkg-echo` utilities into your Echo-based REST API project.

## Overview

The `pkg-echo` package provides ready-to-use utilities for common REST API patterns:
- JWT authentication and middleware
- Password hashing with bcrypt
- GORM database helpers
- Request validation
- Standardized JSON responses

## Installation

```bash
go get github.com/Yoochan45/go-api-utils
```

---

## Integration Patterns

### Database Setup

```go
package main

import (
    "github.com/Yoochan45/go-api-utils/pkg/config"
    "github.com/Yoochan45/go-api-utils/pkg-echo/orm"
    "gorm.io/gorm"
)

func initDatabase() *gorm.DB {
    cfg := config.LoadEnv()
    
    db, err := orm.ConnectGORM(cfg.DatabaseURL)
    if err != nil {
        panic(err)
    }
    
    // Auto-migrate your models
    orm.AutoMigrate(db, &User{}, &Product{})
    
    return db
}
```

### Authentication Service

```go
package service

import (
    "time"
    "github.com/Yoochan45/go-api-utils/pkg-echo/auth"
    "gorm.io/gorm"
)

type AuthService struct {
    db        *gorm.DB
    jwtSecret string
}

func (s *AuthService) Register(email, password string) error {
    hashedPassword, err := auth.HashPassword(password)
    if err != nil {
        return err
    }
    
    user := User{Email: email, Password: hashedPassword}
    return s.db.Create(&user).Error
}

func (s *AuthService) Login(email, password string) (string, error) {
    var user User
    if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
        return "", err
    }
    
    if !auth.ComparePassword(user.Password, password) {
        return "", errors.New("invalid credentials")
    }
    
    // Generate JWT with custom claims
    token, err := auth.GenerateCustomToken(map[string]interface{}{
        "user_id": user.ID,
        "email":   user.Email,
        "role":    user.Role,
    }, s.jwtSecret, 24*time.Hour)
    
    return token, err
}
```

### HTTP Handlers

```go
package handler

import (
    "github.com/Yoochan45/go-api-utils/pkg-echo/request"
    "github.com/Yoochan45/go-api-utils/pkg-echo/response"
    "github.com/labstack/echo/v4"
)

type AuthHandler struct {
    service *AuthService
}

func (h *AuthHandler) Register(c echo.Context) error {
    var req RegisterRequest
    
    if !request.BindAndValidate(c, &req, map[string]string{
        "email":    req.Email,
        "password": req.Password,
    }) {
        return nil // error response already sent
    }
    
    if !request.ValidateEmail(c, req.Email) {
        return nil
    }
    
    if err := h.service.Register(req.Email, req.Password); err != nil {
        return response.BadRequest(c, err.Error())
    }
    
    return response.Created(c, "user registered successfully", nil)
}

func (h *AuthHandler) Login(c echo.Context) error {
    var req LoginRequest
    
    if !request.BindAndValidate(c, &req, map[string]string{
        "email":    req.Email,
        "password": req.Password,
    }) {
        return nil
    }
    
    token, err := h.service.Login(req.Email, req.Password)
    if err != nil {
        return response.Unauthorized(c, "invalid credentials")
    }
    
    return response.Success(c, "login successful", map[string]string{
        "token": token,
    })
}
```

### Protected Routes

```go
package handler

import (
    "github.com/Yoochan45/go-api-utils/pkg-echo/middleware"
    "github.com/Yoochan45/go-api-utils/pkg-echo/request"
    "github.com/Yoochan45/go-api-utils/pkg-echo/response"
    "github.com/labstack/echo/v4"
)

func (h *ProfileHandler) GetProfile(c echo.Context) error {
    // Extract authenticated user data from JWT
    tokenData := middleware.GetTokenData(c)
    
    userID := request.GetInt(tokenData, "user_id")
    email := request.GetString(tokenData, "email")
    
    return response.Success(c, "profile retrieved", map[string]interface{}{
        "user_id": userID,
        "email":   email,
    })
}
```

### Route Registration

```go
package main

import (
    "github.com/Yoochan45/go-api-utils/pkg-echo/middleware"
    "github.com/labstack/echo/v4"
)

func setupRoutes(e *echo.Echo, jwtSecret string) {
    // Public routes
    e.POST("/auth/register", authHandler.Register)
    e.POST("/auth/login", authHandler.Login)
    
    // Protected routes (require authentication)
    api := e.Group("/api")
    api.Use(middleware.JWTMiddleware(middleware.JWTConfig{
        SecretKey:      jwtSecret,
        UseCustomToken: true,
    }))
    
    api.GET("/profile", profileHandler.GetProfile)
    api.GET("/products", productHandler.List)
}
```

### Complete Server Setup

```go
package main

import (
    "github.com/Yoochan45/go-api-utils/pkg/config"
    "github.com/Yoochan45/go-api-utils/pkg-echo/orm"
    "github.com/labstack/echo/v4"
    echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
    cfg := config.LoadEnv()
    
    // Initialize database
    db, err := orm.ConnectGORM(cfg.DatabaseURL)
    if err != nil {
        panic(err)
    }
    
    // Auto-migrate models
    orm.AutoMigrate(db, &User{}, &Product{})
    
    // Initialize Echo
    e := echo.New()
    
    // Global middleware
    e.Use(echoMiddleware.Logger())
    e.Use(echoMiddleware.Recover())
    e.Use(echoMiddleware.CORS())
    
    // Setup routes
    setupRoutes(e, cfg.JWTSecret)
    
    // Start server
    e.Logger.Fatal(e.Start(":" + cfg.Port))
}
```

---

## Transaction Handling

Use `orm.WithTransaction` for operations that require atomicity:

```go
import "github.com/Yoochan45/go-api-utils/pkg-echo/orm"

func (s *OrderService) CreateOrder(userID int, items []Item) error {
    return orm.WithTransaction(s.db, func(tx *gorm.DB) error {
        // Create order
        order := Order{UserID: userID}
        if err := tx.Create(&order).Error; err != nil {
            return err
        }
        
        // Create order items
        for _, item := range items {
            orderItem := OrderItem{OrderID: order.ID, ProductID: item.ID}
            if err := tx.Create(&orderItem).Error; err != nil {
                return err
            }
        }
        
        return nil
    })
}
```

---

## Custom Validation

```go
import "github.com/Yoochan45/go-api-utils/pkg-echo/validator"

func validateUser(user *User) error {
    if !validator.IsValidEmail(user.Email) {
        return errors.New("invalid email format")
    }
    
    if !validator.MinLength(user.Password, 8) {
        return errors.New("password must be at least 8 characters")
    }
    
    return nil
}
```

---

## API Response Format

All response helpers follow this consistent format:

### Success Response
```json
{
  "success": true,
  "message": "operation successful",
  "data": { /* your data */ }
}
```

### Error Response
```json
{
  "success": false,
  "error": "error message"
}
```

---

## Available Response Helpers

- `response.Success(c, message, data)` - 200 OK
- `response.Created(c, message, data)` - 201 Created
- `response.NoContent(c)` - 204 No Content
- `response.BadRequest(c, message)` - 400 Bad Request
- `response.Unauthorized(c, message)` - 401 Unauthorized
- `response.Forbidden(c, message)` - 403 Forbidden
- `response.NotFound(c, message)` - 404 Not Found
- `response.InternalServerError(c, message)` - 500 Internal Server Error

---

## Request Validation Helpers

- `request.BindAndValidate(c, &struct, requiredFields)` - Bind and validate in one call
- `request.ValidateEmail(c, email)` - Validate email format
- `request.GetInt(data, key)` - Safely extract int from map
- `request.GetString(data, key)` - Safely extract string from map
- `request.GetBool(data, key)` - Safely extract bool from map
- `request.GetFloat(data, key)` - Safely extract float64 from map

---

## Environment Variables

Create a `.env` file in your project root:

```env
DATABASE_URL=postgresql://user:password@localhost:5432/dbname?sslmode=disable
PORT=8080
JWT_SECRET=your-secret-key-here
```

---

## Project Structure Example

```
your-project/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── handler/             # HTTP handlers
│   ├── service/             # Business logic
│   ├── repository/          # Data access layer
│   └── model/               # Domain models
├── .env
└── go.mod
```

---

## Dependencies

```bash
go get github.com/labstack/echo/v4
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto
```

---

## Further Reading

- [Main Documentation](../../README.md)
- [Echo Framework](https://echo.labstack.com/)
- [GORM Documentation](https://gorm.io/)
- [JWT Best Practices](https://jwt.io/introduction)

---

**License:** MIT  
**Author:** Aisiya Qutwatunnada (@Yoochan45)