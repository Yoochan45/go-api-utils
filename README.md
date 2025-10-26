# Go API Utils

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

A lightweight utility library for building REST APIs in Go, supporting both standard `net/http` and Echo framework.

## Features

### Standard Library (pkg/)
- Database connection helpers for PostgreSQL with connection pooling
- Standardized JSON response utilities
- Request parsing and URL parameter extraction
- SQL query builders for common CRUD operations
- CORS and request logging middleware
- Environment variable configuration loader

### Echo Framework (pkg-echo/)
- JWT authentication with custom claims support
- Password hashing using bcrypt
- GORM ORM connection helpers and transaction wrapper
- Input validation (email format, required fields, string length)
- Request binding and validation utilities
- Echo-specific JSON response helpers

## Installation

```bash
go get github.com/Yoochan45/go-api-utils
```

## Quick Start

### Standard net/http API

```go
package main

import (
    "net/http"
    "github.com/Yoochan45/go-api-utils/pkg/database"
    "github.com/Yoochan45/go-api-utils/pkg/middleware"
    "github.com/Yoochan45/go-api-utils/pkg/response"
)

func main() {
    // Connect to database
    db, _ := database.ConnectPostgresURL("postgresql://user:pass@localhost/db")
    defer database.Close(db)
    
    // Setup routes
    mux := http.NewServeMux()
    mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
        rows, _ := db.Query("SELECT id, name, price FROM products")
        defer rows.Close()
        
        var products []Product
        for rows.Next() {
            var p Product
            rows.Scan(&p.ID, &p.Name, &p.Price)
            products = append(products, p)
        }
        
        response.Success(w, "products retrieved", products)
    })
    
    // Add middleware
    handler := middleware.Logger(middleware.CORS(mux))
    http.ListenAndServe(":8080", handler)
}
```

### Echo Framework with JWT

```go
package main

import (
    "time"
    "github.com/Yoochan45/go-api-utils/pkg/config"
    "github.com/Yoochan45/go-api-utils/pkg-echo/auth"
    "github.com/Yoochan45/go-api-utils/pkg-echo/middleware"
    "github.com/Yoochan45/go-api-utils/pkg-echo/orm"
    "github.com/Yoochan45/go-api-utils/pkg-echo/request"
    "github.com/Yoochan45/go-api-utils/pkg-echo/response"
    "github.com/labstack/echo/v4"
)

func main() {
    cfg := config.LoadEnv()
    
    // Connect GORM
    db, _ := orm.ConnectGORM(cfg.DatabaseURL)
    orm.AutoMigrate(db, &User{})
    
    e := echo.New()
    
    // Public routes
    e.POST("/login", handleLogin)
    
    // Protected routes
    api := e.Group("/api")
    api.Use(middleware.JWTMiddleware(middleware.JWTConfig{
        SecretKey:      "your-secret",
        UseCustomToken: true,
    }))
    api.GET("/profile", handleProfile)
    
    e.Start(":8080")
}

func handleLogin(c echo.Context) error {
    var req LoginRequest
    if !request.BindAndValidate(c, &req, map[string]string{
        "email": req.Email, "password": req.Password,
    }) {
        return nil
    }
    
    var user User
    db.Where("email = ?", req.Email).First(&user)
    
    if !auth.ComparePassword(user.Password, req.Password) {
        return response.Unauthorized(c, "invalid credentials")
    }
    
    token, _ := auth.GenerateCustomToken(map[string]interface{}{
        "user_id": user.ID,
        "email": user.Email,
    }, "your-secret", 24*time.Hour)
    
    return response.Success(c, "login success", map[string]string{"token": token})
}

func handleProfile(c echo.Context) error {
    data := middleware.GetTokenData(c)
    userID := request.GetInt(data, "user_id")
    
    return response.Success(c, "profile", map[string]interface{}{
        "user_id": userID,
        "email": request.GetString(data, "email"),
    })
}
```

## Package Documentation

### pkg/database

Connect to PostgreSQL database with automatic connection pooling.

