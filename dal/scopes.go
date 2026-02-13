package dal

import (
	"strings"

	"gorm.io/gorm"
)

// Pagination constants.
const (
	DefaultPageSize = 10  // Default number of records per page.
	MaxPageSize     = 100 // Maximum allowed records per page.
)

// Paginate returns a scope function that applies pagination to a query.
//
// Example:
//
//	db.Scopes(Paginate(1, 20)).Find(&users) // page 1, 20 items per page
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = DefaultPageSize
	} else {
		pageSize = min(pageSize, MaxPageSize)
	}

	offset := (page - 1) * pageSize

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(pageSize)
	}
}

// Condition returns a scope function that filters a query by column and value.
// The column name is quoted to prevent SQL injection.
//
// Example:
//
//	db.Scopes(Condition("user_id", 123)).Find(&users)
func Condition(column string, value any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" = ?", value)
	}
}

// OrderBy returns a scope function that sorts query results.
// The direction must be "asc" or "desc" (case-insensitive).
// The field name is quoted to prevent SQL injection.
//
// Example:
//
//	db.Scopes(OrderBy("created_at", "desc")).Find(&users)
func OrderBy(field string, direction string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		dir := "ASC"
		if strings.ToLower(direction) == "desc" {
			dir = "DESC"
		}
		return db.Order(db.Statement.Quote(field) + " " + dir)
	}
}

// Limit returns a scope function that limits the number of records returned.
//
// Example:
//
//	db.Scopes(Limit(10)).Find(&users)
func Limit(limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	}
}

// LikeCondition returns a scope function that filters a query using a LIKE clause.
// Wildcards are automatically added around the value.
// The column name is quoted to prevent SQL injection.
//
// Example:
//
//	db.Scopes(LikeCondition("name", "john")).Find(&users) // WHERE name LIKE '%john%'
func LikeCondition(column string, value string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" LIKE ?", "%"+value+"%")
	}
}

// SelectFields returns a scope function that specifies the fields to include.
//
// Example:
//
//	db.Scopes(SelectFields("id", "name", "email")).Find(&users) // SELECT id, name, email
func SelectFields(fields ...string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Select(fields)
	}
}
