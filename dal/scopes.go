package dal

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	DefaultPageSize = 10

	MaxPageSize = 100
)

type ScalarValue interface {
	~bool | ~string | Number | time.Time
}

type RangeValue interface {
	Number | time.Time
}

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

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

func Equal[T ScalarValue](column string, value T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" = ?", value)
	}
}

func NotEqual[T ScalarValue](column string, value T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" != ?", value)
	}
}

func GreaterThan[T RangeValue](column string, value T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" > ?", value)
	}
}

func LessThan[T RangeValue](column string, value T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" < ?", value)
	}
}

func GreaterThanOrEqual[T RangeValue](column string, value T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" >= ?", value)
	}
}

func LessThanOrEqual[T RangeValue](column string, value T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" <= ?", value)
	}
}

func In[T ScalarValue](column string, values []T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(values) == 0 {
			return db.Where("1 = 0")
		}
		return db.Where(db.Statement.Quote(column)+" IN ?", values)
	}
}

func NotIn[T ScalarValue](column string, values []T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(values) == 0 {
			return db
		}
		return db.Where(db.Statement.Quote(column)+" NOT IN ?", values)
	}
}

func Between[T RangeValue](column string, lower, upper T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" BETWEEN ? AND ?", lower, upper)
	}
}

func NotBetween[T RangeValue](column string, lower, upper T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+" NOT BETWEEN ? AND ?", lower, upper)
	}
}

func IsNull(column string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column) + " IS NULL")
	}
}

func IsNotNull(column string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column) + " IS NOT NULL")
	}
}

func Order(column, direction string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		sortDirection := "ASC"
		if strings.EqualFold(direction, "desc") {
			sortDirection = "DESC"
		}
		return db.Order(db.Statement.Quote(column) + " " + sortDirection)
	}
}

func Limit(limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if limit <= 0 {
			return db
		}
		return db.Limit(limit)
	}
}

func Contains(column, value string) func(db *gorm.DB) *gorm.DB {
	escaped := escapeLike(value)
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+` LIKE ? ESCAPE '\'`, "%"+escaped+"%")
	}
}

func StartsWith(column, value string) func(db *gorm.DB) *gorm.DB {
	escaped := escapeLike(value)
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(db.Statement.Quote(column)+` LIKE ? ESCAPE '\'`, escaped+"%")
	}
}

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
