package slicex

import "errors"

var ErrNilCallback = errors.New("slicex: callback is nil")

func Unique[S ~[]E, E comparable](s S) S {
	seen := make(map[E]struct{}, len(s))
	out := make(S, 0, len(s))
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func Diff[S ~[]E, E comparable](a, b S) S {
	exclude := make(map[E]struct{}, len(b))
	for _, v := range b {
		exclude[v] = struct{}{}
	}

	out := make(S, 0, len(a))
	for _, v := range a {
		if _, ok := exclude[v]; !ok {
			out = append(out, v)
		}
	}
	return out
}

func Intersect[S ~[]E, E comparable](a, b S) S {
	inB := make(map[E]struct{}, len(b))
	for _, v := range b {
		inB[v] = struct{}{}
	}

	seen := make(map[E]struct{})
	out := make(S, 0, min(len(a), len(b)))
	for _, v := range a {
		if _, ok := inB[v]; !ok {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func Union[S ~[]E, E comparable](ss ...S) S {
	var total int
	for _, s := range ss {
		total += len(s)
	}

	seen := make(map[E]struct{}, total)
	out := make(S, 0, total)
	for _, s := range ss {
		for _, v := range s {
			if _, ok := seen[v]; ok {
				continue
			}
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}

func GroupBy[T any, K comparable](s []T, key func(T) K) (map[K][]T, error) {
	if key == nil {
		return nil, ErrNilCallback
	}
	out := make(map[K][]T)
	for _, v := range s {
		k := key(v)
		out[k] = append(out[k], v)
	}
	return out, nil
}

func Partition[S ~[]E, E any](s S, keep func(E) bool) (yes S, no S, err error) {
	if keep == nil {
		return nil, nil, ErrNilCallback
	}
	yes = make(S, 0, len(s))
	no = make(S, 0, len(s))
	for _, v := range s {
		if keep(v) {
			yes = append(yes, v)
		} else {
			no = append(no, v)
		}
	}
	return yes, no, nil
}

func ToMap[T any, K comparable](s []T, key func(T) K) (map[K]T, error) {
	if key == nil {
		return nil, ErrNilCallback
	}
	out := make(map[K]T, len(s))
	for _, v := range s {
		out[key(v)] = v
	}
	return out, nil
}

func First[T any](s []T, keep func(T) bool) (T, bool, error) {
	if keep == nil {
		var zero T
		return zero, false, ErrNilCallback
	}
	for _, v := range s {
		if keep(v) {
			return v, true, nil
		}
	}
	var zero T
	return zero, false, nil
}
