package ptr

// PtrOf returns a pointer to the provided value.
// It creates a safe copy of the input value to ensure the returned pointer
// does not reference a temporary or local variable, which is a common pitfall in Go.
//
// Example:
//
//	s := PtrOf("hello") // s is of type *string, points to "hello"
//	i := PtrOf(123)     // i is of type *int, points to 123
func PtrOf[T any](v T) *T {
	val := v // Create a copy to get a stable address.
	return &val
}

// ValueOf returns the value that 'ptr' points to.
// If 'ptr' is nil, it returns 'defaultVal' instead, providing a safe way
// to dereference pointers that might be unset.
//
// Example:
//
//	var strPtr *string // strPtr is nil
//	val := ValueOf(strPtr, "default") // val will be "default"
//
//	numPtr := PtrOf(100)
//	num := ValueOf(numPtr, 0) // num will be 100
func ValueOf[T any](ptr *T, defaultVal T) T {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// PtrIf returns a pointer to 'v' if 'cond' is true.
// Otherwise, it returns nil. This is useful for conditionally populating
// optional fields in structs or API requests.
//
// Example:
//
//	userName := "Alice"
//	optionalUser := PtrIf(userName != "", userName) // optionalUser will be *string("Alice")
//
//	emptyName := ""
//	noUser := PtrIf(emptyName != "", emptyName) // noUser will be nil
func PtrIf[T any](cond bool, v T) *T {
	if cond {
		return PtrOf(v)
	}
	return nil
}

// IsNil checks whether the given pointer is nil.
// This function works universally for any pointer type.
//
// Example:
//
//	var p *int
//	IsNil(p) // Returns true
func IsNil[T any](ptr *T) bool {
	return ptr == nil
}

// IsNotNil checks whether the given pointer is not nil.
// This function works universally for any pointer type.
//
// Example:
//
//	p := PtrOf(5)
//	IsNotNil(p) // Returns true
func IsNotNil[T any](ptr *T) bool {
	return ptr != nil
}

// ZeroPtr returns a pointer to the zero value of type T.
// This is functionally equivalent to using the built-in new(T).
//
// Example:
//
//	intZeroPtr := ZeroPtr[int]()       // *int pointing to 0
//	stringZeroPtr := ZeroPtr[string]() // *string pointing to ""
func ZeroPtr[T any]() *T {
	return new(T)
}
