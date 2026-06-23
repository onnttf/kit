package container

import (
	"errors"
	"slices"
)

var ErrNilCallback = errors.New("container: callback is nil")

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

func ValidValues[T comparable](values []T, keep ...func(T) bool) []T {
	if values == nil {
		return nil
	}

	var zero T
	out := make([]T, 0, len(values))
	for _, value := range values {
		if value == zero {
			continue
		}

		valid := true
		for _, predicate := range keep {
			if predicate != nil && !predicate(value) {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}

		out = append(out, value)
	}

	return out
}

func Deduplicate[T comparable](values []T) []T {
	if values == nil {
		return nil
	}

	if len(values) == 0 {
		return []T{}
	}

	out := make([]T, 0, len(values))
	seen := make(map[T]struct{}, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func ToMap[T any, K comparable](input []T, keySelector func(T) K) (map[K]T, error) {
	if keySelector == nil {
		return nil, ErrNilCallback
	}

	if len(input) == 0 {
		return make(map[K]T), nil
	}

	result := make(map[K]T, len(input))
	for _, item := range input {
		key := keySelector(item)
		result[key] = item
	}

	return result, nil
}

func FlatMap[T any, R any](input []T, mapper func(T) []R) ([]R, error) {
	if mapper == nil {
		return nil, ErrNilCallback
	}

	if input == nil {
		return nil, nil
	}
	result := make([]R, 0, len(input)*2)
	for _, item := range input {
		result = append(result, mapper(item)...)
	}
	return result, nil
}

func Reduce[T any, R any](input []T, initial R, reducer func(R, T) R) (R, error) {
	if reducer == nil {
		return initial, ErrNilCallback
	}

	result := initial
	for _, item := range input {
		result = reducer(result, item)
	}
	return result, nil
}

func First[T any](input []T, predicate func(T) bool) (T, bool, error) {
	if predicate == nil {
		var zero T
		return zero, false, ErrNilCallback
	}

	for _, item := range input {
		if predicate(item) {
			return item, true, nil
		}
	}
	var zero T
	return zero, false, nil
}

func Partition[T any](input []T, predicate func(T) bool) (matches []T, nonMatches []T, err error) {
	if predicate == nil {
		return nil, nil, ErrNilCallback
	}

	matches = make([]T, 0)
	nonMatches = make([]T, 0)
	for _, item := range input {
		if predicate(item) {
			matches = append(matches, item)
		} else {
			nonMatches = append(nonMatches, item)
		}
	}
	return matches, nonMatches, nil
}

func GroupBy[T any, K comparable](input []T, keyFunc func(T) K) (map[K][]T, error) {
	if keyFunc == nil {
		return nil, ErrNilCallback
	}

	result := make(map[K][]T, len(input))
	for _, item := range input {
		key := keyFunc(item)
		result[key] = append(result[key], item)
	}
	return result, nil
}
