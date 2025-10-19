package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// PostgresConfig holds database connection configuration
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ConnectPostgres establishes connection to PostgreSQL using individual parameters
// Use this for local development or when you have separate config values
// Example:
//
//	config := PostgresConfig{Host: "localhost", Port: "5432", User: "postgres", Password: "secret", DBName: "mydb", SSLMode: "disable"}
//	db, err := ConnectPostgres(config)
func ConnectPostgres(config PostgresConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("PostgreSQL connection established successfully")
	return db, nil
}

// ConnectPostgresURL establishes connection using DATABASE_URL format
// Use this for cloud databases (Supabase, Railway, Render, Heroku)
// Example:
//
//	db, err := ConnectPostgresURL("postgresql://user:password@host:5432/dbname")
func ConnectPostgresURL(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL cannot be empty")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("PostgreSQL connection established successfully (via URL)")
	return db, nil
}

// MustConnect is a helper that panics if connection fails
// Use this when you want the app to crash if database is not available
// Example:
//
//	db := MustConnect("postgresql://user:password@host:5432/dbname")
func MustConnect(databaseURL string) *sql.DB {
	db, err := ConnectPostgresURL(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}

// Close gracefully closes the database connection
// Always defer this after opening connection
// Example:
//
//	db, _ := ConnectPostgresURL(url)
//	defer Close(db)
func Close(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database connection closed")
		}
	}
}
