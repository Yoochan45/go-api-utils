package orm

import (
	"fmt"
	"log"

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
