package ptr

func To[T any](v T) *T {
	val := v
	return &val
}

func DerefOr[T any](p *T, defaultVal T) T {
	if p == nil {
		return defaultVal
	}
	return *p
}

func ToIf[T any](cond bool, v T) *T {
	if cond {
		return To(v)
	}
	return nil
}

func Zero[T any]() *T {
	return new(T)
}

func Deref[T any](p *T) T {
	var zero T
	return DerefOr(p, zero)
}
