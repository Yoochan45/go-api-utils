package database

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/yoockh/go-api-utils/pkg/config"
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

	// Try to parse URL and prefer hostaddr query param (force IPv4) if provided.
	u, err := url.Parse(databaseURL)
	if err == nil && u.Scheme != "" {
		// extract components
		user := ""
		pass := ""
		if u.User != nil {
			user = u.User.Username()
			p, ok := u.User.Password()
			if ok {
				pass = p
			}
		}
		host := u.Hostname()
		port := u.Port()
		dbname := strings.TrimPrefix(u.Path, "/")
		q := u.Query()
		sslmode := q.Get("sslmode")
		hostaddr := q.Get("hostaddr")

		// if hostaddr provided, use it instead of hostname
		if hostaddr != "" {
			host = hostaddr
		}

		// fallback defaults
		if port == "" {
			port = "5432"
		}
		if sslmode == "" {
			sslmode = "require"
		}

		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, pass, dbname, sslmode,
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

		log.Println("PostgreSQL connection established successfully (via URL)")
		return db, nil
	}

	// Fallback to previous behavior if parse failed
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

// Init initializes DB based on provided config.
// If SKIP_DB=1 is set in env, Init will skip connecting and return (nil, nil).
// Prefer DATABASE_URL if present, otherwise use individual Postgres fields.
func Init(cfg *config.Config) (*sql.DB, error) {
	if os.Getenv("SKIP_DB") == "1" {
		log.Println("SKIP_DB=1 set, skipping DB connection")
		return nil, nil
	}

	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	// Prefer full DATABASE_URL
	if cfg.DatabaseURL != "" {
		// Try to parse and prefer hostaddr if present (see ConnectPostgresURL impl)
		return ConnectPostgresURL(cfg.DatabaseURL)
	}

	pg := PostgresConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}
	return ConnectPostgres(pg)
}
