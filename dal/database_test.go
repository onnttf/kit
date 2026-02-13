package dal

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"
)

// Test model for database operations
type TestUser struct {
	ID    int    `gorm:"primarykey"`
	Name  string `gorm:"column:name"`
	Email string `gorm:"column:email"`
	Age   int    `gorm:"column:age"`
}

func (TestUser) TableName() string {
	return "test_users"
}

// Test Suite for Error Variables

func TestErrorVariables(t *testing.T) {
	t.Run("ErrDatabase exists", func(t *testing.T) {
		if ErrDatabase == nil {
			t.Error("ErrDatabase should not be nil")
		}

		if ErrDatabase.Error() == "" {
			t.Error("ErrDatabase should have a message")
		}
	})

	t.Run("ErrNoRowsAffected exists", func(t *testing.T) {
		if ErrNoRowsAffected == nil {
			t.Error("ErrNoRowsAffected should not be nil")
		}

		if ErrNoRowsAffected.Error() == "" {
			t.Error("ErrNoRowsAffected should have a message")
		}
	})

	t.Run("ErrNotFound exists", func(t *testing.T) {
		if ErrNotFound == nil {
			t.Error("ErrNotFound should not be nil")
		}

		if ErrNotFound.Error() == "" {
			t.Error("ErrNotFound should have a message")
		}
	})

	t.Run("errors are distinct", func(t *testing.T) {
		if errors.Is(ErrDatabase, ErrNoRowsAffected) {
			t.Error("ErrDatabase and ErrNoRowsAffected should be distinct")
		}

		if errors.Is(ErrDatabase, ErrNotFound) {
			t.Error("ErrDatabase and ErrNotFound should be distinct")
		}

		if errors.Is(ErrNoRowsAffected, ErrNotFound) {
			t.Error("ErrNoRowsAffected and ErrNotFound should be distinct")
		}
	})
}

// Test Suite for NewRepo Function

func TestNewRepo(t *testing.T) {
	repo := NewRepo[TestUser]()

	if repo == nil {
		t.Fatal("NewRepo should return non-nil repository")
	}
}

// Test Suite for Insert Method

func TestRepo_Insert_NilInput(t *testing.T) {
	repo := NewRepo[TestUser]()
	ctx := context.Background()
	db := &gorm.DB{}

	err := repo.Insert(ctx, db, nil)
	if err == nil {
		t.Fatal("Expected error for nil input")
	}

	if err.Error() != "insert: input is nil" {
		t.Errorf("Expected 'insert: input is nil', got %v", err)
	}
}

func TestRepo_Insert_ValidInput(t *testing.T) {
	t.Skip("Requires real database connection - skipping")

	repo := NewRepo[TestUser]()
	ctx := context.Background()
	db := &gorm.DB{}

	user := &TestUser{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   30,
	}

	// Note: This will fail without a real DB connection
	// In production code, you would use a test database or mock
	_ = repo.Insert(ctx, db, user)
}

// Test Suite for BatchInsert Method

func TestRepo_BatchInsert_EmptyInput(t *testing.T) {
	repo := NewRepo[TestUser]()
	ctx := context.Background()
	db := &gorm.DB{}

	err := repo.BatchInsert(ctx, db, []*TestUser{}, 10)
	if err == nil {
		t.Fatal("Expected error for empty input")
	}

	if err.Error() != "batch insert: input is empty" {
		t.Errorf("Expected 'batch insert: input is empty', got %v", err)
	}
}

func TestRepo_BatchInsert_NilElement(t *testing.T) {
	repo := NewRepo[TestUser]()
	ctx := context.Background()
	db := &gorm.DB{}

	users := []*TestUser{
		{Name: "Alice"},
		nil,
		{Name: "Bob"},
	}

	err := repo.BatchInsert(ctx, db, users, 10)
	if err == nil {
		t.Fatal("Expected error for nil element")
	}

	expectedMsg := "batch insert: input at index 1 is nil"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got %v", expectedMsg, err)
	}
}

func TestRepo_BatchInsert_DefaultBatchSize(t *testing.T) {
	t.Skip("Requires real database connection - skipping")

	repo := NewRepo[TestUser]()
	ctx := context.Background()
	db := &gorm.DB{}

	users := []*TestUser{
		{Name: "Alice"},
		{Name: "Bob"},
	}

	// Zero or negative batch size should default to 10
	_ = repo.BatchInsert(ctx, db, users, 0)
	_ = repo.BatchInsert(ctx, db, users, -5)
}

