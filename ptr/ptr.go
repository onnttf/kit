package ptr

func To[T any](v T) *T {
	val := v
	return &val
}

func Value[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

func Or[T any](p *T, fallback T) T {
	if p == nil {
		return fallback
	}
	return *p
}
