package dal

import (
	"strings"

	"gorm.io/gorm"
)

// escapeLike escapes LIKE wildcards (%, _, \) in s so they are matched literally.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// Pagination constants.
const (
	DefaultPageSize = 10  // Default number of records per page.
	MaxPageSize     = 100 // Maximum allowed records per page.
)

// Paginate returns a scope function that applies pagination to a query.
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

// Equal returns a scope function that filters a query by column = value.
func Equal(column string, value any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" = ?", value)
	}
}

// NotEqual returns a scope function that filters a query by column != value.
func NotEqual(column string, value any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" != ?", value)
	}
}

// GreaterThan returns a scope function that filters a query by column > value.
func GreaterThan(column string, value any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" > ?", value)
	}
}

// LessThan returns a scope function that filters a query by column < value.
func LessThan(column string, value any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" < ?", value)
	}
}

// GreaterThanOrEqual returns a scope function that filters a query by column >= value.
func GreaterThanOrEqual(column string, value any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" >= ?", value)
	}
}

// LessThanOrEqual returns a scope function that filters a query by column <= value.
func LessThanOrEqual(column string, value any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" <= ?", value)
	}
}

// In returns a scope function that filters a query by column IN (values).
func In(column string, values any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" IN ?", values)
	}
}

// NotIn returns a scope function that filters a query by column NOT IN (values).
func NotIn(column string, values any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" NOT IN ?", values)
	}
}

// Between returns a scope function that filters a query by column BETWEEN min AND max.
func Between(column string, min, max any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" BETWEEN ? AND ?", min, max)
	}
}

// NotBetween returns a scope function that filters a query by column NOT BETWEEN min AND max.
func NotBetween(column string, min, max any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" NOT BETWEEN ? AND ?", min, max)
	}
}

// IsNull returns a scope function that filters a query by column IS NULL.
func IsNull(column string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column) + " IS NULL")
	}
}

// IsNotNull returns a scope function that filters a query by column IS NOT NULL.
func IsNotNull(column string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column) + " IS NOT NULL")
	}
}

// OrderBy returns a scope function that sorts query results.
// The field name is quoted to prevent SQL injection.
func OrderBy(field, direction string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		sortDirection := "ASC"
		if strings.ToLower(direction) == "desc" {
			sortDirection = "DESC"
		}
		return db.Order(db.Statement.Quote(field) + " " + sortDirection)
	}
}

// Limit returns a scope function that limits the number of records returned.
func Limit(limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	}
}

// Like returns a scope function that filters a query using a LIKE clause.
// Wildcards (%, _, \) in value are escaped so they are matched literally.
func Like(column, value string) func(db *gorm.DB) *gorm.DB {
	escaped := escapeLike(value)
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+` LIKE ? ESCAPE '\'`, "%"+escaped+"%")
	}
}
