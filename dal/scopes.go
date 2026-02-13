package dal

import (
	"strings"

	"gorm.io/gorm"
)

// Pagination constants define the default and maximum number of records per page
const (
	DefaultPageSize = 10  // Default number of records per page
	MaxPageSize     = 100 // Maximum allowed records per page
)

// Paginate returns a scope function that applies pagination to a query using offset and limit, normalizing page and page size
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	// Normalize page number
	if page <= 0 {
		page = 1
	}

	// Normalize page size
	switch {
	case pageSize <= 0:
		pageSize = DefaultPageSize
	case pageSize > MaxPageSize:
		pageSize = MaxPageSize
	}

	offset := (page - 1) * pageSize

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(pageSize)
	}
}

// Condition returns a scope function that filters a query with a WHERE clause for the specified column and value.
// The column name is safely quoted to prevent SQL injection.
//
// Example:
//
//	db.Scopes(Condition("user_id", 123)).Find(&users)
func Condition(column string, value any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Use Quote to safely escape column name and prevent SQL injection
		return db.Where(db.Statement.Quote(column)+" = ?", value)
	}
}

// OrderBy returns a scope function that sorts query results by the specified field in ascending or descending order.
// The field name is safely quoted to prevent SQL injection. Direction must be "asc" or "desc" (case-insensitive).
//
// Example:
//
//	db.Scopes(OrderBy("created_at", "desc")).Find(&users)
func OrderBy(field string, direction string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Normalize direction to prevent injection
		dir := "ASC"
		if strings.ToLower(direction) == "desc" {
			dir = "DESC"
		}
		// Use Quote to safely escape field name
		return db.Order(db.Statement.Quote(field) + " " + dir)
	}
}

// Limit returns a scope function that restricts the number of records returned by a query
func Limit(limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	}
}

// LikeCondition returns a scope function that filters a query with a LIKE clause using wildcard matching.
// The column name is safely quoted to prevent SQL injection. Wildcards are automatically added around the value.
//
// Example:
//
//	db.Scopes(LikeCondition("name", "john")).Find(&users)  // WHERE name LIKE '%john%'
func LikeCondition(column string, value string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Use Quote to safely escape column name and prevent SQL injection
		return db.Where(db.Statement.Quote(column)+" LIKE ?", "%"+value+"%")
	}
}

// SelectFields returns a scope function that specifies the fields to include in query results
func SelectFields(fields ...string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Select(fields)
	}
}
