# Go API Utils

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

A lightweight utility library for building REST APIs in Go, supporting both standard net/http and Echo framework.

## Features

- Standard Library (pkg/)
  - PostgreSQL connection helpers (with pooling)
  - Standardized JSON responses
  - Request parsing and URL param helpers
  - CRUD SQL query builders
  - CORS and request logging middleware
  - Environment config loader
- Echo Framework (pkg-echo/)
  - JWT auth (basic and custom claims)
  - Password hashing (bcrypt) with configurable cost (BCRYPT_COST)
  - GORM helpers: connect, auto-migrate, transaction, pagination
  - Request binding and validation helpers
  - Response helpers (Success, SuccessData, Paginated, errors)
  - Middleware: JWT, role guard, user getters
  - Validator utilities
  - Health-check handler

## Installation

```bash
go get github.com/yoockh/go-api-utils
```

## Environment Variables

```env
DATABASE_URL=postgresql://user:password@localhost:5432/dbname?sslmode=disable
PORT=8080
JWT_SECRET=your-secret-key

# Optional: individual DB config (pkg/)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=mydb
DB_SSLMODE=disable

# Optional: skip DB in pkg/database.Init
SKIP_DB=1

# Optional: bcrypt cost override (default is bcrypt.DefaultCost)
BCRYPT_COST=12
```

## Quick Start

### Standard net/http API

```go
package main

import (
    "net/http"
    "github.com/yoockh/go-api-utils/pkg/database"
    "github.com/yoockh/go-api-utils/pkg/middleware"
    "github.com/yoockh/go-api-utils/pkg/response"
)

type Product struct {
    ID    int
    Name  string
    Price int
}

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
    "github.com/yoockh/go-api-utils/pkg/config"
    "github.com/yoockh/go-api-utils/pkg-echo/auth"
    echomw "github.com/yoockh/go-api-utils/pkg-echo/middleware"
    "github.com/yoockh/go-api-utils/pkg-echo/orm"
    "github.com/yoockh/go-api-utils/pkg-echo/request"
    "github.com/yoockh/go-api-utils/pkg-echo/response"
    "github.com/labstack/echo/v4"
)

type User struct{ ID int }

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
    api.Use(echomw.JWTMiddleware(echomw.JWTConfig{
        SecretKey:      "your-secret",
        UseCustomToken: true,
    }))
    api.GET("/profile", handleProfile)

    e.Start(":8080")
}

func handleLogin(c echo.Context) error {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    // Bind then validate required fields by JSON tag
    if !request.BindAndRequireFields(c, &req, "email", "password") {
        return nil // error response already sent
    }
    if !request.ValidateEmail(c, req.Email) {
        return nil // error response already sent
    }

    // Validate user (pseudo)
    // ...

    token, _ := auth.GenerateCustomToken(map[string]any{
        "user_id": 1,
        "email":   req.Email,
        "role":    "user",
    }, "your-secret", 24*time.Hour)

    return response.Success(c, "login success", map[string]string{"token": token})
}

func handleProfile(c echo.Context) error {
    data := echomw.GetTokenData(c)
    return response.Success(c, "profile", map[string]any{
        "user_id": request.GetInt(data, "user_id"),
        "email":   request.GetString(data, "email"),
        "role":    request.GetString(data, "role"),
    })
}
```

## Package Reference

### pkg/config
- LoadEnv() — load env with defaults
- MustLoadEnv() — load or panic

```go
cfg := config.LoadEnv()
_ = []string{cfg.DatabaseURL, cfg.Port}
```

### pkg/database
- ConnectPostgres(config), ConnectPostgresURL(url)
- MustConnect(config)
- Init(config) // respects SKIP_DB
- Close(db)

```go
db, err := database.ConnectPostgresURL(os.Getenv("DATABASE_URL"))
defer database.Close(db)
```

### pkg/response (net/http)

// How to use (examples)
```go
// 200 OK with wrapper
response.Success(w, "operation successful", data)

// 201 Created
response.Created(w, "resource created", newResource)

// 204 No Content
response.NoContent(w)

// Errors
response.BadRequest(w, "invalid input")              // 400
response.Unauthorized(w, "authentication required")  // 401
response.Forbidden(w, "access denied")               // 403
response.NotFound(w, "resource not found")           // 404
response.InternalServerError(w, "server error")      // 500
```

Function reference:
- Success
  - What it does: Send 200 OK with {success:true, message, data}
  - Signature: func Success(w http.ResponseWriter, message string, data interface{}) error
  - Example:
    ```go
    response.Success(w, "users retrieved", users)
    ```
- Created
  - What it does: Send 201 Created with wrapper
  - Signature: func Created(w http.ResponseWriter, message string, data interface{}) error
  - Example:
    ```go
    response.Created(w, "user created", user)
    ```
- NoContent
  - What it does: Send 204 No Content
  - Signature: func NoContent(w http.ResponseWriter) error
  - Example:
    ```go
    response.NoContent(w)
    ```
- BadRequest
  - What it does: Send 400 with error message
  - Signature: func BadRequest(w http.ResponseWriter, message string) error
  - Example:
    ```go
    response.BadRequest(w, "invalid payload")
    ```
