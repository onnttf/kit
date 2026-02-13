package ptr

// To returns a pointer to a copy of v.
//
// Example:
//
//	x := 42
//	p := To(x) // p points to 42
func To[T any](v T) *T {
	val := v
	return &val
}

// DerefOr returns the value pointed to by p, or defaultVal if p is nil.
//
// Example:
//
//	p := To(42)
//	DerefOr(p, 0)   // returns 42
//	DerefOr(nil, 100) // returns 100
func DerefOr[T any](p *T, defaultVal T) T {
	if p == nil {
		return defaultVal
	}
	return *p
}

// ToIf returns a pointer to a copy of v if cond is true, otherwise nil.
//
// Example:
//
//	ToIf(true, 42)   // returns *int pointing to 42
//	ToIf(false, 42)  // returns nil
func ToIf[T any](cond bool, v T) *T {
	if cond {
		return To(v)
	}
	return nil
}

// IsNil reports whether p is nil.
//
// Example:
//
//	var p *int
//	IsNil(p)     // returns true
//	IsNil(To(42)) // returns false
func IsNil[T any](p *T) bool {
	return p == nil
}

// IsNotNil reports whether p is not nil.
//
// Example:
//
//	var p *int
//	IsNotNil(p)     // returns false
//	IsNotNil(To(42)) // returns true
func IsNotNil[T any](p *T) bool {
	return p != nil
}

// Zero returns a pointer to the zero value of type T.
//
// Example:
//
//	p := Zero[int]() // returns *int pointing to 0
func Zero[T any]() *T {
	return new(T)
}

// Deref returns the value pointed to by p, or the zero value if p is nil.
//
// Example:
//
//	p := To(42)
//	Deref(p)   // returns 42
//	Deref(nil) // returns 0
func Deref[T any](p *T) T {
	var zero T
	return DerefOr(p, zero)
}
