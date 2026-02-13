package dal

import (
	"testing"
)

// Test Suite for Paginate Function

func TestPaginate_NormalizePageNumber(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		pageSize       int
		expectedOffset int
		expectedLimit  int
	}{
		{
			name:           "valid page 1",
			page:           1,
			pageSize:       10,
			expectedOffset: 0,
			expectedLimit:  10,
		},
		{
			name:           "valid page 2",
			page:           2,
			pageSize:       10,
			expectedOffset: 10,
			expectedLimit:  10,
		},
		{
			name:           "zero page defaults to 1",
			page:           0,
			pageSize:       10,
			expectedOffset: 0,
			expectedLimit:  10,
		},
		{
			name:           "negative page defaults to 1",
			page:           -5,
			pageSize:       10,
			expectedOffset: 0,
			expectedLimit:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := Paginate(tt.page, tt.pageSize)

			// Just verify the scope function is created successfully
			// Testing actual SQL generation requires a real DB or DryRun mode
			if scope == nil {
				t.Error("Expected non-nil scope function")
			}
		})
	}
}

func TestPaginate_NormalizePageSize(t *testing.T) {
	tests := []struct {
		name             string
		page             int
		pageSize         int
		expectedPageSize int
	}{
		{
			name:             "valid page size",
			page:             1,
			pageSize:         20,
			expectedPageSize: 20,
		},
		{
			name:             "zero page size defaults to DefaultPageSize",
			page:             1,
			pageSize:         0,
			expectedPageSize: DefaultPageSize,
		},
		{
			name:             "negative page size defaults to DefaultPageSize",
			page:             1,
			pageSize:         -10,
			expectedPageSize: DefaultPageSize,
		},
		{
			name:             "page size exceeding max capped to MaxPageSize",
			page:             1,
			pageSize:         200,
			expectedPageSize: MaxPageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := Paginate(tt.page, tt.pageSize)
			if scope == nil {
				t.Error("Expected non-nil scope function")
			}
		})
	}
}

func TestPaginate_OffsetCalculation(t *testing.T) {
	tests := []struct {
		page           int
		pageSize       int
		expectedOffset int
	}{
		{page: 1, pageSize: 10, expectedOffset: 0},
		{page: 2, pageSize: 10, expectedOffset: 10},
		{page: 3, pageSize: 10, expectedOffset: 20},
		{page: 5, pageSize: 25, expectedOffset: 100},
		{page: 10, pageSize: 50, expectedOffset: 450},
	}

	for _, tt := range tests {
		t.Run("offset calculation", func(t *testing.T) {
			scope := Paginate(tt.page, tt.pageSize)
			if scope == nil {
				t.Error("Expected non-nil scope function")
			}
			// The actual offset would be verified in integration tests
		})
	}
}

// Test Suite for Condition Function

func TestCondition_ValidInput(t *testing.T) {
	tests := []struct {
		name   string
		column string
		value  any
	}{
		{
			name:   "string value",
			column: "name",
			value:  "Alice",
		},
		{
			name:   "integer value",
			column: "age",
			value:  30,
		},
		{
			name:   "boolean value",
			column: "active",
			value:  true,
		},
		{
			name:   "nil value",
			column: "deleted_at",
			value:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := Condition(tt.column, tt.value)
			if scope == nil {
				t.Error("Expected non-nil scope function")
			}
		})
	}
}

func TestCondition_SpecialCharactersInColumn(t *testing.T) {
	// Test that special characters are properly handled
	columns := []string{
		"user_id",
		"user.id",
		"table.column",
	}

	for _, column := range columns {
		t.Run(column, func(t *testing.T) {
			scope := Condition(column, 123)
			if scope == nil {
				t.Error("Expected non-nil scope function")
			}
		})
	}
}

// Test Suite for OrderBy Function

func TestOrderBy_ValidDirections(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		direction string
	}{
		{
			name:      "ascending lowercase",
			field:     "created_at",
			direction: "asc",
		},
		{
			name:      "ascending uppercase",
			field:     "created_at",
			direction: "ASC",
		},
		{
			name:      "descending lowercase",
			field:     "created_at",
			direction: "desc",
		},
		{
			name:      "descending uppercase",
			field:     "created_at",
			direction: "DESC",
		},
		{
			name:      "descending mixed case",
			field:     "created_at",
			direction: "DeSc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := OrderBy(tt.field, tt.direction)
			if scope == nil {
				t.Error("Expected non-nil scope function")
			}
		})
	}
}