- Unauthorized
  - What it does: Send 401 with error message
  - Signature: func Unauthorized(w http.ResponseWriter, message string) error
  - Example:
    ```go
    response.Unauthorized(w, "login required")
    ```
- Forbidden
  - What it does: Send 403 with error message
  - Signature: func Forbidden(w http.ResponseWriter, message string) error
  - Example:
    ```go
    response.Forbidden(w, "not allowed")
    ```
- NotFound
  - What it does: Send 404 with error message
  - Signature: func NotFound(w http.ResponseWriter, message string) error
  - Example:
    ```go
    response.NotFound(w, "user not found")
    ```
- InternalServerError
  - What it does: Send 500 with error message
  - Signature: func InternalServerError(w http.ResponseWriter, message string) error
  - Example:
    ```go
    response.InternalServerError(w, "unexpected error")
    ```

### pkg/request (net/http)
- ParseJSON
- GetIDFromURL
- GetQueryParam, GetQueryParamInt
- GetPathSegment

```go
var u User
_ = request.ParseJSON(r, &u)
id, _ := request.GetIDFromURL(r)
```

### pkg/repository
- BuildInsertQuery, BuildUpdateQuery, BuildSelectQuery
- CheckRowsAffected
- ScanRows

```go
q, args := repository.BuildInsertQuery("users", map[string]any{"name": "John"})
db.Exec(q, args...)
```

### pkg/middleware (net/http)
- CORS, Logger

```go
handler := middleware.Logger(middleware.CORS(mux))
```

---

### pkg-echo/auth
- HashPassword, ComparePassword (BCRYPT_COST supported)
- GenerateToken, ValidateToken
- GenerateCustomToken, ValidateCustomToken

```go
hashed, _ := auth.HashPassword("secret")
ok := auth.ComparePassword(hashed, "secret")
```

### pkg-echo/middleware
- JWTMiddleware(config)
- RequireRoles(roles...)
- GetTokenData(c)
- CurrentUserID(c), CurrentEmail(c), CurrentRole(c)

```go
api := e.Group("/api")
api.Use(echomw.JWTMiddleware(echomw.JWTConfig{
    SecretKey: "secret", UseCustomToken: true,
}))
api.Use(echomw.RequireRoles("admin"))

uid := echomw.CurrentUserID(c)
email := echomw.CurrentEmail(c)
role := echomw.CurrentRole(c)
_ = []any{uid, email, role}
```

### pkg-echo/request
- BindAndRequireFields(c, v, fields...)
- RequireFields(v, fields...) -> (ok, msg)
- ValidateEmail(c, email)
- QueryString, QueryInt, PathParamUint
- GetInt, GetUint, GetString, GetBool, GetFloat

```go
if !request.BindAndRequireFields(c, &req, "email", "password") { return nil }
if !request.ValidateEmail(c, req.Email) { return nil }

q := request.QueryString(c, "q", "")
page := request.QueryInt(c, "page", 1)
id := request.PathParamUint(c, "id")
_ = []any{q, page, id}
```

Note:
- BindAndValidate(c, v, map[string]string{...}) is deprecated. Prefer BindAndRequireFields.

### pkg-echo/response

// How to use (examples)
```go
// 200 OK with wrapper
response.Success(c, "ok", data)

// 200 OK raw data (no wrapper)
response.SuccessData(c, data)

// 201 Created
response.Created(c, "created", created)

// 204 No Content
response.NoContent(c)

// Errors
response.BadRequest(c, "invalid input")              // 400
response.Unauthorized(c, "authentication required")  // 401
response.Forbidden(c, "access denied")               // 403
response.NotFound(c, "not found")                    // 404
response.InternalServerError(c, "server error")      // 500

// 200 OK with pagination metadata
meta := map[string]any{"page": 1, "per_page": 10, "total": 42, "total_pages": 5}
response.Paginated(c, "list retrieved", items, meta)
```

Function reference:
- Success
  - What it does: 200 OK with {success:true, message, data}
  - Signature: func Success(c echo.Context, message string, data interface{}) error
  - Example:
    ```go
    return response.Success(c, "users retrieved", users)
    ```
- SuccessData
  - What it does: 200 OK with raw data (no wrapper)
  - Signature: func SuccessData(c echo.Context, data interface{}) error
  - Example:
    ```go
    return response.SuccessData(c, users)
    ```
- Created
  - What it does: 201 Created with wrapper
  - Signature: func Created(c echo.Context, message string, data interface{}) error
  - Example:
    ```go
    return response.Created(c, "user created", user)
    ```
- NoContent
  - What it does: 204 No Content
  - Signature: func NoContent(c echo.Context) error
  - Example:
    ```go
    return response.NoContent(c)
    ```
- BadRequest
  - What it does: 400 with error message
  - Signature: func BadRequest(c echo.Context, message string) error
  - Example:
    ```go
    return response.BadRequest(c, "invalid payload")
    ```
- Unauthorized
  - What it does: 401 with error message
  - Signature: func Unauthorized(c echo.Context, message string) error
  - Example:
    ```go
    return response.Unauthorized(c, "login required")
    ```
