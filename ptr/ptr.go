package ptr

// To returns a pointer to a copy of v.
//
// Example:
//
//	p := To(42)
func To[T any](v T) *T {
	val := v
	return &val
}

// DerefOr returns the value pointed to by p, or defaultVal if p is nil.
//
// Example:
//
//	DerefOr(To(42), 0)
//	DerefOr(nil, 100)
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
//	ToIf(true, 42)
//	ToIf(false, 42)
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
//	IsNil((*int)(nil))
//	IsNil(To(42))
func IsNil[T any](p *T) bool {
	return p == nil
}

// IsNotNil reports whether p is not nil.
//
// Example:
//
//	IsNotNil((*int)(nil))
//	IsNotNil(To(42))
func IsNotNil[T any](p *T) bool {
	return p != nil
}

// Zero returns a pointer to the zero value of type T.
//
// Example:
//
//	p := Zero[int]()
func Zero[T any]() *T {
	return new(T)
}

// Deref returns the value pointed to by p, or the zero value if p is nil.
//
// Example:
//
//	Deref(To(42))
//	Deref[int](nil)
func Deref[T any](p *T) T {
	var zero T
	return DerefOr(p, zero)
}
