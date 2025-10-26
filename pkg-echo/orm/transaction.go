package orm

import "gorm.io/gorm"

// WithTransaction runs fn inside a database transaction using gorm.DB.Transaction.
// It commits on nil error, otherwise rolls back.
// Example:
//
//	err := orm.WithTransaction(db, func(tx *gorm.DB) error { ...; return nil })
func WithTransaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
