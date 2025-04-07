package container

// Deduplicate returns a new slice containing only the unique elements from input.
// The order of elements in the result preserves the first occurrence in the input.
// Returns nil if the input slice is empty.
func Deduplicate[T comparable](input []T) []T {
	if len(input) == 0 {
		return nil
	}

	uniqueItems := make(map[T]struct{}, len(input))
	uniqueResult := make([]T, 0, len(input))

	for _, item := range input {
		// Only add each item once to the result
		if _, exists := uniqueItems[item]; !exists {
			uniqueItems[item] = struct{}{}
			uniqueResult = append(uniqueResult, item)
		}
	}

	return uniqueResult
}

// ToMap transforms a slice into a map where each value from the input
// is stored with a key determined by the keySelector function.
// If multiple values produce the same key, later values will overwrite earlier ones.
// Returns an empty map if the input slice is empty.
func ToMap[T any, K comparable](input []T, keySelector func(T) K) map[K]T {
	if len(input) == 0 {
		return make(map[K]T)
	}

	mappedResult := make(map[K]T, len(input))

	for _, item := range input {
		key := keySelector(item)
		mappedResult[key] = item
	}

	return mappedResult
}