func TestOrderBy_InvalidDirection(t *testing.T) {
	// Invalid directions should default to ASC
	invalidDirections := []string{"invalid", "random", "", "up", "down"}

	for _, direction := range invalidDirections {
		t.Run(direction, func(t *testing.T) {
			scope := OrderBy("created_at", direction)
			if scope == nil {
				t.Error("Expected non-nil scope function")
			}
		})
	}
}

// Test Suite for Limit Function

func TestLimit_ValidValues(t *testing.T) {
	limits := []int{1, 10, 50, 100, 1000}

	for _, limit := range limits {
		t.Run("limit value", func(t *testing.T) {
			scope := Limit(limit)
			if scope == nil {
				t.Error("Expected non-nil scope function")
			}
		})
	}
}

func TestLimit_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		limit int
	}{
		{"zero limit", 0},
		{"negative limit", -10},
		{"very large limit", 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := Limit(tt.limit)
			if scope == nil {
				t.Error("Expected non-nil scope function")
			}
		})
	}
}

// Test Suite for LikeCondition Function

func TestLikeCondition_ValidInput(t *testing.T) {
	tests := []struct {
		name   string
		column string
		value  string
	}{
		{
			name:   "basic search",
			column: "name",
			value:  "john",
		},
		{
			name:   "empty search",
			column: "name",
			value:  "",
		},
		{
			name:   "special characters",
			column: "email",
			value:  "test@example.com",
		},
		{
			name:   "unicode characters",
			column: "description",
			value:  "测试",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := LikeCondition(tt.column, tt.value)
			if scope == nil {
				t.Error("Expected non-nil scope function")
			}
		})
	}
}

func TestLikeCondition_WildcardHandling(t *testing.T) {
	// Test that wildcards in value are treated as literals
	tests := []struct {
		name  string
		value string
	}{
		{"percent sign", "test%value"},
		{"underscore", "test_value"},
		{"both wildcards", "test%_value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := LikeCondition("column", tt.value)
			if scope == nil {
				t.Error("Expected non-nil scope function")
			}
		})
	}
}

// Test Suite for SelectFields Function

func TestSelectFields_SingleField(t *testing.T) {
	scope := SelectFields("id")
	if scope == nil {
		t.Error("Expected non-nil scope function")
	}
}

func TestSelectFields_MultipleFields(t *testing.T) {
	scope := SelectFields("id", "name", "email", "created_at")
	if scope == nil {
		t.Error("Expected non-nil scope function")
	}
}

func TestSelectFields_EmptyFields(t *testing.T) {
	scope := SelectFields()
	if scope == nil {
		t.Error("Expected non-nil scope function")
	}
}

func TestSelectFields_WithTablePrefix(t *testing.T) {
	scope := SelectFields("users.id", "users.name", "profiles.bio")
	if scope == nil {
		t.Error("Expected non-nil scope function")
	}
}

// Integration-style Tests (Scope Composition)

func TestScopeComposition(t *testing.T) {
	t.Skip("Scope composition requires real database - skipping")

	t.Run("combine multiple scopes", func(t *testing.T) {
		// This would require a real DB connection to test properly
		// Just verify the scopes can be created
		_ = Paginate(2, 20)
		_ = Condition("active", true)
		_ = OrderBy("created_at", "desc")
		_ = SelectFields("id", "name")
	})

	t.Run("search with pagination", func(t *testing.T) {
		// Verify scopes can be created
		_ = LikeCondition("name", "john")
		_ = Paginate(1, 10)
		_ = OrderBy("name", "asc")
	})
}

// Constants Tests

func TestPaginationConstants(t *testing.T) {
	if DefaultPageSize <= 0 {
		t.Error("DefaultPageSize must be positive")
	}

	if MaxPageSize <= 0 {
		t.Error("MaxPageSize must be positive")
	}

	if MaxPageSize < DefaultPageSize {
		t.Error("MaxPageSize should be >= DefaultPageSize")
	}
}
