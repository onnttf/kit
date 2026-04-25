package container

// Difference returns the elements in s1 that are not in s2.
// If s1 is nil, it returns nil. If s2 is nil or empty, it returns a copy of s1.
func Difference[T comparable](s1, s2 []T) []T {
	if s1 == nil {
		return nil
	}

	if len(s2) == 0 {
		result := make([]T, len(s1))
		copy(result, s1)
		return result
	}

	lookup := make(map[T]struct{}, len(s2))
	for _, item := range s2 {
		lookup[item] = struct{}{}
	}

	result := make([]T, 0, len(s1))
	for _, item := range s1 {
		if _, found := lookup[item]; !found {
			result = append(result, item)
		}
	}

	return result
}

// Intersection returns the unique elements common to both s1 and s2.
// If either slice is nil, it returns nil. If there are no common elements, it returns an empty slice.
func Intersection[T comparable](s1, s2 []T) []T {
	if s1 == nil || s2 == nil {
		return nil
	}

	if len(s1) == 0 || len(s2) == 0 {
		return []T{}
	}

	lookup := make(map[T]struct{}, len(s2))
	for _, item := range s2 {
		lookup[item] = struct{}{}
	}

	estimatedCap := min(len(s1), len(s2))
	result := make([]T, 0, estimatedCap)
	seen := make(map[T]struct{}, estimatedCap)

	for _, item := range s1 {
		if _, found := lookup[item]; found {
			if _, added := seen[item]; !added {
				result = append(result, item)
				seen[item] = struct{}{}
			}
		}
	}

	return result
}

// Union returns the unique elements from s1 and s2.
// The order of first occurrence is preserved. If both slices are nil, it returns nil.
func Union[T comparable](s1, s2 []T) []T {
	if s1 == nil && s2 == nil {
		return nil
	}

	totalLen := len(s1) + len(s2)
	result := make([]T, 0, totalLen)
	seen := make(map[T]struct{}, totalLen)

	for _, item := range s1 {
		if _, exists := seen[item]; !exists {
			result = append(result, item)
			seen[item] = struct{}{}
		}
	}

	for _, item := range s2 {
		if _, exists := seen[item]; !exists {
			result = append(result, item)
			seen[item] = struct{}{}
		}
	}

	return result
}

// Deduplicate returns unique elements from the input slice.
// The order of first occurrence is preserved. If input is nil, it returns nil.
func Deduplicate[T comparable](input []T) []T {
	if input == nil {
		return nil
	}

	if len(input) == 0 {
		return []T{}
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

// ToMap returns a map from the input slice using keySelector to generate keys.
// If multiple items produce the same key, the last item wins.
// If input is nil or empty, it returns an empty map.
func ToMap[T any, K comparable](input []T, keySelector func(T) K) map[K]T {
	if len(input) == 0 {
		return make(map[K]T)
	}

	result := make(map[K]T, len(input))
	for _, item := range input {
		key := keySelector(item)
		result[key] = item
	}

	return result
}

// FlatMap returns a new slice by applying mapper to each element and flattening the results.
// If input is nil, it returns nil.
func FlatMap[T any, R any](input []T, mapper func(T) []R) []R {
	if input == nil {
		return nil
	}
	result := make([]R, 0, len(input)*2)
	for _, item := range input {
		result = append(result, mapper(item)...)
	}
	return result
}

// Reduce reduces the slice to a single value using the reducer function.
// It returns the initial value if input is nil or empty.
func Reduce[T any, R any](input []T, initial R, reducer func(R, T) R) R {
	result := initial
	for _, item := range input {
		result = reducer(result, item)
	}
	return result
}

// Chunk returns a slice of chunks, each containing up to size elements.
// If input is nil, it returns nil. If size is not positive, it returns an empty slice.
func Chunk[T any](input []T, size int) [][]T {
	if input == nil {
		return nil
	}
	if size <= 0 {
		return [][]T{}
	}
	result := make([][]T, 0, (len(input)+size-1)/size)
	for i := 0; i < len(input); i += size {
		end := min(i+size, len(input))
		result = append(result, input[i:end])
	}
	return result
}

// First returns the first element matching predicate and true.
// If no match is found, it returns the zero value and false.
func First[T any](input []T, predicate func(T) bool) (T, bool) {
	for _, item := range input {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// Partition splits the slice into two groups based on predicate.
// The first slice contains elements for which predicate is true, the second contains the rest.
func Partition[T any](input []T, predicate func(T) bool) (matches []T, nonMatches []T) {
	matches = make([]T, 0)
	nonMatches = make([]T, 0)
	for _, item := range input {
		if predicate(item) {
			matches = append(matches, item)
		} else {
			nonMatches = append(nonMatches, item)
		}
	}
	return
}

// GroupBy groups elements by a key extracted by keyFunc.
// It returns a map from keys to slices of elements that share that key.
func GroupBy[T any, K comparable](input []T, keyFunc func(T) K) map[K][]T {
	result := make(map[K][]T, len(input))
	for _, item := range input {
		key := keyFunc(item)
		result[key] = append(result[key], item)
	}
	return result
}
