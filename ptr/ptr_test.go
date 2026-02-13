package ptr

import (
	"testing"
)

// Test Suite for PtrOf Function

// TestPtrOf_BasicTypes verifies PtrOf works correctly with basic data types
func TestPtrOf_BasicTypes(t *testing.T) {
	// Test with string
	originalString := "hello"
	stringPtr := PtrOf(originalString)

	if stringPtr == nil {
		t.Error("Expected PtrOf to return non-nil pointer for string")
		return
	}

	if *stringPtr != originalString {
		t.Errorf("Expected dereferenced pointer to equal %q, got %q", originalString, *stringPtr)
	}

	// Test with integer
	originalInt := 42
	intPtr := PtrOf(originalInt)

	if intPtr == nil {
		t.Error("Expected PtrOf to return non-nil pointer for int")
		return
	}

	if *intPtr != originalInt {
		t.Errorf("Expected dereferenced pointer to equal %d, got %d", originalInt, *intPtr)
	}

	// Test with boolean
	originalBool := true
	boolPtr := PtrOf(originalBool)

	if boolPtr == nil {
		t.Error("Expected PtrOf to return non-nil pointer for bool")
		return
	}

	if *boolPtr != originalBool {
		t.Errorf("Expected dereferenced pointer to equal %t, got %t", originalBool, *boolPtr)
	}
}

// TestPtrOf_PointerStability ensures the returned pointer is stable and independent
func TestPtrOf_PointerStability(t *testing.T) {
	originalValue := "test"
	ptr1 := PtrOf(originalValue)
	ptr2 := PtrOf(originalValue)

	// Each call should return a different pointer address
	if ptr1 == ptr2 {
		t.Error("Expected different pointer addresses for separate PtrOf calls")
	}

	// But both should contain the same value
	if *ptr1 != *ptr2 {
		t.Errorf("Expected both pointers to contain same value: %q vs %q", *ptr1, *ptr2)
	}
}

// TestPtrOf_ZeroValues verifies PtrOf handles zero values correctly
func TestPtrOf_ZeroValues(t *testing.T) {
	// Test with zero string
	emptyStringPtr := PtrOf("")
	if emptyStringPtr == nil {
		t.Error("Expected PtrOf to return non-nil pointer for empty string")
		return
	}
	if *emptyStringPtr != "" {
		t.Error("Expected PtrOf to handle empty string correctly")
	}

	// Test with zero int
	zeroIntPtr := PtrOf(0)
	if zeroIntPtr == nil {
		t.Error("Expected PtrOf to return non-nil pointer for zero int")
		return
	}
	if *zeroIntPtr != 0 {
		t.Error("Expected PtrOf to handle zero int correctly")
	}

	// Test with zero bool
	falseBoolPtr := PtrOf(false)
	if falseBoolPtr == nil {
		t.Error("Expected PtrOf to return non-nil pointer for false bool")
		return
	}
	if *falseBoolPtr != false {
		t.Error("Expected PtrOf to handle false bool correctly")
	}
}

// Test Suite for ValueOf Function

// TestValueOf_ValidPointer tests ValueOf with valid non-nil pointers
func TestValueOf_ValidPointer(t *testing.T) {
	// Test with string pointer
	originalString := "hello world"
	stringPtr := PtrOf(originalString)
	result := ValueOf(stringPtr, "default")

	if result != originalString {
		t.Errorf("Expected ValueOf to return %q, got %q", originalString, result)
	}

	// Test with integer pointer
	originalInt := 100
	intPtr := PtrOf(originalInt)
	result2 := ValueOf(intPtr, 0)

	if result2 != originalInt {
		t.Errorf("Expected ValueOf to return %d, got %d", originalInt, result2)
	}
}

// TestValueOf_NilPointer tests ValueOf behavior with nil pointers
func TestValueOf_NilPointer(t *testing.T) {
	// Test with nil string pointer
	var nilStringPtr *string
	defaultString := "default value"
	result := ValueOf(nilStringPtr, defaultString)

	if result != defaultString {
		t.Errorf("Expected ValueOf to return default %q, got %q", defaultString, result)
	}

	// Test with nil integer pointer
	var nilIntPtr *int
	defaultInt := 42
	result2 := ValueOf(nilIntPtr, defaultInt)

	if result2 != defaultInt {
		t.Errorf("Expected ValueOf to return default %d, got %d", defaultInt, result2)
	}
}

// TestValueOf_DifferentDefaultValues ensures default values are returned correctly
func TestValueOf_DifferentDefaultValues(t *testing.T) {
	var nilPtr *string

	// Test with empty string default
	result1 := ValueOf(nilPtr, "")
	if result1 != "" {
		t.Errorf("Expected empty string default, got %q", result1)
	}

	// Test with non-empty string default
	result2 := ValueOf(nilPtr, "fallback")
	if result2 != "fallback" {
		t.Errorf("Expected 'fallback' default, got %q", result2)
	}
}

// Test Suite for PtrIf Function

