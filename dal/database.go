package dal

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Repository defines a set of generic CRUD operations for database entities of type T.
type Repository[T any] interface {
	// Insert inserts a new record into the database.
	Insert(ctx context.Context, db *gorm.DB, newValue *T) error

	// BatchInsert inserts multiple new records into the database in batches.
	BatchInsert(ctx context.Context, db *gorm.DB, newValues []*T, batchSize int) error

	// Update updates an existing record in the database using the provided scopes.
	Update(ctx context.Context, db *gorm.DB, newValue *T, scopes ...func(db *gorm.DB) *gorm.DB) error

	// UpdateFields updates specific fields of records matching the given scopes.
	UpdateFields(ctx context.Context, db *gorm.DB, newValue map[string]any, scopes ...func(db *gorm.DB) *gorm.DB) error

	// FindOne retrieves one record matching the provided scopes.
	FindOne(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (*T, error)

	// Find retrieves multiple records matching the provided scopes.
	Find(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) ([]T, error)

	// Count returns the number of records matching the provided scopes.
	Count(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (int64, error)

	// Delete removes records from the database matching the provided scopes.
	Delete(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) error
}

// Repo is a generic implementation of Repository.
type Repo[T any] struct{}

// NewRepo creates and returns a new repository instance.
func NewRepo[T any]() *Repo[T] {
	return &Repo[T]{}
}

// Common database errors
var (
	ErrDatabase = errors.New("database error occurred")
	ErrNoChange = errors.New("no records changed")
)

// handleError processes the result of a DB operation and returns appropriate errors.
func handleError(result *gorm.DB, action string) error {
	if result.Error != nil {
		return errors.Join(ErrDatabase, fmt.Errorf("%s failed: %w", action, result.Error))
	}
	if result.RowsAffected == 0 {
		return errors.Join(ErrNoChange, fmt.Errorf("no rows affected during %s", action))
	}
	return nil
}

// Insert adds a new record to the database.
func (r *Repo[T]) Insert(ctx context.Context, db *gorm.DB, newValue *T) error {
	if newValue == nil {
		return fmt.Errorf("insert failed: newValue is nil")
	}
	result := db.WithContext(ctx).Create(newValue)
	return handleError(result, "insert")
}

// BatchInsert inserts multiple records into the database in batches.
func (r *Repo[T]) BatchInsert(ctx context.Context, db *gorm.DB, newValues []*T, batchSize int) error {
	if len(newValues) == 0 {
		return fmt.Errorf("batch insert failed: newValues is empty")
	}
	for i, newValue := range newValues {
		if newValue == nil {
			return fmt.Errorf("batch insert failed: newValue at index %d is nil", i)
		}
	}
	if batchSize <= 0 {
		batchSize = 10
	}
	result := db.WithContext(ctx).CreateInBatches(newValues, batchSize)
	return handleError(result, "batch insert")
}

// Update modifies existing records matching the scopes in the database.
func (r *Repo[T]) Update(ctx context.Context, db *gorm.DB, newValue *T, scopes ...func(db *gorm.DB) *gorm.DB) error {
	if newValue == nil {
		return fmt.Errorf("update failed: newValue is nil")
	}
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Updates(newValue)
	return handleError(result, "update")
}

// UpdateFields modifies specific fields of records matching the scopes.
func (r *Repo[T]) UpdateFields(ctx context.Context, db *gorm.DB, newValue map[string]any, scopes ...func(db *gorm.DB) *gorm.DB) error {
	if len(newValue) == 0 {
		return fmt.Errorf("update fields failed: newValue is empty")
	}
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Updates(newValue)
	return handleError(result, "update fields")
}

// FindOne retrieves one record matching the provided scopes.
func (r *Repo[T]) FindOne(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (*T, error) {
	var record T
	result := db.WithContext(ctx).Scopes(scopes...).Limit(1).Find(&record)
	if result.Error != nil {
		return nil, errors.Join(ErrDatabase, fmt.Errorf("find one failed: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &record, nil
}

// Find retrieves multiple records matching the provided scopes.
func (r *Repo[T]) Find(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) ([]T, error) {
	var records []T
	result := db.WithContext(ctx).Scopes(scopes...).Find(&records)
	if result.Error != nil {
		return nil, errors.Join(ErrDatabase, fmt.Errorf("find failed: %w", result.Error))
	}
	return records, nil
}

// Count returns the number of records matching the provided scopes.
func (r *Repo[T]) Count(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Count(&count)
	if result.Error != nil {
		return 0, errors.Join(ErrDatabase, fmt.Errorf("count failed: %w", result.Error))
	}
	return count, nil
}

// Delete removes records from the database matching the scopes.
func (r *Repo[T]) Delete(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) error {
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Delete(new(T))
	return handleError(result, "delete")
}
