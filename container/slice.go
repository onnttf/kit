package container

// Difference returns elements in sliceA that are not present in sliceB
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

// Intersection returns elements common to both sliceA and sliceB, preserving uniqueness
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

// Union returns all unique elements from sliceA and sliceB, preserving the order of first occurrence
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

// Deduplicate returns unique elements from the input slice, preserving the order of first occurrence
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

// ToMap returns a map of elements from the input slice, using keySelector to generate keys, retaining the last value for duplicate keys
func ToMap[T any, K comparable](input []T, keySelector func(T) K) map[K]T {
	result := make(map[K]T, len(input))

	for _, item := range input {
		key := keySelector(item)
		result[key] = item
	}

	return result
}