// Test Suite for Update Method

func TestRepo_Update_NilInput(t *testing.T) {
	repo := NewRepo[TestUser]()
	ctx := context.Background()
	db := &gorm.DB{}

	err := repo.Update(ctx, db, nil)
	if err == nil {
		t.Fatal("Expected error for nil input")
	}

	if err.Error() != "update: input is nil" {
		t.Errorf("Expected 'update: input is nil', got %v", err)
	}
}

func TestRepo_Update_WithScopes(t *testing.T) {
	t.Skip("Requires real database connection - skipping")

	repo := NewRepo[TestUser]()
	ctx := context.Background()
	db := &gorm.DB{}

	user := &TestUser{Name: "UpdatedName"}

	// Test with scope
	_ = repo.Update(ctx, db, user, Condition("id", 1))
}

// Test Suite for UpdateFields Method

func TestRepo_UpdateFields_EmptyInput(t *testing.T) {
	repo := NewRepo[TestUser]()
	ctx := context.Background()
	db := &gorm.DB{}

	err := repo.UpdateFields(ctx, db, map[string]any{})
	if err == nil {
		t.Fatal("Expected error for empty input")
	}

	if err.Error() != "update fields: input is empty" {
		t.Errorf("Expected 'update fields: input is empty', got %v", err)
	}
}

func TestRepo_UpdateFields_ValidInput(t *testing.T) {
	t.Skip("Requires real database connection - skipping")

	repo := NewRepo[TestUser]()
	ctx := context.Background()
	db := &gorm.DB{}

	fields := map[string]any{
		"name": "NewName",
		"age":  35,
	}

	_ = repo.UpdateFields(ctx, db, fields, Condition("id", 1))
}

// Test Suite for QueryOne Method

func TestRepo_QueryOne_ReturnsErrNotFound(t *testing.T) {
	t.Skip("Requires real database connection - skipping")

	repo := NewRepo[TestUser]()
	ctx := context.Background()

	// Create a mock DB that simulates no rows found
	// In real tests, you would use a test database
	db := &gorm.DB{
		Statement: &gorm.Statement{},
	}

	_, err := repo.QueryOne(ctx, db, Condition("id", 999))

	// Without a real DB, this will fail with a different error
	// But in a real test with a DB, you would check:
	// if !errors.Is(err, ErrNotFound) {
	//     t.Errorf("Expected ErrNotFound, got %v", err)
	// }
	_ = err
}

// Test Suite for Query Method

func TestRepo_Query_ReturnsEmptySlice(t *testing.T) {
	t.Skip("Requires real database connection - skipping")

	repo := NewRepo[TestUser]()
	ctx := context.Background()
	db := &gorm.DB{
		Statement: &gorm.Statement{},
	}

	// Query with conditions
	_, err := repo.Query(ctx, db, Condition("age", 999))

	// Without a real DB, this will likely error
	// But in a real test, an empty result set should not error
	_ = err
}

func TestRepo_Query_WithMultipleScopes(t *testing.T) {
	t.Skip("Requires real database connection - skipping")

	repo := NewRepo[TestUser]()
	ctx := context.Background()
	db := &gorm.DB{
		Statement: &gorm.Statement{},
	}

	// Query with multiple scopes
	_, err := repo.Query(ctx, db,
		Condition("age", 30),
		OrderBy("name", "asc"),
		Limit(10),
	)

	_ = err
}

// Test Suite for Count Method

func TestRepo_Count_ReturnsZero(t *testing.T) {
	t.Skip("Requires real database connection - skipping")

	repo := NewRepo[TestUser]()
	ctx := context.Background()
	db := &gorm.DB{
		Statement: &gorm.Statement{},
	}

	// Count with conditions
	_, err := repo.Count(ctx, db, Condition("age", 999))

	// Without a real DB, this will error
	// But in a real test, zero count should not error
	_ = err
}

// Test Suite for Delete Method

func TestRepo_Delete_WithScopes(t *testing.T) {
	t.Skip("Requires real database connection - skipping")

	repo := NewRepo[TestUser]()
	ctx := context.Background()
	db := &gorm.DB{
		Statement: &gorm.Statement{},
	}

	// Delete with conditions
	err := repo.Delete(ctx, db, Condition("id", 1))

	// Without a real DB, this will error
	_ = err
}

