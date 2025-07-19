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

// Condition returns a scope function that filters a query with a WHERE clause for the specified column and value
func Condition(column string, value interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(column, value)
	}
}

// OrderBy returns a scope function that sorts query results by the specified field in ascending or descending order
func OrderBy(field string, direction string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		switch strings.ToLower(direction) {
		case "asc":
			return db.Order(field)
		case "desc":
			return db.Order(field + " DESC")
		default:
			return db
		}
	}
}

// Limit returns a scope function that restricts the number of records returned by a query
func Limit(limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	}
}

// LikeCondition returns a scope function that filters a query with a LIKE clause using wildcard matching
func LikeCondition(column string, value string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(column+" LIKE ?", "%"+value+"%")
	}
}

// SelectFields returns a scope function that specifies the fields to include in query results
func SelectFields(fields ...string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Select(fields)
	}
}
