package container

// Difference returns the elements in sliceA that are not in sliceB.
func Difference[T comparable](sliceA, sliceB []T) []T {
	result := make([]T, 0)
	lookup := make(map[T]struct{}, len(sliceB))

	for _, item := range sliceB {
		lookup[item] = struct{}{}
	}

	for _, item := range sliceA {
		if _, found := lookup[item]; !found {
			result = append(result, item)
		}
	}

	return result
}

// Intersection returns the elements present in both sliceA and sliceB.
func Intersection[T comparable](sliceA, sliceB []T) []T {
	result := make([]T, 0)
	lookup := make(map[T]struct{}, len(sliceB))

	for _, item := range sliceB {
		lookup[item] = struct{}{}
	}

	seen := make(map[T]struct{})
	for _, item := range sliceA {
		if _, found := lookup[item]; found {
			if _, added := seen[item]; !added {
				result = append(result, item)
				seen[item] = struct{}{}
			}
		}
	}

	return result
}

// Union returns all distinct elements from sliceA and sliceB.
func Union[T comparable](sliceA, sliceB []T) []T {
	result := make([]T, 0)
	seen := make(map[T]struct{}, len(sliceA)+len(sliceB))

	for _, item := range sliceA {
		if _, exists := seen[item]; !exists {
			result = append(result, item)
			seen[item] = struct{}{}
		}
	}

	for _, item := range sliceB {
		if _, exists := seen[item]; !exists {
			result = append(result, item)
			seen[item] = struct{}{}
		}
	}

	return result
}

// Deduplicate returns a slice containing unique elements from input.
// It preserves the order of the first occurrence.
func Deduplicate[T comparable](input []T) []T {
	if len(input) == 0 {
		return nil
	}

	seen := make(map[T]struct{}, len(input))
	result := make([]T, 0, len(input))

	for _, item := range input {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// ToMap returns a map constructed from input using keySelector to generate keys.
// If duplicate keys are produced, the last value is retained.
func ToMap[T any, K comparable](input []T, keySelector func(T) K) map[K]T {
	result := make(map[K]T, len(input))

	for _, item := range input {
		key := keySelector(item)
		result[key] = item
	}

	return result
}