```go
// Using DATABASE_URL environment variable
db, err := database.ConnectPostgresURL(os.Getenv("DATABASE_URL"))

// Using individual config parameters
db, err := database.ConnectPostgres(database.PostgresConfig{
    Host:     "localhost",
    Port:     "5432",
    User:     "postgres",
    Password: "secret",
    DBName:   "mydb",
    SSLMode:  "disable",
})

// Auto-initialize (checks SKIP_DB env variable)
cfg := config.LoadEnv()
db, err := database.Init(cfg)

// Close connection
database.Close(db)
```

**Available Functions:**
- `ConnectPostgres(config)` - Connect using config struct
- `ConnectPostgresURL(url)` - Connect using DATABASE_URL string
- `MustConnect(config)` - Connect or panic on error
- `Init(config)` - Initialize with SKIP_DB support
- `Close(db)` - Close database connection

---

### pkg/response

Send standardized JSON responses for net/http handlers.

```go
// Success response (200 OK)
response.Success(w, "operation successful", data)

// Created response (201 Created)
response.Created(w, "resource created", newResource)

// No content (204 No Content)
response.NoContent(w)

// Error responses
response.BadRequest(w, "invalid input")          // 400
response.Unauthorized(w, "authentication failed") // 401
response.Forbidden(w, "access denied")            // 403
response.NotFound(w, "resource not found")        // 404
response.InternalServerError(w, "server error")   // 500
```

**Available Functions:**
- `Success(w, message, data)` - 200 OK response
- `Created(w, message, data)` - 201 Created response
- `NoContent(w)` - 204 No Content
- `BadRequest(w, message)` - 400 Bad Request
- `Unauthorized(w, message)` - 401 Unauthorized
- `Forbidden(w, message)` - 403 Forbidden
- `NotFound(w, message)` - 404 Not Found
- `InternalServerError(w, message)` - 500 Internal Server Error

---

### pkg/request

Parse and extract data from HTTP requests.

```go
// Parse JSON body
var user User
if err := request.ParseJSON(r, &user); err != nil {
    response.BadRequest(w, "invalid JSON")
}

// Extract ID from URL path (/users/123)
id, err := request.GetIDFromURL(r)

// Get query parameters with defaults
page := request.GetQueryParamInt(r, "page", 1)
search := request.GetQueryParam(r, "search", "")

// Extract path segment
segment := request.GetPathSegment(r.URL.Path, 2) // /api/users/123 -> "123"
```

**Available Functions:**
- `ParseJSON(r, v)` - Parse request body as JSON
- `GetIDFromURL(r)` - Extract numeric ID from URL path
- `GetQueryParam(r, key, defaultValue)` - Get string query parameter
- `GetQueryParamInt(r, key, defaultValue)` - Get integer query parameter
- `GetPathSegment(path, index)` - Extract path segment by index

---

### pkg/repository

Build SQL queries for common database operations.

```go
// Build INSERT query
query, args := repository.BuildInsertQuery("users", map[string]interface{}{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30,
})
db.Exec(query, args...)

// Build UPDATE query
query, args := repository.BuildUpdateQuery("users", 123, map[string]interface{}{
    "name": "Jane Doe",
    "email": "jane@example.com",
})
db.Exec(query, args...)

// Build SELECT query
query := repository.BuildSelectQuery("users", 
    []string{"id", "name", "email"}, 
    "age > 18 AND active = true",
)
rows, _ := db.Query(query)

// Check if rows were affected
result, _ := db.Exec(query, args...)
if !repository.CheckRowsAffected(result) {
    // No rows updated
}
```

**Available Functions:**
- `BuildInsertQuery(table, data)` - Generate INSERT statement
- `BuildUpdateQuery(table, id, data)` - Generate UPDATE statement
- `BuildSelectQuery(table, columns, where)` - Generate SELECT statement
- `CheckRowsAffected(result)` - Check if any rows were affected
- `ScanRows(rows, dest)` - Scan multiple rows into slice

---

### pkg/middleware

HTTP middleware for CORS and logging.

