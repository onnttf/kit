package ptr

import (
	"testing"
)

// Test Suite for To Function

// TestTo_BasicTypes verifies To works correctly with basic data types
func TestTo_BasicTypes(t *testing.T) {
	// Test with string
	originalString := "hello"
	stringPtr := To(originalString)

	if stringPtr == nil {
		t.Error("Expected To to return non-nil pointer for string")
		return
	}

	if *stringPtr != originalString {
		t.Errorf("Expected dereferenced pointer to equal %q, got %q", originalString, *stringPtr)
	}

	// Test with integer
	originalInt := 42
	intPtr := To(originalInt)

	if intPtr == nil {
		t.Error("Expected To to return non-nil pointer for int")
		return
	}

	if *intPtr != originalInt {
		t.Errorf("Expected dereferenced pointer to equal %d, got %d", originalInt, *intPtr)
	}

	// Test with boolean
	originalBool := true
	boolPtr := To(originalBool)

	if boolPtr == nil {
		t.Error("Expected To to return non-nil pointer for bool")
		return
	}

	if *boolPtr != originalBool {
		t.Errorf("Expected dereferenced pointer to equal %t, got %t", originalBool, *boolPtr)
	}
}

// TestTo_PointerStability ensures the returned pointer is stable and independent
func TestTo_PointerStability(t *testing.T) {
	originalValue := "test"
	ptr1 := To(originalValue)
	ptr2 := To(originalValue)

	// Each call should return a different pointer address
	if ptr1 == ptr2 {
		t.Error("Expected different pointer addresses for separate To calls")
	}

	// But both should contain the same value
	if *ptr1 != *ptr2 {
		t.Errorf("Expected both pointers to contain same value: %q vs %q", *ptr1, *ptr2)
	}
}

// TestTo_ZeroValues verifies To handles zero values correctly
func TestTo_ZeroValues(t *testing.T) {
	// Test with zero string
	emptyStringPtr := To("")
	if emptyStringPtr == nil {
		t.Error("Expected To to return non-nil pointer for empty string")
		return
	}
	if *emptyStringPtr != "" {
		t.Error("Expected To to handle empty string correctly")
	}

	// Test with zero int
	zeroIntPtr := To(0)
	if zeroIntPtr == nil {
		t.Error("Expected To to return non-nil pointer for zero int")
		return
	}
	if *zeroIntPtr != 0 {
		t.Error("Expected To to handle zero int correctly")
	}

	// Test with zero bool
	falseBoolPtr := To(false)
	if falseBoolPtr == nil {
		t.Error("Expected To to return non-nil pointer for false bool")
		return
	}
	if *falseBoolPtr != false {
		t.Error("Expected To to handle false bool correctly")
	}
}

// Test Suite for DerefOr Function

// TestDerefOr_ValidPointer tests DerefOr with valid non-nil pointers
func TestDerefOr_ValidPointer(t *testing.T) {
	// Test with string pointer
	originalString := "hello world"
	stringPtr := To(originalString)
	result := DerefOr(stringPtr, "default")

	if result != originalString {
		t.Errorf("Expected DerefOr to return %q, got %q", originalString, result)
	}

	// Test with integer pointer
	originalInt := 100
	intPtr := To(originalInt)
	result2 := DerefOr(intPtr, 0)

	if result2 != originalInt {
		t.Errorf("Expected DerefOr to return %d, got %d", originalInt, result2)
	}
}

// TestDerefOr_NilPointer tests DerefOr behavior with nil pointers
func TestDerefOr_NilPointer(t *testing.T) {
	// Test with nil string pointer
	var nilStringPtr *string
	defaultString := "default value"
	result := DerefOr(nilStringPtr, defaultString)

	if result != defaultString {
		t.Errorf("Expected DerefOr to return default %q, got %q", defaultString, result)
	}

	// Test with nil integer pointer
	var nilIntPtr *int
	defaultInt := 42
	result2 := DerefOr(nilIntPtr, defaultInt)

	if result2 != defaultInt {
		t.Errorf("Expected DerefOr to return default %d, got %d", defaultInt, result2)
	}
}

// TestDerefOr_DifferentDefaultValues ensures default values are returned correctly
func TestDerefOr_DifferentDefaultValues(t *testing.T) {
	var nilPtr *string

	// Test with empty string default
	result1 := DerefOr(nilPtr, "")
	if result1 != "" {
		t.Errorf("Expected empty string default, got %q", result1)
	}

	// Test with non-empty string default
	result2 := DerefOr(nilPtr, "fallback")
	if result2 != "fallback" {
		t.Errorf("Expected 'fallback' default, got %q", result2)
	}
}

