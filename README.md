# Go API Utils

Fast Go API development helper

## Features

### Standard (net/http)
- Quick database connection (PostgreSQL)  
- JSON response helpers  
- Request parsing utilities  
- CRUD query builders  
- CORS & Logger middleware  
- Environment config loader  

### Echo Framework
- JWT authentication middleware
- Password hashing (bcrypt)
- GORM ORM helpers
- Transaction wrapper
- Echo response helpers
- Request validator

## Installation

```bash
go get github.com/Yoochan45/go-api-utils
```

## Quick Start

### 1. Basic API (No Database)

```go
package main

import (
    "github.com/Yoochan45/go-api-utils/pkg/middleware"
    "github.com/Yoochan45/go-api-utils/pkg/response"
    "net/http"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        response.Success(w, "Hello World", nil)
    })
    
    handler := middleware.Logger(middleware.CORS(mux))
    http.ListenAndServe(":8080", handler)
}
```

### 2. Connect to Database

```go
// Using DATABASE_URL (Supabase, Railway, Render)
db, _ := database.ConnectPostgresURL(os.Getenv("DATABASE_URL"))
defer database.Close(db)

// Using individual config (Local)
db, _ := database.ConnectPostgres(database.PostgresConfig{
    Host: "localhost",
    Port: "5432",
    User: "postgres",
    Password: "secret",
    DBName: "mydb",
    SSLMode: "disable",
})
```

### 3. Full CRUD Example

See `examples/03-crud-api/main.go`

### 4. Echo + JWT + GORM Example

```go
package main

import (
    "github.com/Yoochan45/go-api-utils/pkg/config"
    "github.com/Yoochan45/go-api-utils/pkg-echo/auth"
    echomiddleware "github.com/Yoochan45/go-api-utils/pkg-echo/middleware"
    "github.com/Yoochan45/go-api-utils/pkg-echo/orm"
    "github.com/Yoochan45/go-api-utils/pkg-echo/response"
    "github.com/Yoochan45/go-api-utils/pkg-echo/request"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "gorm.io/gorm"
)

type User struct {
    gorm.Model
    Email    string `gorm:"unique"`
    Password string
}

var (
    db        *gorm.DB
    jwtSecret = "your-secret-key"
)

func main() {
    cfg := config.LoadEnv()
    
    // Connect GORM
    var err error
    db, err = orm.ConnectGORM(cfg.DatabaseURL)
    if err != nil {
        panic(err)
    }
    orm.AutoMigrate(db, &User{})
    
    e := echo.New()
    
    // Global middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())
    
    // Public routes
    e.POST("/register", handleRegister)
    e.POST("/login", handleLogin)
    
    // Protected routes
    api := e.Group("/api")
    api.Use(echomiddleware.JWTMiddleware(echomiddleware.JWTConfig{
        SecretKey: jwtSecret,
    }))
    api.GET("/profile", handleProfile)
    
    e.Start(":" + cfg.Port)
}

func handleRegister(c echo.Context) error {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    // Bind and validate
    if !request.BindAndValidate(c, &req, map[string]string{
        "email":    req.Email,
        "password": req.Password,
    }) {
        return nil
    }
    
    // Validate email format
    if !request.ValidateEmail(c, req.Email) {
        return nil
    }
    
    // Hash password
    hashed, _ := auth.HashPassword(req.Password)
    user := User{Email: req.Email, Password: hashed}
    
    // Save to database
    if err := db.Create(&user).Error; err != nil {
        return response.BadRequest(c, "email already exists")
    }
    
    return response.Created(c, "user created", map[string]interface{}{
        "user_id": user.ID,
        "email":   user.Email,
    })
}

func handleLogin(c echo.Context) error {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    if !request.BindAndValidate(c, &req, map[string]string{
        "email":    req.Email,
        "password": req.Password,
    }) {
        return nil
    }
    
    // Get user from database
    var user User
    if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
        return response.Unauthorized(c, "invalid credentials")
    }
    
    // Compare password
    if !auth.ComparePassword(user.Password, req.Password) {
        return response.Unauthorized(c, "invalid credentials")
    }
    
    // Generate JWT token (24 hours)
    token, _ := auth.GenerateToken(int(user.ID), user.Email, "user", jwtSecret, 24*3600*1000000000)
    
    return response.Success(c, "login success", map[string]interface{}{
        "token": token,
    })
}

func handleProfile(c echo.Context) error {
    // Get user from JWT claims (set by middleware)
    userID := echomiddleware.GetUserID(c)
    email := echomiddleware.GetUserEmail(c)
    
    return response.Success(c, "profile retrieved", map[string]interface{}{
        "user_id": userID,
        "email":   email,
    })
}
```

## Package Documentation

### Standard (net/http) Packages

#### `pkg/database`
- `ConnectPostgres()` - Connect using config struct
- `ConnectPostgresURL()` - Connect using DATABASE_URL
- `MustConnect()` - Connect or panic
- `Init()` - Initialize database (auto-detect SKIP_DB)
- `Close()` - Close connection

#### `pkg/response`
- `Success()` - 200 OK response
- `Created()` - 201 Created response
- `NoContent()` - 204 No Content
- `BadRequest()` - 400 Bad Request
- `NotFound()` - 404 Not Found
- `Unauthorized()` - 401 Unauthorized
- `Forbidden()` - 403 Forbidden
- `InternalServerError()` - 500 Error

