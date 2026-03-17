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

	// Delete removes entities matching the provided scopes from the database.
	Delete(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) error
}

// Repo provides a generic implementation of the Repository interface
type Repo[T any] struct{}

// NewRepo creates a new repository instance for type T
func NewRepo[T any]() *Repo[T] {
	return &Repo[T]{}
}

// ErrDatabase indicates an unexpected database operation failure.
var ErrDatabase = errors.New("unexpected database error")

// ErrNoRowsAffected indicates a write operation did not affect any rows.
var ErrNoRowsAffected = errors.New("no rows affected")

// ErrNotFound indicates no records were found for a query.
var ErrNotFound = errors.New("record not found")

// handleExecError evaluates a GORM write operation and returns an error on failure.
func handleExecError(result *gorm.DB, action string) error {
	if result.Error != nil {
		dbErr := fmt.Errorf("%w: %v", ErrDatabase, result.Error)
		return fmt.Errorf("%s: %w", action, dbErr)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%s: %w", action, ErrNoRowsAffected)
	}
	return nil
}

// handleQueryError evaluates a GORM query operation and returns an error on failure.
func handleQueryError(result *gorm.DB, action string) error {
	if result.Error != nil {
		dbErr := fmt.Errorf("%w: %v", ErrDatabase, result.Error)
		return fmt.Errorf("%s: %w", action, dbErr)
	}
	return nil
}

// Insert adds a new entity to the database, returning an error if the input is nil.
func (r *Repo[T]) Insert(ctx context.Context, db *gorm.DB, newValue *T) error {
	if newValue == nil {
		return errors.New("newValue is nil")
	}
	result := db.WithContext(ctx).Create(newValue)
	return handleExecError(result, "insert")
}

// BatchInsert adds multiple entities to the database in batches, using a default batch size of 10 if unspecified.
func (r *Repo[T]) BatchInsert(ctx context.Context, db *gorm.DB, newValues []*T, batchSize int) error {
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

// Update modifies an existing entity in the database, applying the specified scopes for filtering.
func (r *Repo[T]) Update(ctx context.Context, db *gorm.DB, newValue *T, scopes ...func(db *gorm.DB) *gorm.DB) error {
	if newValue == nil {
		return errors.New("newValue is nil")
	}
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Updates(newValue)
	return handleExecError(result, "update")
}

// UpdateFields modifies specific fields of entities in the database, applying the specified scopes.
func (r *Repo[T]) UpdateFields(ctx context.Context, db *gorm.DB, newValue map[string]any, scopes ...func(db *gorm.DB) *gorm.DB) error {
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
//	user, err := repo.QueryOne(ctx, db, Condition("id", 123))
//	if errors.Is(err, dal.ErrNotFound) {
//	    // handle not found case
//	}
func (r *Repo[T]) QueryOne(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (*T, error) {
	var record T
	result := db.WithContext(ctx).Scopes(scopes...).Limit(1).Find(&record)
	if err := handleQueryError(result, "query one"); err != nil {
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &record, nil
}

// Query retrieves multiple entities from the database matching the specified scopes.
// Returns an empty slice (not an error) if no matching records are found.
func (r *Repo[T]) Query(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) ([]T, error) {
	var records []T
	result := db.WithContext(ctx).Scopes(scopes...).Find(&records)
	if err := handleQueryError(result, "query"); err != nil {
		return nil, err
	}
	return records, nil
}

// Count returns the number of entities in the database matching the specified scopes.
// Returns 0 (not an error) if no matching records are found.
func (r *Repo[T]) Count(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Count(&count)
	if err := handleQueryError(result, "count"); err != nil {
		return 0, err
	}
	return count, nil
}

// Delete removes entities from the database matching the specified scopes.
func (r *Repo[T]) Delete(ctx context.Context, db *gorm.DB, scopes ...func(db *gorm.DB) *gorm.DB) error {
	result := db.WithContext(ctx).Model(new(T)).Scopes(scopes...).Delete(new(T))
	return handleExecError(result, "delete")
}
