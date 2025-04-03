package container

// Deduplicate removes duplicate elements from a slice.
func Deduplicate[T comparable](input []T) []T {
	if len(input) == 0 {
		return nil
	}

	seen := make(map[T]struct{}, len(input))
	result := make([]T, 0, len(input))

	for _, v := range input {
		if _, exists := seen[v]; !exists {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}

	return result
}

// ToMap converts a slice to a map using the provided keySelector function.
func ToMap[T any, K comparable](input []T, keySelector func(T) K) map[K]T {
	if len(input) == 0 {
		return make(map[K]T)
	}

	result := make(map[K]T, len(input))

	for _, v := range input {
		key := keySelector(v)
		result[key] = v
	}

	return result
}
