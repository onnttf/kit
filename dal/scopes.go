package dal

import (
	"strings"

	"gorm.io/gorm"
)

// Pagination constants
const (
	DefaultPageSize = 10  // Default number of items per page
	MaxPageSize     = 100 // Maximum allowed items per page
)

// Paginate returns a GORM scope function that applies pagination.
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

// Condition returns a GORM scope function that adds a 'where' condition.
func Condition(column string, value interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(column, value)
	}
}

// OrderBy returns a GORM scope function that orders the results by the specified field and direction.
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

// Limit returns a GORM scope function that limits the number of results.
func Limit(limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	}
}

// LikeCondition returns a GORM scope function that adds a 'where like' condition.
func LikeCondition(column string, value string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(column+" LIKE ?", "%"+value+"%")
	}
}

// SelectFields returns a GORM scope function that specifies the fields to select.
func SelectFields(fields ...string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Select(fields)
	}
}
