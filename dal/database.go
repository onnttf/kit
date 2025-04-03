package dal

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Repository defines common CRUD operations.
type Repository[T any] interface {
	Insert(ctx context.Context, db *gorm.DB, newValue *T) error
	BatchInsert(ctx context.Context, db *gorm.DB, valuesToInsert []*T, batchSize int) error
	Update(ctx context.Context, db *gorm.DB, newValue *T, funcs ...func(db *gorm.DB) *gorm.DB) error
	UpdateWithMap(ctx context.Context, db *gorm.DB, newValue map[string]any, funcs ...func(db *gorm.DB) *gorm.DB) error
	Query(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) (*T, error)
	QueryList(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) ([]T, error)
	Count(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) (int64, error)
	Delete(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) error
}

type Repo[T any] struct{}

// NewRepo creates and returns a new repository instance.
func NewRepo[T any]() *Repo[T] {
	return &Repo[T]{}
}

// ErrDatabase indicates a general database-related error.
var ErrDatabase = fmt.Errorf("database error occurred")

// ErrNoChange is returned when no records are modified in an operation.
var ErrNoChange = fmt.Errorf("no records changed")

// handleError processes the result of a DB operation and returns appropriate errors.
func handleError(result *gorm.DB, action string) error {
	// If DB error exists, join it with ErrDatabase
	if result.Error != nil {
		return errors.Join(ErrDatabase, fmt.Errorf("%s, err: %w", action, result.Error))
	}
	// If no rows were affected, join it with ErrNoChange
	if result.RowsAffected == 0 {
		return errors.Join(ErrNoChange, fmt.Errorf("during %s", action))
	}
	return nil
}

// Insert adds a new record to the database.
func (r *Repo[T]) Insert(ctx context.Context, db *gorm.DB, newValue *T) error {
	if newValue == nil {
		return fmt.Errorf("invalid argument: newValue is nil")
	}
	result := db.WithContext(ctx).Create(newValue)
	return handleError(result, "insert")
}

// BatchInsert inserts multiple records in batches.
func (r *Repo[T]) BatchInsert(ctx context.Context, db *gorm.DB, valuesToInsert []*T, batchSize int) error {
	if len(valuesToInsert) == 0 {
		return fmt.Errorf("invalid argument: valuesToInsert is empty")
	}
	// Validate values in the batch
	for i, v := range valuesToInsert {
		if v == nil {
			return fmt.Errorf("invalid argument: value at index %d is nil", i)
		}
	}
	// Set default batch size if invalid
	if batchSize <= 0 {
		batchSize = 10
	}
	result := db.WithContext(ctx).CreateInBatches(valuesToInsert, batchSize)
	return handleError(result, "batch insert")
}

// Update modifies an existing record in the database.
func (r *Repo[T]) Update(ctx context.Context, db *gorm.DB, newValue *T, funcs ...func(db *gorm.DB) *gorm.DB) error {
	if newValue == nil {
		return fmt.Errorf("invalid argument: newValue is nil")
	}
	result := db.WithContext(ctx).Model(new(T)).Scopes(funcs...).Updates(newValue)
	return handleError(result, "update")
}

// UpdateWithMap modifies a record using a map of field updates.
func (r *Repo[T]) UpdateWithMap(ctx context.Context, db *gorm.DB, newValue map[string]any, funcs ...func(db *gorm.DB) *gorm.DB) error {
	if newValue == nil {
		return fmt.Errorf("invalid argument: newValue is nil")
	}
	result := db.WithContext(ctx).Model(new(T)).Scopes(funcs...).Updates(newValue)
	return handleError(result, "update with map")
}

// Delete removes a record from the database.
func (r *Repo[T]) Delete(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) error {
	result := db.WithContext(ctx).Model(new(T)).Scopes(funcs...).Delete(new(T))
	return handleError(result, "delete")
}

// Query retrieves a single record based on the provided conditions.
func (r *Repo[T]) Query(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) (*T, error) {
	var record T
	result := db.WithContext(ctx).Scopes(funcs...).Limit(1).Find(&record)
	if result.Error != nil {
		return nil, errors.Join(ErrDatabase, fmt.Errorf("failed to query one record, err: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &record, nil
}

// QueryList retrieves multiple records based on the provided conditions.
func (r *Repo[T]) QueryList(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) ([]T, error) {
	var recordList []T
	result := db.WithContext(ctx).Scopes(funcs...).Find(&recordList)
	if result.Error != nil {
		return nil, errors.Join(ErrDatabase, fmt.Errorf("failed to query list of records, err: %w", result.Error))
	}
	return recordList, nil
}

// Count returns the number of records that match the provided conditions.
func (r *Repo[T]) Count(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	result := db.WithContext(ctx).Model(new(T)).Scopes(funcs...).Count(&count)
	if result.Error != nil {
		return 0, errors.Join(ErrDatabase, fmt.Errorf("failed to count records, err: %w", result.Error))
	}
	return count, nil
}

const (
	defaultPageSize = 10  // Default number of items per page
	maxPageSize     = 100 // Maximum allowed items per page
)

// Paginate returns a GORM scope function that applies pagination based on the given page and pageSize.
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = defaultPageSize
	} else if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	offset := (page - 1) * pageSize

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(pageSize)
	}
}
