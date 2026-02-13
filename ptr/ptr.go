package ptr

// PtrOf returns a pointer to a copy of v.
//
// Example:
//
//	x := 42
//	p := PtrOf(x) // p points to 42
func PtrOf[T any](v T) *T {
	val := v
	return &val
}

// ValueOf returns the value pointed to by ptr, or defaultVal if ptr is nil.
//
// Example:
//
//	p := PtrOf(42)
//	ValueOf(p, 0)    // returns 42
//	ValueOf(nil, 100) // returns 100
func ValueOf[T any](ptr *T, defaultVal T) T {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// PtrIf returns a pointer to a copy of v if cond is true, otherwise nil.
//
// Example:
//
//	PtrIf(true, 42)   // returns *int pointing to 42
//	PtrIf(false, 42)  // returns nil
func PtrIf[T any](cond bool, v T) *T {
	if cond {
		return PtrOf(v)
	}
	return nil
}

// IsNil reports whether ptr is nil.
//
// Example:
//
//	var p *int
//	IsNil(p)       // returns true
//	IsNil(PtrOf(42)) // returns false
func IsNil[T any](ptr *T) bool {
	return ptr == nil
}

// IsNotNil reports whether ptr is not nil.
//
// Example:
//
//	var p *int
//	IsNotNil(p)       // returns false
//	IsNotNil(PtrOf(42)) // returns true
func IsNotNil[T any](ptr *T) bool {
	return ptr != nil
}

// ZeroPtr returns a pointer to the zero value of type T.
//
// Example:
//
//	p := ZeroPtr[int]() // returns *int pointing to 0
func ZeroPtr[T any]() *T {
	return new(T)
}

// Deref returns the value pointed to by ptr, or the zero value if ptr is nil.
//
// Example:
//
//	p := PtrOf(42)
//	Deref(p)   // returns 42
//	Deref(nil) // returns 0
func Deref[T any](ptr *T) T {
	var zero T
	return ValueOf(ptr, zero)
}