- Forbidden
  - What it does: 403 with error message
  - Signature: func Forbidden(c echo.Context, message string) error
  - Example:
    ```go
    return response.Forbidden(c, "not allowed")
    ```
- NotFound
  - What it does: 404 with error message
  - Signature: func NotFound(c echo.Context, message string) error
  - Example:
    ```go
    return response.NotFound(c, "user not found")
    ```
- InternalServerError
  - What it does: 500 with error message
  - Signature: func InternalServerError(c echo.Context, message string) error
  - Example:
    ```go
    return response.InternalServerError(c, "unexpected error")
    ```
- Paginated
  - What it does: 200 OK with {success:true, message, data, meta}
  - Signature: func Paginated(c echo.Context, message string, data interface{}, meta interface{}) error
  - Example:
    ```go
    meta := map[string]any{"page": page, "per_page": per, "total": total}
    return response.Paginated(c, "products", products, meta)
    ```

### pkg-echo/orm
- ConnectGORM(dsn)
- AutoMigrate(db, models...)
- WithTransaction(db, fn)
- ApplyPagination(db, page, perPage)
- CountAndPaginate(base, model, page, perPage, out) -> (total, err)

```go
var products []Product
base := db.Where("active = ?", true)
total, err := orm.CountAndPaginate(base, &Product{}, page, perPage, &products)
if err != nil { return response.InternalServerError(c, "failed to fetch") }

meta := map[string]any{
  "page": page, "per_page": perPage, "total": total,
  "total_pages": (total + int64(perPage) - 1) / int64(perPage),
}
return response.Paginated(c, "products", products, meta)
```

### pkg-echo/validator
- IsValidEmail, IsEmpty, MinLength
- ValidateRequired(map[string]string) -> (ok, msg)

```go
ok, msg := validator.ValidateRequired(map[string]string{"email": req.Email})
```

### pkg-echo/health
- NewHandler(db) — simple health endpoint

```go
e.GET("/health", health.NewHandler(db))
```

## Common Use Cases

### User Registration with Password Hashing

```go
func handleRegister(c echo.Context) error {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    if !request.BindAndRequireFields(c, &req, "email", "password") {
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

    return response.Created(c, "user registered", map[string]any{
        "user_id": user.ID,
        "email":   user.Email,
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

    if !request.BindAndRequireFields(c, &req, "email", "password") {
        return nil
    }
    if !request.ValidateEmail(c, req.Email) {
        return nil
    }

    var user User
    if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
        return response.Unauthorized(c, "invalid credentials")
    }

    if !auth.ComparePassword(user.Password, req.Password) {
        return response.Unauthorized(c, "invalid credentials")
    }

    token, _ := auth.GenerateCustomToken(map[string]any{
        "user_id": user.ID, "email": user.Email,
    }, "secret-key", 24*time.Hour)

    return response.Success(c, "login successful", map[string]string{
        "token": token,
    })
}
```

### Protected Route + Role Guard

```go
api := e.Group("/api/admin")
api.Use(echomw.JWTMiddleware(echomw.JWTConfig{SecretKey: "secret", UseCustomToken: true}))
api.Use(echomw.RequireRoles("admin"))
```

### Transaction

```go
err := orm.WithTransaction(db, func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil { return err }
    if err := tx.Create(&profile).Error; err != nil { return err }
    return nil
})
```

## Examples

- examples/01-basic-api — Basic REST API with middleware
- examples/02-database-connection — Database connection patterns
- examples/03-crud-api — Full CRUD operations with PostgreSQL
- examples/04-echo-jwt-api — Echo + JWT integration

Run:
```bash
cd examples/01-basic-api && go run main.go
cd examples/02-database-connection && go run main.go
cd examples/03-crud-api && go run main.go
```

## Dependencies

- pkg/: github.com/lib/pq
- pkg-echo/: github.com/labstack/echo/v4, github.com/golang-jwt/jwt/v5, golang.org/x/crypto, gorm.io/gorm, gorm.io/driver/postgres

## Response Format

Success:
```json
{ "success": true, "message": "operation successful", "data": {} }
```

Error:
```json
{ "success": false, "error": "error message" }
```

## Contributing

Contributions are welcome! Please open a PR.

## License

MIT

## Author

Aisiya Qutwatunnada — GitHub: @yoockh

## Changelog

### v0.2.6 (Latest)
- Echo request helpers: BindAndRequireFields, RequireFields (BindAndValidate deprecated)
- Request helpers: QueryString, QueryInt, PathParamUint; token data GetUint
- Response: SuccessData, Paginated
- JWT middleware: stronger Authorization parsing; expired token maps to 401 “token expired”; context getters CurrentUserID/Email/Role
- Middleware: RequireRoles for simple RBAC
- ORM: ApplyPagination, CountAndPaginate helpers; WithTransaction uses db.Transaction
- Health: health.NewHandler(db) endpoint
- Validator: ValidateRequired trims whitespace
- Auth: bcrypt cost override via BCRYPT_COST; token validation maps jwt.ErrTokenExpired -> ErrExpiredToken
- Docs: updated examples and usage

### v0.2.0
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
