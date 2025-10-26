package orm

import (
	"gorm.io/gorm"
)

// WithTransaction wraps function in database transaction
// Example:
//
//	err := orm.WithTransaction(db, func(tx *gorm.DB) error {
//	    if err := tx.Create(&user).Error; err != nil {
//	        return err
//	    }
//	    if err := tx.Create(&customer).Error; err != nil {
//	        return err
//	    }
//	    return nil
//	})
func WithTransaction(db *gorm.DB, fn func(*gorm.DB) error) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