// Test Suite for ToIf Function

// TestToIf_TrueCondition tests ToIf when condition is true
func TestToIf_TrueCondition(t *testing.T) {
	value := "conditional value"
	result := ToIf(true, value)

	if result == nil {
		t.Error("Expected ToIf to return non-nil pointer when condition is true")
		return
	}

	if *result != value {
		t.Errorf("Expected dereferenced pointer to equal %q, got %q", value, *result)
	}
}

// TestToIf_FalseCondition tests ToIf when condition is false
func TestToIf_FalseCondition(t *testing.T) {
	value := "should not be used"
	result := ToIf(false, value)

	if result != nil {
		t.Error("Expected ToIf to return nil when condition is false")
	}
}

// TestToIf_ConditionalLogic tests ToIf with realistic conditional scenarios
func TestToIf_ConditionalLogic(t *testing.T) {
	// Test with non-empty string condition
	userName := "Alice"
	userPtr := ToIf(userName != "", userName)

	if userPtr == nil {
		t.Error("Expected non-nil pointer for non-empty username")
		return
	}

	if *userPtr != userName {
		t.Errorf("Expected %q, got %q", userName, *userPtr)
	}

	// Test with empty string condition
	emptyName := ""
	emptyPtr := ToIf(emptyName != "", emptyName)

	if emptyPtr != nil {
		t.Error("Expected nil pointer for empty username condition")
	}

	// Test with numeric condition
	score := 85
	scorePtr := ToIf(score > 80, score)

	if scorePtr == nil {
		t.Error("Expected non-nil pointer for score > 80")
		return
	}

	if *scorePtr != score {
		t.Errorf("Expected score %d, got %d", score, *scorePtr)
	}
}

// Test Suite for IsNil Function

// TestIsNil_WithNilPointer tests IsNil with nil pointers
func TestIsNil_WithNilPointer(t *testing.T) {
	var nilStringPtr *string
	var nilIntPtr *int
	var nilBoolPtr *bool

	if !IsNil(nilStringPtr) {
		t.Error("Expected IsNil to return true for nil string pointer")
	}

	if !IsNil(nilIntPtr) {
		t.Error("Expected IsNil to return true for nil int pointer")
	}

	if !IsNil(nilBoolPtr) {
		t.Error("Expected IsNil to return true for nil bool pointer")
	}
}

// TestIsNil_WithValidPointer tests IsNil with valid pointers
func TestIsNil_WithValidPointer(t *testing.T) {
	stringPtr := To("test")
	intPtr := To(42)
	boolPtr := To(true)

	if IsNil(stringPtr) {
		t.Error("Expected IsNil to return false for valid string pointer")
	}

	if IsNil(intPtr) {
		t.Error("Expected IsNil to return false for valid int pointer")
	}

	if IsNil(boolPtr) {
		t.Error("Expected IsNil to return false for valid bool pointer")
	}
}

// Test Suite for IsNotNil Function

// TestIsNotNil_WithValidPointer tests IsNotNil with valid pointers
func TestIsNotNil_WithValidPointer(t *testing.T) {
	stringPtr := To("test")
	intPtr := To(42)
	boolPtr := To(false)

	if !IsNotNil(stringPtr) {
		t.Error("Expected IsNotNil to return true for valid string pointer")
	}

	if !IsNotNil(intPtr) {
		t.Error("Expected IsNotNil to return true for valid int pointer")
	}

	if !IsNotNil(boolPtr) {
		t.Error("Expected IsNotNil to return true for valid bool pointer")
	}
}

// TestIsNotNil_WithNilPointer tests IsNotNil with nil pointers
func TestIsNotNil_WithNilPointer(t *testing.T) {
	var nilStringPtr *string
	var nilIntPtr *int
	var nilBoolPtr *bool

	if IsNotNil(nilStringPtr) {
		t.Error("Expected IsNotNil to return false for nil string pointer")
	}

	if IsNotNil(nilIntPtr) {
		t.Error("Expected IsNotNil to return false for nil int pointer")
	}

	if IsNotNil(nilBoolPtr) {
		t.Error("Expected IsNotNil to return false for nil bool pointer")
	}
}

// Test Suite for Zero Function

