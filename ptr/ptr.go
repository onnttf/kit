package ptr

// PtrOf returns a pointer to a copy of the provided value, avoiding temporary references
func PtrOf[T any](v T) *T {
	val := v // Create a copy to ensure a stable address
	return &val
}

// ValueOf returns the value pointed to by ptr, or defaultVal if ptr is nil
func ValueOf[T any](ptr *T, defaultVal T) T {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// PtrIf returns a pointer to a copy of v if cond is true, otherwise nil
func PtrIf[T any](cond bool, v T) *T {
	if cond {
		return PtrOf(v)
	}
	return nil
}

// IsNil returns true if the provided pointer is nil
func IsNil[T any](ptr *T) bool {
	return ptr == nil
}

// IsNotNil returns true if the provided pointer is not nil
func IsNotNil[T any](ptr *T) bool {
	return ptr != nil
}

// ZeroPtr returns a pointer to the zero value of type T
func ZeroPtr[T any]() *T {
	return new(T)
}