```go
mux := http.NewServeMux()
mux.HandleFunc("/api/users", handleUsers)

// Add CORS headers
handler := middleware.CORS(mux)

// Add request logging
handler = middleware.Logger(handler)

http.ListenAndServe(":8080", handler)
```

**Available Functions:**
- `CORS(next)` - Add CORS headers (allow all origins)
- `Logger(next)` - Log HTTP requests with method, path, and duration

---

### pkg/config

Load environment variables from .env file.

```go
// Load config (returns default if env not set)
cfg := config.LoadEnv()
fmt.Println(cfg.DatabaseURL)
fmt.Println(cfg.Port) // default: "8080"

// Load config or panic
cfg := config.MustLoadEnv()
```

**Available Functions:**
- `LoadEnv()` - Load environment variables with defaults
- `MustLoadEnv()` - Load environment variables or panic

**Supported Environment Variables:**
- `DATABASE_URL` - PostgreSQL connection string
- `PORT` - Server port (default: 8080)
- `SKIP_DB` - Skip database connection (1 to enable)

---

### pkg-echo/auth

JWT token generation/validation and password hashing.

```go
// Hash password before saving to database
hashedPassword, err := auth.HashPassword("user-password")

// Compare password during login
isValid := auth.ComparePassword(hashedPassword, "user-password")

// Generate JWT token with custom claims
token, err := auth.GenerateCustomToken(map[string]interface{}{
    "user_id": 123,
    "email": "user@example.com",
    "role": "admin",
}, "secret-key", 24*time.Hour)

// Validate JWT token
data, err := auth.ValidateCustomToken(tokenString, "secret-key")
userID := int(data["user_id"].(float64))
```

**Available Functions:**
- `HashPassword(password)` - Hash password using bcrypt
- `ComparePassword(hashed, plain)` - Verify password against hash
- `GenerateToken(userID, email, role, secret, expiry)` - Generate basic JWT token
- `GenerateCustomToken(data, secret, expiry)` - Generate JWT with custom claims
- `ValidateToken(token, secret)` - Validate basic JWT token
- `ValidateCustomToken(token, secret)` - Validate custom JWT token
- `ParseClaims(token)` - Parse claims without validation

---

### pkg-echo/middleware

JWT authentication middleware for Echo framework.

```go
e := echo.New()

// Public routes (no authentication)
e.POST("/login", handleLogin)

// Protected routes (require JWT token)
api := e.Group("/api")
api.Use(middleware.JWTMiddleware(middleware.JWTConfig{
    SecretKey:      "your-secret-key",
    UseCustomToken: true,
}))
api.GET("/profile", handleProfile)
api.POST("/products", handleCreateProduct)

// Access user data in handler
func handleProfile(c echo.Context) error {
    tokenData := middleware.GetTokenData(c)
    userID := request.GetInt(tokenData, "user_id")
    email := request.GetString(tokenData, "email")
    
    return response.Success(c, "profile", map[string]interface{}{
        "user_id": userID,
        "email": email,
    })
}
```

**Available Functions:**
- `JWTMiddleware(config)` - Validate JWT from Authorization header
- `GetUserID(c)` - Get user ID from context (for basic tokens)
- `GetUserEmail(c)` - Get email from context (for basic tokens)
- `GetUserRole(c)` - Get role from context (for basic tokens)
- `GetClaims(c)` - Get full claims from context (for basic tokens)
- `GetTokenData(c)` - Get custom token data as map (for custom tokens)

---

### pkg-echo/response

Standardized JSON responses for Echo handlers.

```go
// Success responses
response.Success(c, "operation successful", data)      // 200 OK with wrapper
response.SuccessData(c, data)                          // 200 OK without wrapper
response.Created(c, "resource created", data)          // 201 Created
response.NoContent(c)                                  // 204 No Content

// Error responses
response.BadRequest(c, "invalid input")                // 400 Bad Request
response.Unauthorized(c, "authentication required")    // 401 Unauthorized
response.Forbidden(c, "access denied")                 // 403 Forbidden
response.NotFound(c, "resource not found")             // 404 Not Found
response.InternalServerError(c, "server error")        // 500 Internal Server Error
```

