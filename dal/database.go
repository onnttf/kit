package dal

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Repository defines a generic interface for CRUD operations on database entities of type T.
type Repository[T any] interface {
	// Insert adds a new entity to the database.
	Insert(ctx context.Context, db *gorm.DB, newValue *T) error

	// BatchInsert adds multiple entities to the database in batches for efficiency.
	BatchInsert(ctx context.Context, db *gorm.DB, newValues []*T, batchSize int) error

	// Update modifies an existing entity using the specified scopes.
	Update(ctx context.Context, db *gorm.DB, newValue *T, scopes ...func(db *gorm.DB) *gorm.DB) error

	// UpdateFields modifies specific fields of entities matching the provided scopes.
	UpdateFields(ctx context.Context, db *gorm.DB, newValue map[string]any, scopes ...func(db *gorm.DB) *gorm.DB) error

	// QueryOne retrieves a single entity matching the specified scopes, returning an error if no rows are found.
	QueryOne(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (*T, error)

	// Query retrieves multiple entities matching the provided scopes.
	Query(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) ([]T, error)

	// Count returns the number of entities matching the provided scopes.
	Count(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (int64, error)

	// Raw executes a raw SQL query and scans results into type T.
	Raw(ctx context.Context, db *gorm.DB, sql string, args ...any) ([]T, error)

	// Delete removes entities matching the provided scopes from the database.
	// Returns an error if no scopes are provided to prevent full table deletion.
	Delete(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) error
}

// Repo provides a generic implementation of the Repository interface.
type Repo[T any] struct{}

// ErrDatabase indicates an unexpected database operation failure.
var ErrDatabase = errors.New("unexpected database error")

// ErrNoRowsAffected indicates a write operation did not affect any rows.
var ErrNoRowsAffected = errors.New("no rows affected")

// ErrNotFound indicates no records were found for a query.
var ErrNotFound = errors.New("record not found")

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
		return errors.New("newValue is nil")
	}
	result := db.WithContext(ctx).Create(newValue)
	return handleExecError(result, "insert")
}

// BatchInsert inserts multiple entities into the database in batches.
func (r *Repo[T]) BatchInsert(ctx context.Context, db *gorm.DB, newValues []*T, batchSize int) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if len(newValues) == 0 {
		return errors.New("newValues is empty")
	}
	for i, newValue := range newValues {
		if newValue == nil {
			return fmt.Errorf("newValues[%d] is nil", i)
		}
	}
	if batchSize <= 0 {
		batchSize = 10
	}
	result := db.WithContext(ctx).CreateInBatches(newValues, batchSize)
	return handleExecError(result, "batch insert")
}

// Update updates an existing entity in the database.
// Only non-zero fields in newValue will be updated.
func (r *Repo[T]) Update(ctx context.Context, db *gorm.DB, newValue *T, scopes ...func(db *gorm.DB) *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if newValue == nil {
		return errors.New("newValue is nil")
	}
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Updates(newValue)
	return handleExecError(result, "update")
}

// UpdateFields updates specific fields of entities in the database.
func (r *Repo[T]) UpdateFields(ctx context.Context, db *gorm.DB, newValue map[string]any, scopes ...func(db *gorm.DB) *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if len(newValue) == 0 {
		return errors.New("newValue is empty")
	}
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Updates(newValue)
	return handleExecError(result, "update fields")
}

// QueryOne retrieves a single entity from the database matching the specified scopes.
// Returns ErrNotFound if no matching record is found, making it easy to distinguish between "not found" and other errors.
//
// Example:
//
//	user, err := repo.QueryOne(ctx, db, Equal("id", 123))
//	if errors.Is(err, dal.ErrNotFound) {
//	    // handle not found case
//	}
func (r *Repo[T]) QueryOne(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (*T, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	var record T
	result := db.WithContext(ctx).Scopes(scopes...).First(&record)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, handleQueryError("query one", result.Error)
	}
	return &record, nil
}

// Query retrieves multiple entities from the database matching the specified scopes.
// Returns an empty slice (not an error) if no matching records are found.
func (r *Repo[T]) Query(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) ([]T, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	records := []T{}
	result := db.WithContext(ctx).Scopes(scopes...).Find(&records)
	if result.Error != nil {
		return nil, handleQueryError("query", result.Error)
	}
	return records, nil
}

// Count returns the number of entities in the database matching the specified scopes.
// Returns 0 (not an error) if no matching records are found.
func (r *Repo[T]) Count(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (int64, error) {
	if db == nil {
		return 0, errors.New("db is nil")
	}
	var count int64
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Count(&count)
	if result.Error != nil {
		return 0, handleQueryError("count", result.Error)
	}
	return count, nil
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
	return handleExecError(result, "delete")
}

// Raw executes a raw SQL query and scans results into type T.
// Use this for complex queries like aggregations, joins, or custom SQL.
//
// Example:
//
//	type StatResult struct {
//	    Total int64  `json:"total"`
//	    Name  string `json:"name"`
//	}
//	repo := dal.NewRepo[StatResult]()
//	results, err := repo.Raw(ctx, db, "SELECT count(*) as total, name FROM users GROUP BY name")
func (r *Repo[T]) Raw(ctx context.Context, db *gorm.DB, sql string, args ...any) ([]T, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	if sql == "" {
		return nil, errors.New("sql is empty")
	}
	var results []T
	result := db.WithContext(ctx).Raw(sql, args...).Find(&results)
	if result.Error != nil {
		return nil, handleQueryError("raw", result.Error)
	}
	return results, nil
}

// Exec executes a raw SQL statement (INSERT/UPDATE/DELETE/DDL).
// Use this for batch updates, complex deletes, or direct SQL execution.
//
// Example:
//
//	err := dal.Exec(ctx, db, "UPDATE users SET status = ? WHERE status = ?", 0, 1)
func Exec(ctx context.Context, db *gorm.DB, sql string, args ...any) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if sql == "" {
		return errors.New("sql is empty")
	}
	result := db.WithContext(ctx).Exec(sql, args...)
	return handleExecError(result, "exec")
}

// handleExecError handles a GORM write operation.
func handleExecError(result *gorm.DB, op string) error {
	if result.Error != nil {
		dbErr := fmt.Errorf("%w: %w", ErrDatabase, result.Error)
		return fmt.Errorf("dal %s failed: %w", op, dbErr)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, ErrNoRowsAffected)
	}
	return nil
}

// handleQueryError wraps a database error for query operations.
func handleQueryError(op string, err error) error {
	dbErr := fmt.Errorf("%w: %w", ErrDatabase, err)
	return fmt.Errorf("dal %s failed: %w", op, dbErr)
}