#### `pkg/request`
- `ParseJSON()` - Parse JSON body
- `GetIDFromURL()` - Extract ID from URL
- `GetQueryParam()` - Get query parameter
- `GetQueryParamInt()` - Get integer query param
- `GetPathSegment()` - Extract path segment

#### `pkg/repository`
- `BuildInsertQuery()` - Generate INSERT query
- `BuildUpdateQuery()` - Generate UPDATE query
- `BuildSelectQuery()` - Generate SELECT query
- `CheckRowsAffected()` - Check if rows were affected
- `ScanRows()` - Scan multiple rows into slice

#### `pkg/middleware`
- `CORS()` - Add CORS headers
- `Logger()` - Log HTTP requests

#### `pkg/config`
- `LoadEnv()` - Load environment variables
- `MustLoadEnv()` - Load or panic

### Echo Framework Packages

#### `pkg-echo/auth`
- `HashPassword(password string)` - Hash password using bcrypt
- `ComparePassword(hashed, plain string)` - Compare plain password with hash
- `GenerateToken(userID, email, role, secret, expiry)` - Generate JWT token
- `ValidateToken(token, secret)` - Validate JWT token and return claims
- `ParseClaims(token)` - Parse claims without validation

#### `pkg-echo/middleware`
- `JWTMiddleware(config JWTConfig)` - JWT authentication middleware for Echo
- `GetUserID(c echo.Context)` - Get user ID from context
- `GetUserEmail(c echo.Context)` - Get email from context
- `GetUserRole(c echo.Context)` - Get role from context
- `GetClaims(c echo.Context)` - Get full JWT claims

#### `pkg-echo/response`
- `Success(c, message, data)` - 200 OK response
- `SuccessData(c, data)` - 200 OK with plain data
- `Created(c, message, data)` - 201 Created
- `NoContent(c)` - 204 No Content
- `BadRequest(c, message)` - 400 Bad Request
- `Unauthorized(c, message)` - 401 Unauthorized
- `Forbidden(c, message)` - 403 Forbidden
- `NotFound(c, message)` - 404 Not Found
- `InternalServerError(c, message)` - 500 Error

#### `pkg-echo/validator`
- `IsValidEmail(email)` - Validate email format
- `IsEmpty(s)` - Check if string is empty
- `MinLength(s, min)` - Check minimum length
- `ValidateRequired(fields)` - Validate required fields

#### `pkg-echo/orm`
- `ConnectGORM(dsn)` - Connect to PostgreSQL using GORM
- `AutoMigrate(db, models...)` - Run auto migration
- `WithTransaction(db, fn)` - Execute function in transaction

#### `pkg-echo/request`
- `BindAndValidate(c, v, fields)` - Bind and validate request
- `ValidateEmail(c, email)` - Validate email and send error

## Examples

Run examples:
```bash
# Standard net/http examples
cd examples/01-basic-api && go run main.go
cd examples/02-database-connection && go run main.go
cd examples/03-crud-api && go run main.go

# Echo + JWT example
cd examples/04-echo-jwt-api && go run main.go
```

## Common Use Cases

### Register User (Echo + GORM)
```go
func handleRegister(c echo.Context) error {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    if !request.BindAndValidate(c, &req, map[string]string{
        "email": req.Email,
        "password": req.Password,
    }) {
        return nil
    }
    
    hashed, _ := auth.HashPassword(req.Password)
    user := User{Email: req.Email, Password: hashed}
    
    if err := db.Create(&user).Error; err != nil {
        return response.BadRequest(c, "email already exists")
    }
    
    return response.Created(c, "user created", user)
}
```

### Login User (Echo + JWT)
```go
func handleLogin(c echo.Context) error {
    var req LoginRequest
    if !request.BindAndValidate(c, &req, map[string]string{
        "email": req.Email,
        "password": req.Password,
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
    
    token, _ := auth.GenerateToken(int(user.ID), user.Email, "user", jwtSecret, 24*time.Hour)
    
    return response.Success(c, "login success", map[string]string{"token": token})
}
```

### Protected Route (Echo + JWT)
```go
func main() {
    e := echo.New()
    
    // Public routes
    e.POST("/login", handleLogin)
    
    // Protected routes
    api := e.Group("/api")
    api.Use(echomiddleware.JWTMiddleware(echomiddleware.JWTConfig{
        SecretKey: "your-secret",
    }))
    api.GET("/profile", handleProfile)
    
    e.Start(":8080")
}

func handleProfile(c echo.Context) error {
    userID := echomiddleware.GetUserID(c)
    email := echomiddleware.GetUserEmail(c)
    
    return response.Success(c, "profile", map[string]interface{}{
        "user_id": userID,
        "email": email,
    })
}
```

### Database Transaction (GORM)
```go
import "github.com/Yoochan45/go-api-utils/pkg-echo/orm"

err := orm.WithTransaction(db, func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    if err := tx.Create(&customer).Error; err != nil {
        return err
    }
    return nil
})
```

## License

MIT

## Developer & Contributors

**Name:** Aisiya Qutwatunnada  
**GitHub:** @Yoochan45