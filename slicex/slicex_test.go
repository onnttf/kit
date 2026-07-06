package slicex

import (
	"errors"
	"reflect"
	"testing"
)

func TestSetOperations(t *testing.T) {
	if got := Unique([]int{1, 2, 1, 3, 2}); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("Unique() = %v", got)
	}
	if got := Diff([]int{1, 2, 3}, []int{2}); !reflect.DeepEqual(got, []int{1, 3}) {
		t.Fatalf("Diff() = %v", got)
	}
	if got := Intersect([]int{1, 2, 2, 3}, []int{2, 3}); !reflect.DeepEqual(got, []int{2, 3}) {
		t.Fatalf("Intersect() = %v", got)
	}
	if got := Union([]int{1, 2}, []int{2, 3}); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("Union() = %v", got)
	}
}

func TestCallbacks(t *testing.T) {
	grouped, err := GroupBy([]string{"a", "bb"}, func(s string) int { return len(s) })
	if err != nil {
		t.Fatalf("GroupBy() error = %v", err)
	}
	if !reflect.DeepEqual(grouped, map[int][]string{1: {"a"}, 2: {"bb"}}) {
		t.Fatalf("GroupBy() = %v", grouped)
	}

	yes, no, err := Partition([]int{1, 2, 3}, func(v int) bool { return v%2 == 1 })
	if err != nil {
		t.Fatalf("Partition() error = %v", err)
	}
	if !reflect.DeepEqual(yes, []int{1, 3}) || !reflect.DeepEqual(no, []int{2}) {
		t.Fatalf("Partition() = %v, %v", yes, no)
	}

	m, err := ToMap([]string{"a", "bb"}, func(s string) int { return len(s) })
	if err != nil {
		t.Fatalf("ToMap() error = %v", err)
	}
	if !reflect.DeepEqual(m, map[int]string{1: "a", 2: "bb"}) {
		t.Fatalf("ToMap() = %v", m)
	}

	v, ok, err := First([]int{1, 2, 3}, func(v int) bool { return v > 1 })
	if err != nil || !ok || v != 2 {
		t.Fatalf("First() = %d, %v, %v", v, ok, err)
	}
}

func TestNilCallbacks(t *testing.T) {
	if _, err := GroupBy[int, int](nil, nil); !errors.Is(err, ErrNilCallback) {
		t.Fatalf("GroupBy(nil) error = %v", err)
	}
	if _, _, err := Partition([]int{1}, nil); !errors.Is(err, ErrNilCallback) {
		t.Fatalf("Partition(nil) error = %v", err)
	}
	if _, err := ToMap[int, int](nil, nil); !errors.Is(err, ErrNilCallback) {
		t.Fatalf("ToMap(nil) error = %v", err)
	}
	if _, _, err := First([]int{1}, nil); !errors.Is(err, ErrNilCallback) {
		t.Fatalf("First(nil) error = %v", err)
	}
}

type namedInts []int

func TestSetOperationsPreserveNamedSliceTypeAndHandleEmpty(t *testing.T) {
	var empty namedInts
	if got := Unique(empty); got == nil || len(got) != 0 {
		t.Fatalf("Unique(empty) = %#v", got)
	}
	if got := Diff(namedInts{1, 2}, namedInts{1, 2}); got == nil || len(got) != 0 {
		t.Fatalf("Diff(all removed) = %#v", got)
	}
	if got := Intersect(namedInts{1, 2}, nil); got == nil || len(got) != 0 {
		t.Fatalf("Intersect(no match) = %#v", got)
	}
	if got := Union[namedInts](); got == nil || len(got) != 0 {
		t.Fatalf("Union(no args) = %#v", got)
	}
}

func TestGroupByToMapDuplicateKeysAndFirstNotFound(t *testing.T) {
	grouped, err := GroupBy([]string{"a", "b", "cc"}, func(s string) int { return len(s) })
	if err != nil {
		t.Fatalf("GroupBy() error = %v", err)
	}
	if !reflect.DeepEqual(grouped, map[int][]string{1: {"a", "b"}, 2: {"cc"}}) {
		t.Fatalf("GroupBy() = %v", grouped)
	}

	m, err := ToMap([]string{"a", "b"}, func(string) int { return 1 })
	if err != nil {
		t.Fatalf("ToMap() error = %v", err)
	}
	if !reflect.DeepEqual(m, map[int]string{1: "b"}) {
		t.Fatalf("ToMap(duplicate keys) = %v", m)
	}

	v, ok, err := First([]int{1, 2, 3}, func(v int) bool { return v > 9 })
	if err != nil || ok || v != 0 {
		t.Fatalf("First(not found) = %d, %v, %v", v, ok, err)
	}
}

func FuzzSetOperationsInvariants(f *testing.F) {
	f.Add([]byte{1, 2, 1, 3}, []byte{2, 4})
	f.Add([]byte{}, []byte{1})
	f.Add([]byte{5, 5, 5}, []byte{5})

	f.Fuzz(func(t *testing.T, a, b []byte) {
		unique := Unique(a)
		seen := make(map[byte]bool)
		for _, v := range unique {
			if seen[v] {
				t.Fatalf("Unique(%v) contains duplicate %v: %v", a, v, unique)
			}
			seen[v] = true
		}

		diff := Diff(a, b)
		inB := make(map[byte]bool)
		for _, v := range b {
			inB[v] = true
		}
		for _, v := range diff {
			if inB[v] {
				t.Fatalf("Diff(%v, %v) contains excluded value %v: %v", a, b, v, diff)
			}
		}

		intersect := Intersect(a, b)
		seen = make(map[byte]bool)
		for _, v := range intersect {
			if !inB[v] {
				t.Fatalf("Intersect(%v, %v) contains value not in b: %v", a, b, v)
			}
			if seen[v] {
				t.Fatalf("Intersect(%v, %v) contains duplicate %v: %v", a, b, v, intersect)
			}
			seen[v] = true
		}
	})
}

func TestNilCallbacksErrorsAreSentinel(t *testing.T) {
	_, err := GroupBy[int, int]([]int{1}, nil)
	if !errors.Is(err, ErrNilCallback) {
		t.Fatalf("GroupBy nil error = %v", err)
	}
}