// Test Suite for Repository Interface Implementation

func TestRepo_ImplementsRepository(t *testing.T) {
	var _ Repository[TestUser] = &Repo[TestUser]{}

	// This test ensures that Repo implements the Repository interface
	// If it doesn't, the code won't compile
}

// Test Suite for Generic Type Parameters

func TestRepo_WithDifferentTypes(t *testing.T) {
	type Product struct {
		ID    int
		Name  string
		Price float64
	}

	t.Run("repository with Product type", func(t *testing.T) {
		repo := NewRepo[Product]()
		if repo == nil {
			t.Error("Expected non-nil repository")
		}
	})

	type Order struct {
		ID         int
		CustomerID int
		Total      float64
	}

	t.Run("repository with Order type", func(t *testing.T) {
		repo := NewRepo[Order]()
		if repo == nil {
			t.Error("Expected non-nil repository")
		}
	})
}

// Test Suite for Context Handling

func TestRepo_ContextCancellation(t *testing.T) {
	t.Skip("Requires real database connection - skipping")

	repo := NewRepo[TestUser]()
	db := &gorm.DB{
		Statement: &gorm.Statement{},
	}

	t.Run("insert with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		user := &TestUser{Name: "Alice"}
		err := repo.Insert(ctx, db, user)

		// With a real DB, this should respect cancellation
		_ = err
	})

	t.Run("query with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := repo.Query(ctx, db)
		_ = err
	})
}

// Error Handling Tests

func TestHandleExecError_Usage(t *testing.T) {
	t.Run("should detect no rows affected", func(t *testing.T) {
		result := &gorm.DB{
			RowsAffected: 0,
			Error:        nil,
		}

		err := handleExecError(result, "test operation")
		if err == nil {
			t.Error("Expected error for 0 rows affected")
		}

		if !errors.Is(err, ErrNoRowsAffected) {
			t.Errorf("Expected ErrNoRowsAffected, got %v", err)
		}
	})

	t.Run("should detect database error", func(t *testing.T) {
		dbErr := errors.New("connection timeout")
		result := &gorm.DB{
			RowsAffected: 0,
			Error:        dbErr,
		}

		err := handleExecError(result, "test operation")
		if err == nil {
			t.Error("Expected error for database error")
		}

		if !errors.Is(err, ErrDatabase) {
			t.Error("Expected wrapped ErrDatabase")
		}
	})
}

func TestHandleQueryError_Usage(t *testing.T) {
	t.Run("should not error on zero rows", func(t *testing.T) {
		result := &gorm.DB{
			RowsAffected: 0,
			Error:        nil,
		}

		err := handleQueryError(result, "test query")
		if err != nil {
			t.Errorf("Query with 0 rows should not error, got %v", err)
		}
	})

	t.Run("should detect database error", func(t *testing.T) {
		dbErr := errors.New("connection lost")
		result := &gorm.DB{
			RowsAffected: 10,
			Error:        dbErr,
		}

		err := handleQueryError(result, "test query")
		if err == nil {
			t.Error("Expected error for database error")
		}

		if !errors.Is(err, ErrDatabase) {
			t.Error("Expected wrapped ErrDatabase")
		}
	})
}

// Documentation and Examples Tests

func ExampleRepo_Insert() {
	repo := NewRepo[TestUser]()
	ctx := context.Background()

	// Assuming db is a *gorm.DB instance
	var db *gorm.DB

	user := &TestUser{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   30,
	}

	err := repo.Insert(ctx, db, user)
	if err != nil {
		// Handle error
		return
	}

	// user.ID will be populated after successful insert
}

func ExampleRepo_QueryOne() {
	repo := NewRepo[TestUser]()
	ctx := context.Background()

	var db *gorm.DB

	user, err := repo.QueryOne(ctx, db, Condition("email", "alice@example.com"))
	if errors.Is(err, ErrNotFound) {
		// Handle not found case
		return
	}
	if err != nil {
		// Handle other errors
		return
	}

	_ = user // Use the found user
}

func ExampleRepo_Query() {
	repo := NewRepo[TestUser]()
	ctx := context.Background()

	var db *gorm.DB

	users, err := repo.Query(ctx, db,
		Condition("age", 30),
		OrderBy("name", "asc"),
		Limit(10),
	)
	if err != nil {
		// Handle error
		return
	}

	_ = users // Process the users
}
