package ptr

// To returns a pointer to v.
func To[T any](v T) *T {
	val := v
	return &val
}

// DerefOr returns the value pointed to by p, or defaultVal when p is nil.
func DerefOr[T any](p *T, defaultVal T) T {
	if p == nil {
		return defaultVal
	}
	return *p
}

// ToIf returns a pointer to v when cond is true, otherwise nil.
func ToIf[T any](cond bool, v T) *T {
	if cond {
		return To(v)
	}
	return nil
}

// Zero returns a pointer to the zero value of T.
func Zero[T any]() *T {
	return new(T)
}

// Deref returns the value pointed to by p, or the zero value of T when p is nil.
func Deref[T any](p *T) T {
	var zero T
	return DerefOr(p, zero)
}
