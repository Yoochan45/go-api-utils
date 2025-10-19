package main

import (
	"log"

	"github.com/Yoochan45/go-api-utils/pkg/config"
	"github.com/Yoochan45/go-api-utils/pkg/database"
)

func main() {
	// Load environment variables
	cfg := config.LoadEnv()

	// Option 1: Connect using DATABASE_URL (for Supabase, Railway, Render)
	if cfg.DatabaseURL != "" {
		log.Println("Connecting using DATABASE_URL...")
		db, err := database.ConnectPostgresURL(cfg.DatabaseURL)
		if err != nil {
			log.Fatal(err)
		}
		defer database.Close(db)

		// Test query
		var version string
		if err := db.QueryRow("SELECT version()").Scan(&version); err != nil {
			log.Fatal(err)
		}
		log.Printf("PostgreSQL version: %s", version)
	} else {
		// Option 2: Connect using individual config (for local development)
		log.Println("Connecting using individual config...")
		db, err := database.ConnectPostgres(database.PostgresConfig{
			Host:     cfg.DBHost,
			Port:     cfg.DBPort,
			User:     cfg.DBUser,
			Password: cfg.DBPassword,
			DBName:   cfg.DBName,
			SSLMode:  cfg.DBSSLMode,
		})
		if err != nil {
			log.Fatal(err)
		}
		defer database.Close(db)

		// Test query
		var version string
		if err := db.QueryRow("SELECT version()").Scan(&version); err != nil {
			log.Fatal(err)
		}
		log.Printf("PostgreSQL version: %s", version)
	}
}