// TestZero_BasicTypes tests Zero with basic types
func TestZero_BasicTypes(t *testing.T) {
	// Test with int
	intZeroPtr := Zero[int]()
	if intZeroPtr == nil {
		t.Error("Expected Zero to return non-nil pointer for int")
		return
	}

	if *intZeroPtr != 0 {
		t.Errorf("Expected zero int value 0, got %d", *intZeroPtr)
	}

	// Test with string
	stringZeroPtr := Zero[string]()
	if stringZeroPtr == nil {
		t.Error("Expected Zero to return non-nil pointer for string")
		return
	}

	if *stringZeroPtr != "" {
		t.Errorf("Expected zero string value '', got %q", *stringZeroPtr)
	}

	// Test with bool
	boolZeroPtr := Zero[bool]()
	if boolZeroPtr == nil {
		t.Error("Expected Zero to return non-nil pointer for bool")
		return
	}

	if *boolZeroPtr != false {
		t.Errorf("Expected zero bool value false, got %t", *boolZeroPtr)
	}
}

// TestZero_PointerIndependence ensures each Zero call returns independent pointers
func TestZero_PointerIndependence(t *testing.T) {
	ptr1 := Zero[int]()
	ptr2 := Zero[int]()

	// Should be different pointer addresses
	if ptr1 == ptr2 {
		t.Error("Expected different pointer addresses for separate Zero calls")
	}

	// But both should point to zero values
	if *ptr1 != 0 || *ptr2 != 0 {
		t.Error("Expected both pointers to point to zero values")
	}

	// Modifying one should not affect the other
	*ptr1 = 100
	if *ptr2 != 0 {
		t.Error("Expected ptr2 to remain unaffected when ptr1 is modified")
	}
}

// Integration and Edge Case Tests

// TestIntegration_WorkflowExample demonstrates typical usage patterns
func TestIntegration_WorkflowExample(t *testing.T) {
	// Simulate a struct with optional fields
	type User struct {
		Name  string
		Email *string
		Age   *int
	}

	// Create user with conditional optional fields
	userName := "John Doe"
	userEmail := "john@example.com"
	userAge := 30

	user := User{
		Name:  userName,
		Email: ToIf(userEmail != "", userEmail),
		Age:   ToIf(userAge > 0, userAge),
	}

	// Verify the user was created correctly
	if user.Name != userName {
		t.Errorf("Expected name %q, got %q", userName, user.Name)
	}

	if IsNil(user.Email) {
		t.Error("Expected non-nil email pointer")
		return
	}

	if DerefOr(user.Email, "") != userEmail {
		t.Errorf("Expected email %q, got %q", userEmail, DerefOr(user.Email, ""))
	}

	if IsNil(user.Age) {
		t.Error("Expected non-nil age pointer")
		return
	}

	if DerefOr(user.Age, 0) != userAge {
		t.Errorf("Expected age %d, got %d", userAge, DerefOr(user.Age, 0))
	}
}

// TestEdgeCases_ComplexTypes tests with complex types like slices and maps
func TestEdgeCases_ComplexTypes(t *testing.T) {
	// Test with slice
	originalSlice := []int{1, 2, 3}
	slicePtr := To(originalSlice)

	if IsNil(slicePtr) {
		t.Error("Expected non-nil slice pointer")
		return
	}

	retrievedSlice := DerefOr(slicePtr, []int{})
	if len(retrievedSlice) != len(originalSlice) {
		t.Errorf("Expected slice length %d, got %d", len(originalSlice), len(retrievedSlice))
	}

	// Test with map
	originalMap := map[string]int{"key": 42}
	mapPtr := To(originalMap)

	if IsNil(mapPtr) {
		t.Error("Expected non-nil map pointer")
		return
	}

	retrievedMap := DerefOr(mapPtr, map[string]int{})
	if retrievedMap["key"] != 42 {
		t.Errorf("Expected map value 42, got %d", retrievedMap["key"])
	}
}

// TestEdgeCases_NilComparison tests nil comparison functions consistency
func TestEdgeCases_NilComparison(t *testing.T) {
	var nilPtr *string
	validPtr := To("test")

	// Test that IsNil and IsNotNil are logical opposites
	if IsNil(nilPtr) == IsNotNil(nilPtr) {
		t.Error("Expected IsNil and IsNotNil to return opposite values for nil pointer")
	}

	if IsNil(validPtr) == IsNotNil(validPtr) {
		t.Error("Expected IsNil and IsNotNil to return opposite values for valid pointer")
	}

	// Test consistency
	if IsNil(nilPtr) != true {
		t.Error("Expected IsNil to return true for nil pointer")
	}

	if IsNotNil(validPtr) != true {
		t.Error("Expected IsNotNil to return true for valid pointer")
	}
}
