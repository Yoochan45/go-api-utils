package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	Port        string
	DatabaseURL string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
}

// LoadEnv loads environment variables from .env file and returns Config
// Use this at app startup to load configuration
// Example:
//
//	config := LoadEnv()
//	db, _ := database.ConnectPostgresURL(config.DatabaseURL)
func LoadEnv() *Config {
	// Try to load .env file (ignore error if not exists)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", ""),
		DBName:      getEnv("DB_NAME", "mydb"),
		DBSSLMode:   getEnv("DB_SSL_MODE", "disable"),
	}
}

// getEnv retrieves environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// MustLoadEnv loads config and panics if DATABASE_URL is not set
// Use this when database is required for app to run
// Example:
//
//	config := MustLoadEnv()
func MustLoadEnv() *Config {
	config := LoadEnv()
	if config.DatabaseURL == "" && config.DBPassword == "" {
		log.Fatal("DATABASE_URL or database credentials must be set")
	}
	return config
}
