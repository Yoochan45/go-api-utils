package repository

import (
	"database/sql"
	"fmt"
	"strings"
)

// BuildInsertQuery generates INSERT SQL query dynamically
// Use this to avoid writing repetitive INSERT queries
// Example:
//
//	query := BuildInsertQuery("products", []string{"name", "price", "stock"})
//	// Returns: INSERT INTO products (name, price, stock) VALUES ($1, $2, $3) RETURNING id
func BuildInsertQuery(table string, columns []string) string {
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
}

// BuildUpdateQuery generates UPDATE SQL query dynamically
// Use this to avoid writing repetitive UPDATE queries
// Example:
//
//	query := BuildUpdateQuery("products", []string{"name", "price", "stock"})
//	// Returns: UPDATE products SET name = $1, price = $2, stock = $3 WHERE id = $4
func BuildUpdateQuery(table string, columns []string) string {
	setClauses := make([]string, len(columns))
	for i, col := range columns {
		setClauses[i] = fmt.Sprintf("%s = $%d", col, i+1)
	}

	return fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d",
		table,
		strings.Join(setClauses, ", "),
		len(columns)+1,
	)
}

// BuildSelectQuery generates SELECT SQL query with optional WHERE clause
// Use this to build dynamic SELECT queries
// Example:
//
//	query := BuildSelectQuery("products", []string{"id", "name", "price"}, "")
//	// Returns: SELECT id, name, price FROM products
func BuildSelectQuery(table string, columns []string, whereClause string) string {
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), table)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	return query
}

// CheckRowsAffected checks if any rows were affected by query
// Use this after UPDATE/DELETE to check if resource exists
// Example:
//
//	result, _ := db.Exec("UPDATE products SET name = $1 WHERE id = $2", name, id)
//	if err := CheckRowsAffected(result); err != nil {
//	    return errors.New("product not found")
//	}
func CheckRowsAffected(result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ScanRows scans multiple rows into slice
// Generic helper for scanning query results
// Note: You still need to provide custom scan function for your struct
func ScanRows[T any](rows *sql.Rows, scanFunc func(*sql.Rows) (T, error)) ([]T, error) {
	var results []T
	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}
