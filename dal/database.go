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

	// Update modifies an existing record that matches the conditions.
	Update(ctx context.Context, db *gorm.DB, record *T, conditions ...func(db *gorm.DB) *gorm.DB) error

	// UpdateFields modifies specific fields of records matching the conditions.
	UpdateFields(ctx context.Context, db *gorm.DB, updates map[string]any, conditions ...func(db *gorm.DB) *gorm.DB) error

	// FindOne retrieves a single record based on the provided conditions.
	FindOne(ctx context.Context, db *gorm.DB, conditions ...func(db *gorm.DB) *gorm.DB) (*T, error)

	// FindMany retrieves multiple records based on the provided conditions.
	FindMany(ctx context.Context, db *gorm.DB, conditions ...func(db *gorm.DB) *gorm.DB) ([]T, error)

	// Count returns the number of records that match the provided conditions.
	Count(ctx context.Context, db *gorm.DB, conditions ...func(db *gorm.DB) *gorm.DB) (int64, error)

	// Delete removes records from the database that match the conditions.
	Delete(ctx context.Context, db *gorm.DB, conditions ...func(db *gorm.DB) *gorm.DB) error
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
func (r *Repo[T]) Update(ctx context.Context, db *gorm.DB, record *T, conditions ...func(db *gorm.DB) *gorm.DB) error {
	if record == nil {
		return fmt.Errorf("cannot update with nil record")
	}

	result := db.WithContext(ctx).Model(new(T)).Scopes(conditions...).Updates(record)
	return handleError(result, "update")
}

// UpdateFields modifies specific fields of records matching the conditions.
func (r *Repo[T]) UpdateFields(ctx context.Context, db *gorm.DB, updates map[string]any, conditions ...func(db *gorm.DB) *gorm.DB) error {
	if updates == nil || len(updates) == 0 {
		return fmt.Errorf("cannot update with empty updates map")
	}

	result := db.WithContext(ctx).Model(new(T)).Scopes(conditions...).Updates(updates)
	return handleError(result, "update fields")
}

// Delete removes records from the database that match the conditions.
func (r *Repo[T]) Delete(ctx context.Context, db *gorm.DB, conditions ...func(db *gorm.DB) *gorm.DB) error {
	result := db.WithContext(ctx).Model(new(T)).Scopes(conditions...).Delete(new(T))
	return handleError(result, "delete")
}

// FindOne retrieves a single record based on the provided conditions.
func (r *Repo[T]) FindOne(ctx context.Context, db *gorm.DB, conditions ...func(db *gorm.DB) *gorm.DB) (*T, error) {
	var record T
	result := db.WithContext(ctx).Scopes(conditions...).Limit(1).Find(&record)

	if result.Error != nil {
		return nil, errors.Join(ErrDatabase, fmt.Errorf("failed to find record: %w", result.Error))
	}

	if result.RowsAffected == 0 {
		return nil, nil // No record found
	}

	return &record, nil
}

// FindMany retrieves multiple records based on the provided conditions.
func (r *Repo[T]) FindMany(ctx context.Context, db *gorm.DB, conditions ...func(db *gorm.DB) *gorm.DB) ([]T, error) {
	var records []T
	result := db.WithContext(ctx).Scopes(conditions...).Find(&records)

	if result.Error != nil {
		return nil, errors.Join(ErrDatabase, fmt.Errorf("failed to find records: %w", result.Error))
	}

	return records, nil
}

// Count returns the number of records that match the provided conditions.
func (r *Repo[T]) Count(ctx context.Context, db *gorm.DB, conditions ...func(db *gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	result := db.WithContext(ctx).Model(new(T)).Scopes(conditions...).Count(&count)

	if result.Error != nil {
		return 0, errors.Join(ErrDatabase, fmt.Errorf("failed to count records: %w", result.Error))
	}

	return count, nil
}

// Pagination constants
const (
	DefaultPageSize = 10  // Default number of items per page
	MaxPageSize     = 100 // Maximum allowed items per page
)

// Paginate returns a GORM scope function that applies pagination.
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	// Normalize page number
	if page <= 0 {
		page = 1
	}

	// Normalize page size
	switch {
	case pageSize <= 0:
		pageSize = DefaultPageSize
	case pageSize > MaxPageSize:
		pageSize = MaxPageSize
	}

	offset := (page - 1) * pageSize

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(pageSize)
	}
}