**Available Functions:**
- `Success(c, message, data)` - 200 OK with success wrapper
- `SuccessData(c, data)` - 200 OK without wrapper
- `Created(c, message, data)` - 201 Created
- `NoContent(c)` - 204 No Content
- `BadRequest(c, message)` - 400 Bad Request
- `Unauthorized(c, message)` - 401 Unauthorized
- `Forbidden(c, message)` - 403 Forbidden
- `NotFound(c, message)` - 404 Not Found
- `InternalServerError(c, message)` - 500 Internal Server Error

---

### pkg-echo/request

Request binding and validation for Echo handlers.

```go
// Bind and validate in one call
var req LoginRequest
if !request.BindAndValidate(c, &req, map[string]string{
    "email": req.Email,
    "password": req.Password,
}) {
    return nil // error response already sent
}

// Validate email format
if !request.ValidateEmail(c, req.Email) {
    return nil // error response already sent
}

// Extract data from JWT token (set by middleware)
tokenData := middleware.GetTokenData(c)
userID := request.GetInt(tokenData, "user_id")
email := request.GetString(tokenData, "email")
isAdmin := request.GetBool(tokenData, "is_admin")
price := request.GetFloat(tokenData, "price")
```

**Available Functions:**
- `BindAndValidate(c, v, fields)` - Bind JSON and validate required fields
- `ValidateEmail(c, email)` - Validate email format and send error if invalid
- `GetInt(data, key)` - Safely extract int from map
- `GetString(data, key)` - Safely extract string from map
- `GetBool(data, key)` - Safely extract bool from map
- `GetFloat(data, key)` - Safely extract float64 from map

---

### pkg-echo/validator

Input validation utilities.

```go
// Validate email format
if !validator.IsValidEmail("user@example.com") {
    return errors.New("invalid email")
}

// Check if string is empty or whitespace
if validator.IsEmpty("   ") {
    return errors.New("field is required")
}

// Check minimum length
if !validator.MinLength("password", 8) {
    return errors.New("password too short")
}

// Validate multiple required fields
valid, errMsg := validator.ValidateRequired(map[string]string{
    "email": user.Email,
    "password": user.Password,
    "name": user.Name,
})
if !valid {
    return errors.New(errMsg)
}
```

**Available Functions:**
- `IsValidEmail(email)` - Check if email format is valid
- `IsEmpty(s)` - Check if string is empty or whitespace only
- `MinLength(s, min)` - Check if string meets minimum length
- `ValidateRequired(fields)` - Validate multiple required fields at once

---

### pkg-echo/orm

GORM database connection and transaction helpers.

```go
// Connect to PostgreSQL using GORM
db, err := orm.ConnectGORM("postgresql://user:pass@localhost:5432/dbname")

// Auto-migrate models
orm.AutoMigrate(db, &User{}, &Product{}, &Order{})

// Execute operations in transaction (auto-rollback on error)
err := orm.WithTransaction(db, func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    if err := tx.Create(&profile).Error; err != nil {
        return err
    }
    return nil
})
```

**Available Functions:**
- `ConnectGORM(dsn)` - Connect to PostgreSQL using GORM
- `AutoMigrate(db, models...)` - Run auto-migration for models
- `WithTransaction(db, fn)` - Execute function in database transaction

---

## Common Use Cases

### User Registration with Password Hashing

```go
func handleRegister(c echo.Context) error {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    if !request.BindAndValidate(c, &req, map[string]string{
        "email": req.Email, "password": req.Password,
    }) {
        return nil
    }
    
    if !request.ValidateEmail(c, req.Email) {
        return nil
    }
    
    hashedPassword, _ := auth.HashPassword(req.Password)
    user := User{Email: req.Email, Password: hashedPassword}
    
    if err := db.Create(&user).Error; err != nil {
        return response.BadRequest(c, "email already exists")
    }
    
    return response.Created(c, "user registered", map[string]interface{}{
        "user_id": user.ID,
        "email": user.Email,
    })
}
```

