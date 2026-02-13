package container

import (
	"reflect"
	"testing"
)

// Test Suite for Difference Function

func TestDifference_Basic(t *testing.T) {
	tests := []struct {
		name   string
		sliceA []int
		sliceB []int
		want   []int
	}{
		{
			name:   "basic difference",
			sliceA: []int{1, 2, 3, 4},
			sliceB: []int{2, 4},
			want:   []int{1, 3},
		},
		{
			name:   "no common elements",
			sliceA: []int{1, 2, 3},
			sliceB: []int{4, 5, 6},
			want:   []int{1, 2, 3},
		},
		{
			name:   "all elements common",
			sliceA: []int{1, 2, 3},
			sliceB: []int{1, 2, 3},
			want:   []int{},
		},
		{
			name:   "empty sliceA",
			sliceA: []int{},
			sliceB: []int{1, 2},
			want:   []int{},
		},
		{
			name:   "empty sliceB",
			sliceA: []int{1, 2, 3},
			sliceB: []int{},
			want:   []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Difference(tt.sliceA, tt.sliceB)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Difference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDifference_NilHandling(t *testing.T) {
	t.Run("nil sliceA returns nil", func(t *testing.T) {
		var sliceA []int
		sliceB := []int{1, 2}
		got := Difference(sliceA, sliceB)
		if got != nil {
			t.Errorf("Expected nil, got %v", got)
		}
	})

	t.Run("nil sliceB returns copy of sliceA", func(t *testing.T) {
		sliceA := []int{1, 2, 3}
		var sliceB []int
		got := Difference(sliceA, sliceB)
		if !reflect.DeepEqual(got, sliceA) {
			t.Errorf("Expected %v, got %v", sliceA, got)
		}
		// Verify it's a copy, not the same slice
		if len(got) > 0 && &got[0] == &sliceA[0] {
			t.Error("Expected a copy, got same underlying array")
		}
	})
}

func TestDifference_WithDuplicates(t *testing.T) {
	sliceA := []int{1, 2, 2, 3, 3, 3}
	sliceB := []int{2, 3}
	got := Difference(sliceA, sliceB)
	want := []int{1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Difference() = %v, want %v", got, want)
	}
}

// Test Suite for Intersection Function

func TestIntersection_Basic(t *testing.T) {
	tests := []struct {
		name   string
		sliceA []int
		sliceB []int
		want   []int
	}{
		{
			name:   "basic intersection",
			sliceA: []int{1, 2, 3, 4},
			sliceB: []int{2, 3, 5},
			want:   []int{2, 3},
		},
		{
			name:   "no common elements",
			sliceA: []int{1, 2, 3},
			sliceB: []int{4, 5, 6},
			want:   []int{},
		},
		{
			name:   "all elements common",
			sliceA: []int{1, 2, 3},
			sliceB: []int{1, 2, 3},
			want:   []int{1, 2, 3},
		},
		{
			name:   "empty sliceA",
			sliceA: []int{},
			sliceB: []int{1, 2},
			want:   []int{},
		},
		{
			name:   "empty sliceB",
			sliceA: []int{1, 2, 3},
			sliceB: []int{},
			want:   []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Intersection(tt.sliceA, tt.sliceB)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Intersection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntersection_NilHandling(t *testing.T) {
	t.Run("nil sliceA returns nil", func(t *testing.T) {
		var sliceA []int
		sliceB := []int{1, 2}
		got := Intersection(sliceA, sliceB)
		if got != nil {
			t.Errorf("Expected nil, got %v", got)
		}
	})

	t.Run("nil sliceB returns nil", func(t *testing.T) {
		sliceA := []int{1, 2, 3}
		var sliceB []int
		got := Intersection(sliceA, sliceB)
		if got != nil {
			t.Errorf("Expected nil, got %v", got)
		}
	})
}

func TestIntersection_WithDuplicates(t *testing.T) {
	sliceA := []int{1, 2, 2, 3, 3, 3}
	sliceB := []int{2, 2, 3}
	got := Intersection(sliceA, sliceB)
	want := []int{2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Intersection() = %v, want %v", got, want)
	}
}

// Test Suite for Union Function

func TestUnion_Basic(t *testing.T) {
	tests := []struct {
		name   string
		sliceA []int
		sliceB []int
		want   []int
	}{
		{
			name:   "basic union",
			sliceA: []int{1, 2, 3},
			sliceB: []int{3, 4, 5},
			want:   []int{1, 2, 3, 4, 5},
		},
		{
			name:   "no overlap",
			sliceA: []int{1, 2},
			sliceB: []int{3, 4},
			want:   []int{1, 2, 3, 4},
		},
		{
			name:   "complete overlap",
			sliceA: []int{1, 2, 3},
			sliceB: []int{1, 2, 3},
			want:   []int{1, 2, 3},
		},
		{
			name:   "empty sliceA",
			sliceA: []int{},
			sliceB: []int{1, 2},
			want:   []int{1, 2},
		},
		{
			name:   "empty sliceB",
			sliceA: []int{1, 2},
			sliceB: []int{},
			want:   []int{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Union(tt.sliceA, tt.sliceB)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Union() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnion_NilHandling(t *testing.T) {
	t.Run("both nil returns nil", func(t *testing.T) {
		var sliceA, sliceB []int
		got := Union(sliceA, sliceB)
		if got != nil {
			t.Errorf("Expected nil, got %v", got)
		}
	})

	t.Run("nil sliceA returns sliceB elements", func(t *testing.T) {
		var sliceA []int
		sliceB := []int{1, 2}
		got := Union(sliceA, sliceB)
		if !reflect.DeepEqual(got, sliceB) {
			t.Errorf("Expected %v, got %v", sliceB, got)
		}
	})
}

func TestUnion_WithDuplicates(t *testing.T) {
	sliceA := []int{1, 2, 2, 3}
	sliceB := []int{2, 3, 3, 4}
	got := Union(sliceA, sliceB)
	want := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Union() = %v, want %v", got, want)
	}
}

// Test Suite for Deduplicate Function

func TestDeduplicate_Basic(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{
			name:  "with duplicates",
			input: []int{1, 2, 2, 3, 3, 3, 4},
			want:  []int{1, 2, 3, 4},
		},
		{
			name:  "no duplicates",
			input: []int{1, 2, 3, 4},
			want:  []int{1, 2, 3, 4},
		},
		{
			name:  "all duplicates",
			input: []int{1, 1, 1, 1},
			want:  []int{1},
		},
		{
			name:  "single element",
			input: []int{1},
			want:  []int{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Deduplicate(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Deduplicate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeduplicate_NilVsEmpty(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		var input []int
		got := Deduplicate(input)
		if got != nil {
			t.Errorf("Expected nil, got %v", got)
		}
	})

	t.Run("empty slice returns empty slice", func(t *testing.T) {
		input := []int{}
		got := Deduplicate(input)
		if got == nil {
			t.Error("Expected empty slice, got nil")
		}
		if len(got) != 0 {
			t.Errorf("Expected empty slice, got %v", got)
		}
	})
}

func TestDeduplicate_PreservesOrder(t *testing.T) {
	input := []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5}
	got := Deduplicate(input)
	want := []int{3, 1, 4, 5, 9, 2, 6}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Deduplicate() = %v, want %v (order not preserved)", got, want)
	}
}

// Test Suite for ToMap Function

func TestToMap_Basic(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	users := []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}

	got := ToMap(users, func(u User) int { return u.ID })

	if len(got) != 3 {
		t.Errorf("Expected map length 3, got %d", len(got))
	}

	if got[1].Name != "Alice" {
		t.Errorf("Expected Alice, got %s", got[1].Name)
	}
	if got[2].Name != "Bob" {
		t.Errorf("Expected Bob, got %s", got[2].Name)
	}
}

func TestToMap_DuplicateKeys(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	users := []User{
		{ID: 1, Name: "Alice"},
		{ID: 1, Name: "Alice2"},
		{ID: 2, Name: "Bob"},
	}

	got := ToMap(users, func(u User) int { return u.ID })

	if len(got) != 2 {
		t.Errorf("Expected map length 2, got %d", len(got))
	}

	// Last value should win
	if got[1].Name != "Alice2" {
		t.Errorf("Expected Alice2 (last value), got %s", got[1].Name)
	}
}

func TestToMap_EmptyInput(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	var users []User
	got := ToMap(users, func(u User) int { return u.ID })

	if got == nil {
		t.Error("Expected empty map, got nil")
	}
	if len(got) != 0 {
		t.Errorf("Expected empty map, got length %d", len(got))
	}
}

func TestToMap_NilInput(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	got := ToMap(nil, func(u User) int { return u.ID })

	if got == nil {
		t.Error("Expected empty map, got nil")
	}
	if len(got) != 0 {
		t.Errorf("Expected empty map, got length %d", len(got))
	}
}

// Benchmark Tests

func BenchmarkDifference(b *testing.B) {
	sliceA := make([]int, 1000)
	sliceB := make([]int, 500)
	for i := range sliceA {
		sliceA[i] = i
	}
	for i := range sliceB {
		sliceB[i] = i * 2
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Difference(sliceA, sliceB)
	}
}

func BenchmarkIntersection(b *testing.B) {
	sliceA := make([]int, 1000)
	sliceB := make([]int, 1000)
	for i := range sliceA {
		sliceA[i] = i
		sliceB[i] = i + 500
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Intersection(sliceA, sliceB)
	}
}

func BenchmarkUnion(b *testing.B) {
	sliceA := make([]int, 1000)
	sliceB := make([]int, 1000)
	for i := range sliceA {
		sliceA[i] = i
		sliceB[i] = i + 500
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Union(sliceA, sliceB)
	}
}

func BenchmarkDeduplicate(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = i % 100 // Create duplicates
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Deduplicate(input)
	}
}