// TestPtrIf_TrueCondition tests PtrIf when condition is true
func TestPtrIf_TrueCondition(t *testing.T) {
	value := "conditional value"
	result := PtrIf(true, value)

	if result == nil {
		t.Error("Expected PtrIf to return non-nil pointer when condition is true")
		return
	}

	if *result != value {
		t.Errorf("Expected dereferenced pointer to equal %q, got %q", value, *result)
	}
}

// TestPtrIf_FalseCondition tests PtrIf when condition is false
func TestPtrIf_FalseCondition(t *testing.T) {
	value := "should not be used"
	result := PtrIf(false, value)

	if result != nil {
		t.Error("Expected PtrIf to return nil when condition is false")
	}
}

// TestPtrIf_ConditionalLogic tests PtrIf with realistic conditional scenarios
func TestPtrIf_ConditionalLogic(t *testing.T) {
	// Test with non-empty string condition
	userName := "Alice"
	userPtr := PtrIf(userName != "", userName)

	if userPtr == nil {
		t.Error("Expected non-nil pointer for non-empty username")
		return
	}

	if *userPtr != userName {
		t.Errorf("Expected %q, got %q", userName, *userPtr)
	}

	// Test with empty string condition
	emptyName := ""
	emptyPtr := PtrIf(emptyName != "", emptyName)

	if emptyPtr != nil {
		t.Error("Expected nil pointer for empty username condition")
	}

	// Test with numeric condition
	score := 85
	scorePtr := PtrIf(score > 80, score)

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
	stringPtr := PtrOf("test")
	intPtr := PtrOf(42)
	boolPtr := PtrOf(true)

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
	stringPtr := PtrOf("test")
	intPtr := PtrOf(42)
	boolPtr := PtrOf(false)

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

// Test Suite for ZeroPtr Function

// TestZeroPtr_BasicTypes tests ZeroPtr with basic types
func TestZeroPtr_BasicTypes(t *testing.T) {
	// Test with int
	intZeroPtr := ZeroPtr[int]()
	if intZeroPtr == nil {
		t.Error("Expected ZeroPtr to return non-nil pointer for int")
		return
	}

	if *intZeroPtr != 0 {
		t.Errorf("Expected zero int value 0, got %d", *intZeroPtr)
	}

	// Test with string
	stringZeroPtr := ZeroPtr[string]()
	if stringZeroPtr == nil {
		t.Error("Expected ZeroPtr to return non-nil pointer for string")
		return
	}

	if *stringZeroPtr != "" {
		t.Errorf("Expected zero string value '', got %q", *stringZeroPtr)
	}

	// Test with bool
	boolZeroPtr := ZeroPtr[bool]()
	if boolZeroPtr == nil {
		t.Error("Expected ZeroPtr to return non-nil pointer for bool")
		return
	}

	if *boolZeroPtr != false {
		t.Errorf("Expected zero bool value false, got %t", *boolZeroPtr)
	}
}

// TestZeroPtr_PointerIndependence ensures each ZeroPtr call returns independent pointers
func TestZeroPtr_PointerIndependence(t *testing.T) {
	ptr1 := ZeroPtr[int]()
	ptr2 := ZeroPtr[int]()

	// Should be different pointer addresses
	if ptr1 == ptr2 {
		t.Error("Expected different pointer addresses for separate ZeroPtr calls")
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
		Email: PtrIf(userEmail != "", userEmail),
		Age:   PtrIf(userAge > 0, userAge),
	}

	// Verify the user was created correctly
	if user.Name != userName {
		t.Errorf("Expected name %q, got %q", userName, user.Name)
	}

	if IsNil(user.Email) {
		t.Error("Expected non-nil email pointer")
		return
	}

	if ValueOf(user.Email, "") != userEmail {
		t.Errorf("Expected email %q, got %q", userEmail, ValueOf(user.Email, ""))
	}

	if IsNil(user.Age) {
		t.Error("Expected non-nil age pointer")
		return
	}

	if ValueOf(user.Age, 0) != userAge {
		t.Errorf("Expected age %d, got %d", userAge, ValueOf(user.Age, 0))
	}
}

// TestEdgeCases_ComplexTypes tests with complex types like slices and maps
func TestEdgeCases_ComplexTypes(t *testing.T) {
	// Test with slice
	originalSlice := []int{1, 2, 3}
	slicePtr := PtrOf(originalSlice)

	if IsNil(slicePtr) {
		t.Error("Expected non-nil slice pointer")
		return
	}

	retrievedSlice := ValueOf(slicePtr, []int{})
	if len(retrievedSlice) != len(originalSlice) {
		t.Errorf("Expected slice length %d, got %d", len(originalSlice), len(retrievedSlice))
	}

	// Test with map
	originalMap := map[string]int{"key": 42}
	mapPtr := PtrOf(originalMap)

	if IsNil(mapPtr) {
		t.Error("Expected non-nil map pointer")
		return
	}

	retrievedMap := ValueOf(mapPtr, map[string]int{})
	if retrievedMap["key"] != 42 {
		t.Errorf("Expected map value 42, got %d", retrievedMap["key"])
	}
}

// TestEdgeCases_NilComparison tests nil comparison functions consistency
func TestEdgeCases_NilComparison(t *testing.T) {
	var nilPtr *string
	validPtr := PtrOf("test")

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
