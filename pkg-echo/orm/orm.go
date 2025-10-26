package orm

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectGORM connects to PostgreSQL using GORM
// Example:
//
//	db, err := orm.ConnectGORM("host=localhost port=5432 user=postgres password=secret dbname=mydb sslmode=disable")
func ConnectGORM(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("GORM connected to PostgreSQL")
	return db, nil
}

// AutoMigrate runs auto migration for given models
// Example:
//
//	orm.AutoMigrate(db, &User{}, &Vehicle{}, &Rental{})
func AutoMigrate(db *gorm.DB, models ...interface{}) error {
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}
	log.Println("auto migration completed")
	return nil
}

// Init initializes GORM connection from a database URL while respecting SKIP_DB
// If SKIP_DB=1, returns (nil, nil). If databaseURL is empty it will try SUPABASE_URL or DATABASE_URL env.
func Init(databaseURL string) (*gorm.DB, error) {
    // respect SKIP_DB
    if os.Getenv("SKIP_DB") == "1" {
        log.Println("SKIP_DB=1 set, skipping DB initialization")
        return nil, nil
    }

    // fallback envs
    if databaseURL == "" {
        if sup := os.Getenv("SUPABASE_URL"); sup != "" {
            databaseURL = sup
            log.Println("Init: using SUPABASE_URL")
        } else if d := os.Getenv("DATABASE_URL"); d != "" {
            databaseURL = d
            log.Println("Init: using DATABASE_URL from environment")
        }
    }

    if databaseURL == "" {
        return nil, fmt.Errorf("no database URL provided")
    }

    db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    log.Println("GORM connected to PostgreSQL (Init)")
    return db, nil
}
