package container

import "slices"

// Difference returns the elements in s1 that are not present in s2.
// The order from s1 is preserved.
func Difference[T comparable](s1, s2 []T) []T {
	if s1 == nil {
		return nil
	}

	if len(s2) == 0 {
		return slices.Clone(s1)
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

// Intersection returns the unique elements that appear in both slices.
// The order follows their first occurrence in s1.
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

// Union returns unique elements from s1 followed by unique elements from s2.
// The order of first occurrence is preserved.
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

// Deduplicate returns the unique elements from input in first-seen order.
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

// ToMap returns a map keyed by keySelector. Later items overwrite earlier ones
// when the selector returns duplicate keys.
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

// FlatMap applies mapper to each input element and concatenates the returned slices.
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

// Reduce folds input into a single value, starting with initial.
func Reduce[T any, R any](input []T, initial R, reducer func(R, T) R) R {
	result := initial
	for _, item := range input {
		result = reducer(result, item)
	}
	return result
}

// First returns the first item matching predicate.
func First[T any](input []T, predicate func(T) bool) (T, bool) {
	for _, item := range input {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// Partition splits input into matching and non-matching elements.
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

// GroupBy groups input elements by the key returned from keyFunc.
func GroupBy[T any, K comparable](input []T, keyFunc func(T) K) map[K][]T {
	result := make(map[K][]T, len(input))
	for _, item := range input {
		key := keyFunc(item)
		result[key] = append(result[key], item)
	}
	return result
}
