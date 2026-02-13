package container

// Difference returns elements in sliceA that are not present in sliceB.
// Returns nil if sliceA is nil, or a copy of sliceA if sliceB is nil/empty.
//
// Example:
//
//	Difference([]int{1, 2, 3}, []int{2}) // returns []int{1, 3}
func Difference[T comparable](sliceA, sliceB []T) []T {
	if sliceA == nil {
		return nil
	}

	if len(sliceB) == 0 {
		result := make([]T, len(sliceA))
		copy(result, sliceA)
		return result
	}

	lookup := make(map[T]struct{}, len(sliceB))
	for _, item := range sliceB {
		lookup[item] = struct{}{}
	}

	result := make([]T, 0, len(sliceA))
	for _, item := range sliceA {
		if _, found := lookup[item]; !found {
			result = append(result, item)
		}
	}

	return result
}

// Intersection returns elements common to both sliceA and sliceB, preserving uniqueness.
// Returns nil if either slice is nil, or an empty slice if there are no common elements.
//
// Example:
//
//	Intersection([]int{1, 2, 3}, []int{2, 3, 4}) // returns []int{2, 3}
func Intersection[T comparable](sliceA, sliceB []T) []T {
	if sliceA == nil || sliceB == nil {
		return nil
	}

	if len(sliceA) == 0 || len(sliceB) == 0 {
		return []T{}
	}

	lookup := make(map[T]struct{}, len(sliceB))
	for _, item := range sliceB {
		lookup[item] = struct{}{}
	}

	// Pre-allocate with smaller capacity estimate
	estimatedCap := min(len(sliceA), len(sliceB))
	result := make([]T, 0, estimatedCap)
	seen := make(map[T]struct{}, estimatedCap)

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

// Union returns all unique elements from sliceA and sliceB, preserving the order of first occurrence.
// Returns nil if both slices are nil.
//
// Example:
//
//	Union([]int{1, 2}, []int{2, 3}) // returns []int{1, 2, 3}
func Union[T comparable](sliceA, sliceB []T) []T {
	if sliceA == nil && sliceB == nil {
		return nil
	}

	totalLen := len(sliceA) + len(sliceB)
	result := make([]T, 0, totalLen)
	seen := make(map[T]struct{}, totalLen)

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

// Deduplicate returns unique elements from the input slice, preserving the order of first occurrence.
// Returns nil if input is nil, or an empty slice if input is empty (len=0 but not nil).
//
// Example:
//
//	Deduplicate([]int{1, 2, 2, 3}) // returns []int{1, 2, 3}
//	Deduplicate([]int(nil))        // returns nil
//	Deduplicate([]int{})           // returns []int{}
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

// ToMap returns a map of elements from the input slice, using keySelector to generate keys.
// If multiple items produce the same key, the last item wins.
// Returns an empty map (not nil) if input is nil or empty.
//
// Example:
//
//	users := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
//	userMap := ToMap(users, func(u User) int { return u.ID })
//	// userMap is map[int]User{1: {ID: 1, Name: "Alice"}, 2: {ID: 2, Name: "Bob"}}
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
// Returns nil if input is nil.
//
// Example:
//
//	FlatMap([]int{1, 2}, func(x int) []int { return []int{x, x*2} }) // returns []int{1, 2, 2, 4}
func FlatMap[T any, R any](input []T, mapper func(T) []R) []R {
	if input == nil {
		return nil
	}
	result := make([]R, 0)
	for _, item := range input {
		result = append(result, mapper(item)...)
	}
	return result
}

// Reduce reduces the slice to a single value using the reducer function.
// Returns zero value if input is nil or empty.
//
// Example:
//
//	Reduce([]int{1, 2, 3}, 0, func(sum, x int) int { return sum + x }) // returns 6
func Reduce[T any, R any](input []T, initial R, reducer func(R, T) R) R {
	result := initial
	for _, item := range input {
		result = reducer(result, item)
	}
	return result
}

// Chunk returns a slice of chunks, each containing up to size elements.
// Returns nil if input is nil, or empty slice if size <= 0.
//
// Example:
//
//	Chunk([]int{1, 2, 3, 4, 5}, 2) // returns [][]int{{1, 2}, {3, 4}, {5}}
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

// First returns the first element matching predicate, or zero value if not found.
// The found parameter indicates whether a matching element was found.
//
// Example:
//
//	First([]int{1, 2, 3}, func(x int) bool { return x > 1 }) // returns 2, true
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
//
// Example:
//
//	Partition([]int{1, 2, 3, 4}, func(x int) bool { return x%2 == 0 }) // returns []int{2, 4}, []int{1, 3}
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