### User Login with JWT Token

```go
func handleLogin(c echo.Context) error {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    if !request.BindAndValidate(c, &req, map[string]string{
        "email": req.Email, "password": req.Password,
    }) {
        return nil
    }
    
    var user User
    if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
        return response.Unauthorized(c, "invalid credentials")
    }
    
    if !auth.ComparePassword(user.Password, req.Password) {
        return response.Unauthorized(c, "invalid credentials")
    }
    
    token, _ := auth.GenerateCustomToken(map[string]interface{}{
        "user_id": user.ID,
        "email": user.Email,
    }, "secret-key", 24*time.Hour)
    
    return response.Success(c, "login successful", map[string]string{
        "token": token,
    })
}
```

### Protected Route with JWT Authentication

```go
func setupRoutes(e *echo.Echo) {
    // Public routes
    e.POST("/login", handleLogin)
    
    // Protected routes
    api := e.Group("/api")
    api.Use(middleware.JWTMiddleware(middleware.JWTConfig{
        SecretKey: "your-secret-key",
        UseCustomToken: true,
    }))
    
    api.GET("/profile", handleProfile)
    api.GET("/orders", handleGetOrders)
}

func handleProfile(c echo.Context) error {
    tokenData := middleware.GetTokenData(c)
    userID := request.GetInt(tokenData, "user_id")
    
    var user User
    db.First(&user, userID)
    
    return response.Success(c, "profile retrieved", user)
}
```

### Database Transaction

```go
func createOrderWithItems(userID int, items []Item) error {
    return orm.WithTransaction(db, func(tx *gorm.DB) error {
        order := Order{UserID: userID}
        if err := tx.Create(&order).Error; err != nil {
            return err
        }
        
        for _, item := range items {
            orderItem := OrderItem{
                OrderID: order.ID,
                ProductID: item.ProductID,
                Quantity: item.Quantity,
            }
            if err := tx.Create(&orderItem).Error; err != nil {
                return err
            }
        }
        
        return nil
    })
}
```

---

## Examples

| Example | Description |
|---------|-------------|
| [01-basic-api](examples/01-basic-api) | Basic REST API with middleware |
| [02-database-connection](examples/02-database-connection) | Database connection patterns |
| [03-crud-api](examples/03-crud-api) | Full CRUD operations with PostgreSQL |
| [04-echo-jwt-api](examples/04-echo-jwt-api) | Echo framework integration guide |

Run examples:
```bash
cd examples/01-basic-api && go run main.go
cd examples/02-database-connection && go run main.go
cd examples/03-crud-api && go run main.go
```

---

## Dependencies

### For pkg/ (Standard Library)
```bash
go get github.com/lib/pq
```

### For pkg-echo/ (Echo Framework)
```bash
go get github.com/labstack/echo/v4
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto
go get gorm.io/gorm
go get gorm.io/driver/postgres
```

---

## Environment Variables

Create `.env` file in your project root:

```env
DATABASE_URL=postgresql://user:password@localhost:5432/dbname?sslmode=disable
PORT=8080
JWT_SECRET=your-secret-key

# Optional: Individual database config
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=mydb
DB_SSLMODE=disable

# Optional: Skip database connection
SKIP_DB=1
```

---

## Response Format

### Success Response (2xx)
```json
{
  "success": true,
  "message": "operation successful",
  "data": { }
}
```

### Error Response (4xx, 5xx)
```json
{
  "success": false,
  "error": "error message"
}
```

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

## License

MIT License - see LICENSE file for details

---

## Author

**Aisiya Qutwatunnada**  
GitHub: [@Yoochan45](https://github.com/Yoochan45)

---

## Changelog

### v0.2.0 (Latest)
- Added Echo framework support (pkg-echo/)
- Added JWT authentication with custom claims
- Added GORM ORM helpers
- Added request validation utilities
- Added transaction wrapper
- Fixed database connection on IPv6 networks
- Improved documentation

### v0.1.0
- Initial release
- PostgreSQL database helpers
- JSON response utilities
- CRUD query builders
- CORS and Logger middleware