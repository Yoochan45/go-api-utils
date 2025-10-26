package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/Yoochan45/go-api-utils/pkg-echo/auth"
	echomiddleware "github.com/Yoochan45/go-api-utils/pkg-echo/middleware"
	"github.com/Yoochan45/go-api-utils/pkg/config"
	"github.com/Yoochan45/go-api-utils/pkg/database"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	db        *sql.DB
	jwtSecret = "your-secret-key-change-in-production"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func main() {
	cfg := config.LoadEnv()

	// init DB
	var err error
	db, err = database.Init(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if db != nil {
		defer database.Close(db)
	}

	// get JWT secret from env or use default
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		jwtSecret = secret
	}

	e := echo.New()

	// global middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// public routes
	e.POST("/register", handleRegister)
	e.POST("/login", handleLogin)

	// protected routes (require JWT)
	api := e.Group("/api")
	api.Use(echomiddleware.JWTMiddleware(echomiddleware.JWTConfig{
		SecretKey: jwtSecret,
	}))
	api.GET("/profile", handleProfile)
	api.GET("/products", handleProducts)

	e.Logger.Fatal(e.Start(":" + cfg.Port))
}

func handleRegister(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
	}

	// save to DB (example query)
	var userID int
	err = db.QueryRow(
		"INSERT INTO users (email, password, name) VALUES ($1, $2, $3) RETURNING id",
		req.Email, hashedPassword, req.Name,
	).Scan(&userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "user created",
		"user_id": userID,
	})
}

func handleLogin(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// get user from DB
	var userID int
	var hashedPassword string
	err := db.QueryRow("SELECT id, password FROM users WHERE email = $1", req.Email).
		Scan(&userID, &hashedPassword)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "database error"})
	}

	// compare password
	if !auth.ComparePassword(hashedPassword, req.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	// generate JWT token (24 hour expiry)
	token, err := auth.GenerateToken(userID, req.Email, "user", jwtSecret, 24*3600*1000000000)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "login success",
		"token":   token,
	})
}

func handleProfile(c echo.Context) error {
	// get user from JWT claims (set by middleware)
	userID := echomiddleware.GetUserID(c)
	email := echomiddleware.GetUserEmail(c)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"email":   email,
	})
}

func handleProducts(c echo.Context) error {
	// example protected route
	return c.JSON(http.StatusOK, map[string]string{
		"message": "this is protected route",
	})
}
