package ptr

// To returns a pointer to a copy of v.
func To[T any](v T) *T {
	val := v
	return &val
}

// DerefOr returns the value pointed to by p, or defaultVal if p is nil.
func DerefOr[T any](p *T, defaultVal T) T {
	if p == nil {
		return defaultVal
	}
	return *p
}

// ToIf returns a pointer to a copy of v if cond is true, otherwise nil.
func ToIf[T any](cond bool, v T) *T {
	if cond {
		return To(v)
	}
	return nil
}

// IsNil reports whether p is nil.
func IsNil[T any](p *T) bool {
	return p == nil
}

// IsNotNil reports whether p is not nil.
func IsNotNil[T any](p *T) bool {
	return p != nil
}

// Zero returns a pointer to the zero value of type T.
func Zero[T any]() *T {
	return new(T)
}

// Deref returns the value pointed to by p, or the zero value if p is nil.
func Deref[T any](p *T) T {
	var zero T
	return DerefOr(p, zero)
}
