package dal

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Repository is a generic interface for CRUD operations on database entities of type T.
type Repository[T any] interface {
	// Insert adds a new entity to the database.
	Insert(ctx context.Context, db *gorm.DB, newValue *T) error

	// BatchInsert adds multiple entities in batches.
	BatchInsert(ctx context.Context, db *gorm.DB, newValues []*T, batchSize int) error

	// Update modifies an existing entity using the specified scopes.
	Update(ctx context.Context, db *gorm.DB, newValue *T, scopes ...func(db *gorm.DB) *gorm.DB) error

	// UpdateFields modifies specific fields of entities matching the provided scopes.
	UpdateFields(ctx context.Context, db *gorm.DB, newValue map[string]any, scopes ...func(db *gorm.DB) *gorm.DB) error

	// QueryOne retrieves a single entity matching the specified scopes.
	// If no rows are found, it returns an error.
	QueryOne(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (*T, error)

	// Query retrieves multiple entities matching the provided scopes.
	Query(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) ([]T, error)

	// Count returns the number of entities matching the provided scopes.
	Count(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (int64, error)

	// Raw executes a raw SQL query and scans results into type T.
	Raw(ctx context.Context, db *gorm.DB, sql string, args ...any) ([]T, error)

	// Delete removes entities matching the provided scopes.
	// It returns an error if no scopes are provided to prevent full table deletion.
	Delete(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) error
}

// Repo is a generic implementation of the Repository interface.
type Repo[T any] struct{}

// ErrDatabase is an unexpected database operation failure.
var ErrDatabase = errors.New("unexpected database error")

// ErrNoRowsAffected indicates a write operation did not affect any rows.
var ErrNoRowsAffected = errors.New("no rows affected")

// NewRepo creates a new repository instance for type T.
func NewRepo[T any]() *Repo[T] {
	return &Repo[T]{}
}

// Insert inserts a new entity into the database.
func (r *Repo[T]) Insert(ctx context.Context, db *gorm.DB, newValue *T) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if newValue == nil {
		return errors.New("new value is nil")
	}
	result := db.WithContext(ctx).Create(newValue)
	return handleExecError("insert", result)
}

// BatchInsert inserts multiple entities into the database in batches.
func (r *Repo[T]) BatchInsert(ctx context.Context, db *gorm.DB, newValues []*T, batchSize int) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if len(newValues) == 0 {
		return errors.New("new values is empty")
	}
	for i, newValue := range newValues {
		if newValue == nil {
			return fmt.Errorf("new values[%d] is nil", i)
		}
	}
	if batchSize <= 0 {
		batchSize = 10
	}
	result := db.WithContext(ctx).CreateInBatches(newValues, batchSize)
	return handleExecError("batch insert", result)
}

// Update updates an existing entity in the database.
// Only non-zero fields in newValue will be updated.
func (r *Repo[T]) Update(ctx context.Context, db *gorm.DB, newValue *T, scopes ...func(db *gorm.DB) *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if newValue == nil {
		return errors.New("new value is nil")
	}
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Updates(newValue)
	return handleExecError("update", result)
}

// UpdateFields updates specific fields of entities in the database.
func (r *Repo[T]) UpdateFields(ctx context.Context, db *gorm.DB, newValue map[string]any, scopes ...func(db *gorm.DB) *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if len(newValue) == 0 {
		return errors.New("new value is empty")
	}
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Updates(newValue)
	return handleExecError("update fields", result)
}

// QueryOne retrieves a single entity from the database matching the specified scopes.
// Returns gorm.ErrRecordNotFound if no matching record is found.
func (r *Repo[T]) QueryOne(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (*T, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	var record T
	result := db.WithContext(ctx).Scopes(scopes...).First(&record)
	return &record, handleQueryError("query one", result)
}

// Query retrieves multiple entities from the database matching the provided scopes.
// Returns an empty slice if no matching records are found.
func (r *Repo[T]) Query(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) ([]T, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	records := []T{}
	result := db.WithContext(ctx).Scopes(scopes...).Find(&records)
	return records, handleQueryError("query", result)
}

// Count returns the number of entities in the database matching the specified scopes.
// Returns 0 if no matching records are found.
func (r *Repo[T]) Count(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (int64, error) {
	if db == nil {
		return 0, errors.New("db is nil")
	}
	var count int64
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Count(&count)
	return count, handleQueryError("count", result)
}

// Delete removes entities from the database matching the provided scopes.
// Returns an error if no scopes are provided to prevent full table deletion.
func (r *Repo[T]) Delete(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if len(scopes) == 0 {
		return errors.New("delete without scope is not allowed")
	}
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Delete(new(T))
	return handleExecError("delete", result)
}

// Raw executes a raw SQL query and scans results into type T.
// Use this for complex queries like aggregations, joins, or custom SQL.
func (r *Repo[T]) Raw(ctx context.Context, db *gorm.DB, sql string, args ...any) ([]T, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	if sql == "" {
		return nil, errors.New("sql is empty")
	}
	results := []T{}
	result := db.WithContext(ctx).Raw(sql, args...).Find(&results)
	return results, handleQueryError("raw", result)
}

// Exec executes a raw SQL statement (INSERT/UPDATE/DELETE/DDL).
// Use this for batch updates, complex deletes, or direct SQL execution.
func Exec(ctx context.Context, db *gorm.DB, sql string, args ...any) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if sql == "" {
		return errors.New("sql is empty")
	}
	result := db.WithContext(ctx).Exec(sql, args...)
	return handleExecError("exec", result)
}

// handleExecError handles a GORM write operation.
func handleExecError(op string, result *gorm.DB) error {
	if result.Error != nil {
		return fmt.Errorf("%s failed: %w: %w", op, ErrDatabase, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, ErrNoRowsAffected)
	}
	return nil
}

// handleQueryError wraps a database error for query operations.
// Returns nil if result has no error.
func handleQueryError(op string, result *gorm.DB) error {
	if result.Error == nil {
		return nil
	}
	return fmt.Errorf("%s failed: %w: %w", op, ErrDatabase, result.Error)
}
