package orm

import "gorm.io/gorm"

// ApplyPagination applies LIMIT/OFFSET to a query based on page and perPage.
// Page starts from 1. Invalid values fall back to page=1, perPage=10.
// Example:
//
//	db = orm.ApplyPagination(db, page, perPage)
func ApplyPagination(db *gorm.DB, page, perPage int) *gorm.DB {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 || perPage > 1000 {
		perPage = 10
	}
	offset := (page - 1) * perPage
	return db.Limit(perPage).Offset(offset)
}

// CountAndPaginate counts rows for the given model and fetches the paginated records into out.
// "base" should contain filters/joins (but not limit/offset). "model" is used for COUNT.
// Example:
//
//	var books []Book
//	total, err := orm.CountAndPaginate(db.Where("author_id = ?", id), &Book{}, page, perPage, &books)
func CountAndPaginate(base *gorm.DB, model interface{}, page, perPage int, out interface{}) (int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 || perPage > 1000 {
		perPage = 10
	}
	var total int64
	if err := base.Session(&gorm.Session{}).Model(model).Count(&total).Error; err != nil {
		return 0, err
	}
	q := ApplyPagination(base, page, perPage)
	if err := q.Find(out).Error; err != nil {
		return 0, err
	}
	return total, nil
}
