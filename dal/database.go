package dal

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Repository defines common CRUD operations for database records.
type Repository[T any] interface {
	// Insert adds a new record to the database.
	Insert(ctx context.Context, db *gorm.DB, record *T) error

	// BatchInsert adds multiple records to the database in batches.
	BatchInsert(ctx context.Context, db *gorm.DB, records []*T, batchSize int) error

	// Update modifies an existing record that matches the scopes.
	Update(ctx context.Context, db *gorm.DB, record *T, scopes ...func(db *gorm.DB) *gorm.DB) error

	// UpdateFields modifies specific fields of records matching the scopes.
	UpdateFields(ctx context.Context, db *gorm.DB, updates map[string]any, scopes ...func(db *gorm.DB) *gorm.DB) error

	// FindOne retrieves a single record based on the provided scopes.
	FindOne(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (*T, error)

	// Find retrieves multiple records based on the provided scopes.
	Find(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) ([]T, error)

	// Count returns the number of records that match the provided scopes.
	Count(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (int64, error)

	// Delete removes records from the database that match the scopes.
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
	// ErrDatabase indicates a general database-related error.
	ErrDatabase = errors.New("database error occurred")

	// ErrNoChange is returned when no records are modified in an operation.
	ErrNoChange = errors.New("no records changed")
)

// handleError processes the result of a DB operation and returns appropriate errors.
func handleError(result *gorm.DB, action string) error {
	// If DB error exists, join it with ErrDatabase
	if result.Error != nil {
		return errors.Join(ErrDatabase, fmt.Errorf("%s failed: %w", action, result.Error))
	}

	// If no rows were affected, join it with ErrNoChange
	if result.RowsAffected == 0 {
		return errors.Join(ErrNoChange, fmt.Errorf("during %s", action))
	}

	return nil
}

// Insert adds a new record to the database.
func (r *Repo[T]) Insert(ctx context.Context, db *gorm.DB, record *T) error {
	if record == nil {
		return fmt.Errorf("cannot insert nil record")
	}

	result := db.WithContext(ctx).Create(record)
	return handleError(result, "insert")
}

// BatchInsert inserts multiple records in batches.
func (r *Repo[T]) BatchInsert(ctx context.Context, db *gorm.DB, records []*T, batchSize int) error {
	if len(records) == 0 {
		return fmt.Errorf("cannot insert empty record list")
	}

	// Validate each record in the batch
	for i, record := range records {
		if record == nil {
			return fmt.Errorf("record at index %d is nil", i)
		}
	}

	// Use default batch size if invalid
	if batchSize <= 0 {
		batchSize = 10
	}

	result := db.WithContext(ctx).CreateInBatches(records, batchSize)
	return handleError(result, "batch insert")
}

// Update modifies an existing record in the database.
func (r *Repo[T]) Update(ctx context.Context, db *gorm.DB, record *T, scopes ...func(db *gorm.DB) *gorm.DB) error {
	if record == nil {
		return fmt.Errorf("cannot update with nil record")
	}

	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Updates(record)
	return handleError(result, "update")
}

// UpdateFields modifies specific fields of records matching the scopes.
func (r *Repo[T]) UpdateFields(ctx context.Context, db *gorm.DB, updates map[string]any, scopes ...func(db *gorm.DB) *gorm.DB) error {
	if len(updates) == 0 {
		return fmt.Errorf("cannot update with empty updates map")
	}

	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Updates(updates)
	return handleError(result, "update fields")
}

// FindOne retrieves a single record based on the provided scopes.
func (r *Repo[T]) FindOne(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (*T, error) {
	var record T
	result := db.WithContext(ctx).Scopes(scopes...).Limit(1).Find(&record)

	if result.Error != nil {
		return nil, errors.Join(ErrDatabase, fmt.Errorf("failed to find record: %w", result.Error))
	}

	if result.RowsAffected == 0 {
		return nil, nil // No record found
	}

	return &record, nil
}

// Find retrieves multiple records based on the provided scopes.
func (r *Repo[T]) Find(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) ([]T, error) {
	var records []T
	result := db.WithContext(ctx).Scopes(scopes...).Find(&records)

	if result.Error != nil {
		return nil, errors.Join(ErrDatabase, fmt.Errorf("failed to find records: %w", result.Error))
	}

	return records, nil
}

// Count returns the number of records that match the provided scopes.
func (r *Repo[T]) Count(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Count(&count)

	if result.Error != nil {
		return 0, errors.Join(ErrDatabase, fmt.Errorf("failed to count records: %w", result.Error))
	}

	return count, nil
}

// Delete removes records from the database that match the scopes.
func (r *Repo[T]) Delete(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) error {
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Delete(new(T))
	return handleError(result, "delete")
}
