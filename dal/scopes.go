package dal

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	// DefaultPageSize is used when a caller passes a non-positive page size.
	DefaultPageSize = 10
	// MaxPageSize caps the page size used by Paginate.
	MaxPageSize = 100
)

// ScalarValue is a scalar SQL value accepted by equality and IN scopes.
type ScalarValue interface {
	~bool | ~string | Number | time.Time
}

// RangeValue is a scalar SQL value accepted by range comparison scopes.
type RangeValue interface {
	Number | time.Time
}

// Number is a scalar numeric SQL value.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// Paginate returns a scope that applies offset and limit for 1-based pages.
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	pageSize = min(pageSize, MaxPageSize)

	offset := (page - 1) * pageSize

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(pageSize)
	}
}

// Equal returns a scope for column = value.
func Equal[T ScalarValue](column string, value T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" = ?", value)
	}
}

// NotEqual returns a scope for column != value.
func NotEqual[T ScalarValue](column string, value T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" != ?", value)
	}
}

// GreaterThan returns a scope for column > value.
func GreaterThan[T RangeValue](column string, value T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" > ?", value)
	}
}

// LessThan returns a scope for column < value.
func LessThan[T RangeValue](column string, value T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" < ?", value)
	}
}

// GreaterThanOrEqual returns a scope for column >= value.
func GreaterThanOrEqual[T RangeValue](column string, value T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" >= ?", value)
	}
}

// LessThanOrEqual returns a scope for column <= value.
func LessThanOrEqual[T RangeValue](column string, value T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" <= ?", value)
	}
}

// In returns a scope for column IN values.
// When values is empty, it returns a condition that never matches.
func In[T ScalarValue](column string, values []T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(values) == 0 {
			return db.Where("1 = 0")
		}
		return db.Where(db.Statement.Quote(column)+" IN ?", values)
	}
}

// NotIn returns a scope for column NOT IN values.
// When values is empty, it applies no filter.
func NotIn[T ScalarValue](column string, values []T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(values) == 0 {
			return db
		}
		return db.Where(db.Statement.Quote(column)+" NOT IN ?", values)
	}
}

// Between returns a scope for column BETWEEN lower AND upper.
func Between[T RangeValue](column string, lower, upper T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" BETWEEN ? AND ?", lower, upper)
	}
}

// NotBetween returns a scope for column NOT BETWEEN lower AND upper.
func NotBetween[T RangeValue](column string, lower, upper T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" NOT BETWEEN ? AND ?", lower, upper)
	}
}

// IsNull returns a scope for column IS NULL.
func IsNull(column string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column) + " IS NULL")
	}
}

// IsNotNull returns a scope for column IS NOT NULL.
func IsNotNull(column string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column) + " IS NOT NULL")
	}
}

// Order returns a scope that orders by column. Only "desc" selects descending order.
func Order(column, direction string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		sortDirection := "ASC"
		if strings.EqualFold(direction, "desc") {
			sortDirection = "DESC"
		}
		return db.Order(db.Statement.Quote(column) + " " + sortDirection)
	}
}

// Limit returns a scope that applies a SQL limit.
// When limit is non-positive, it applies no limit.
func Limit(limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if limit <= 0 {
			return db
		}
		return db.Limit(limit)
	}
}

// Contains returns a scope that searches value as an escaped LIKE contains match.
func Contains(column, value string) func(db *gorm.DB) *gorm.DB {
	escaped := escapeLike(value)
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+` LIKE ? ESCAPE '\'`, "%"+escaped+"%")
	}
}

// StartsWith returns a scope that searches value as an escaped LIKE prefix match.
func StartsWith(column, value string) func(db *gorm.DB) *gorm.DB {
	escaped := escapeLike(value)
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+` LIKE ? ESCAPE '\'`, escaped+"%")
	}
}

// EndsWith returns a scope that searches value as an escaped LIKE suffix match.
func EndsWith(column, value string) func(db *gorm.DB) *gorm.DB {
	escaped := escapeLike(value)
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+` LIKE ? ESCAPE '\'`, "%"+escaped)
	}
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
