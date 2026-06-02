package dal

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Repository defines generic CRUD operations for a GORM-backed model.
type Repository[T any] interface {
	Insert(ctx context.Context, db *gorm.DB, newValue *T) error

	BatchInsert(ctx context.Context, db *gorm.DB, newValues []*T, batchSize int) error

	Update(ctx context.Context, db *gorm.DB, newValue *T, scopes ...func(db *gorm.DB) *gorm.DB) error

	UpdateFields(ctx context.Context, db *gorm.DB, newValue map[string]any, scopes ...func(db *gorm.DB) *gorm.DB) error

	QueryOne(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (*T, error)

	Query(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) ([]T, error)

	Count(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (int64, error)

	Raw(ctx context.Context, db *gorm.DB, sql string, args ...any) ([]T, error)

	Delete(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) error
}

type Repo[T any] struct{}

var (
	// ErrDatabase wraps unexpected GORM operation errors.
	ErrDatabase = errors.New("unexpected database error")
	// ErrNoRowsAffected indicates that a write operation matched no rows.
	ErrNoRowsAffected = errors.New("no rows affected")
)

func NewRepo[T any]() *Repo[T] {
	return &Repo[T]{}
}

func (r *Repo[T]) Insert(ctx context.Context, db *gorm.DB, newValue *T) error {
	if db == nil {
		return errors.New("insert: db is nil")
	}
	if newValue == nil {
		return errors.New("insert: new value is nil")
	}
	result := db.WithContext(ctx).Create(newValue)
	return handleExecError("insert", result)
}

// BatchInsert persists model values in batches. A non-positive batchSize uses a default.
func (r *Repo[T]) BatchInsert(ctx context.Context, db *gorm.DB, newValues []*T, batchSize int) error {
	if db == nil {
		return errors.New("batch insert: db is nil")
	}
	if len(newValues) == 0 {
		return errors.New("batch insert: new values are empty")
	}
	for i, newValue := range newValues {
		if newValue == nil {
			return fmt.Errorf("batch insert: new values[%d] is nil", i)
		}
	}
	if batchSize <= 0 {
		batchSize = 10
	}
	result := db.WithContext(ctx).CreateInBatches(newValues, batchSize)
	return handleExecError("batch insert", result)
}

// Update updates non-zero fields in newValue for rows matched by scopes.
func (r *Repo[T]) Update(ctx context.Context, db *gorm.DB, newValue *T, scopes ...func(db *gorm.DB) *gorm.DB) error {
	if db == nil {
		return errors.New("update: db is nil")
	}
	if newValue == nil {
		return errors.New("update: new value is nil")
	}
	if len(scopes) == 0 {
		return errors.New("update: scope is required")
	}
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Updates(newValue)
	return handleExecError("update", result)
}

// UpdateFields updates explicit fields for rows matched by scopes.
func (r *Repo[T]) UpdateFields(
	ctx context.Context,
	db *gorm.DB,
	newValue map[string]any,
	scopes ...func(db *gorm.DB) *gorm.DB,
) error {
	if db == nil {
		return errors.New("update fields: db is nil")
	}
	if len(newValue) == 0 {
		return errors.New("update fields: new value is empty")
	}
	if len(scopes) == 0 {
		return errors.New("update fields: scope is required")
	}
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Updates(newValue)
	return handleExecError("update fields", result)
}

func (r *Repo[T]) QueryOne(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (*T, error) {
	if db == nil {
		return nil, errors.New("query one: db is nil")
	}
	var record T
	result := db.WithContext(ctx).Scopes(scopes...).First(&record)
	if err := handleQueryError("query one", result); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repo[T]) Query(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) ([]T, error) {
	if db == nil {
		return nil, errors.New("query: db is nil")
	}
	records := []T{}
	result := db.WithContext(ctx).Scopes(scopes...).Find(&records)
	return records, handleQueryError("query", result)
}

func (r *Repo[T]) Count(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (int64, error) {
	if db == nil {
		return 0, errors.New("count: db is nil")
	}
	var count int64
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Count(&count)
	return count, handleQueryError("count", result)
}

// Delete removes rows matched by scopes. At least one scope is required.
func (r *Repo[T]) Delete(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) error {
	if db == nil {
		return errors.New("delete: db is nil")
	}
	if len(scopes) == 0 {
		return errors.New("delete: scope is required")
	}
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Delete(new(T))
	return handleExecError("delete", result)
}

// Raw executes a query and scans rows into []T.
func (r *Repo[T]) Raw(ctx context.Context, db *gorm.DB, sql string, args ...any) ([]T, error) {
	if db == nil {
		return nil, errors.New("raw: db is nil")
	}
	if sql == "" {
		return nil, errors.New("raw: sql is empty")
	}
	results := []T{}
	result := db.WithContext(ctx).Raw(sql, args...).Find(&results)
	return results, handleQueryError("raw", result)
}

// Exec executes a statement and allows zero affected rows.
func Exec(ctx context.Context, db *gorm.DB, sql string, args ...any) error {
	if db == nil {
		return errors.New("exec: db is nil")
	}
	if sql == "" {
		return errors.New("exec: sql is empty")
	}
	result := db.WithContext(ctx).Exec(sql, args...)
	if result.Error != nil {
		return fmt.Errorf("exec failed: %w: %w", ErrDatabase, result.Error)
	}
	return nil
}

func handleExecError(op string, result *gorm.DB) error {
	if result.Error != nil {
		return fmt.Errorf("%s failed: %w: %w", op, ErrDatabase, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, ErrNoRowsAffected)
	}
	return nil
}

func handleQueryError(op string, result *gorm.DB) error {
	if result.Error == nil {
		return nil
	}
	return fmt.Errorf("%s failed: %w: %w", op, ErrDatabase, result.Error)
}
