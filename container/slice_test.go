package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDifference(t *testing.T) {
	tests := []struct {
		name string
		s1   []int
		s2   []int
		want []int
	}{
		{"normal", []int{1, 2, 3, 4}, []int{2, 4}, []int{1, 3}},
		{"nil s1", nil, []int{1, 2}, nil},
		{"empty s2", []int{1, 2, 3}, []int{}, []int{1, 2, 3}},
		{"empty s1", []int{}, []int{1, 2}, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Difference(tt.s1, tt.s2))
		})
	}
}

func TestIntersection(t *testing.T) {
	tests := []struct {
		name string
		s1   []int
		s2   []int
		want []int
	}{
		{"normal", []int{1, 2, 3, 4}, []int{2, 4, 6}, []int{2, 4}},
		{"nil s1", nil, []int{1, 2}, nil},
		{"empty", []int{}, []int{1, 2}, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Intersection(tt.s1, tt.s2))
		})
	}
}

func TestUnion(t *testing.T) {
	tests := []struct {
		name string
		s1   []int
		s2   []int
		want []int
	}{
		{"normal", []int{1, 2}, []int{2, 3}, []int{1, 2, 3}},
		{"both nil", nil, nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Union(tt.s1, tt.s2))
		})
	}
}

func TestDeduplicate(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"normal", []int{1, 2, 2, 3, 3, 3}, []int{1, 2, 3}},
		{"nil", nil, nil},
		{"empty", []int{}, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Deduplicate(tt.input))
		})
	}
}

func TestToMap(t *testing.T) {
	type Person struct{ Name string }
	input := []Person{{Name: "Alice"}, {Name: "Bob"}}
	result := ToMap(input, func(p Person) string { return p.Name })
	assert.Len(t, result, 2)
}

func TestToMap_LastDuplicateKeyWins(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}
	input := []Person{{Name: "Alice", Age: 20}, {Name: "Alice", Age: 30}}

	result := ToMap(input, func(p Person) string { return p.Name })

	assert.Equal(t, 30, result["Alice"].Age)
}

func TestFlatMap(t *testing.T) {
	result := FlatMap([]int{1, 2}, func(int) []string { return []string{"a", "b"} })
	assert.Equal(t, []string{"a", "b", "a", "b"}, result)
}

func TestReduce(t *testing.T) {
	sum := Reduce([]int{1, 2, 3, 4}, 0, func(acc, val int) int { return acc + val })
	assert.Equal(t, 10, sum)
}

func TestFirst(t *testing.T) {
	result, ok := First([]int{1, 2, 3}, func(n int) bool { return n > 2 })
	assert.True(t, ok)
	assert.Equal(t, 3, result)
}

func TestFirst_NoMatch(t *testing.T) {
	result, ok := First([]int{1, 2, 3}, func(n int) bool { return n > 3 })
	assert.False(t, ok)
	assert.Zero(t, result)
}

func TestPartition(t *testing.T) {
	matches, nonMatches := Partition([]int{1, 2, 3, 4, 5}, func(n int) bool { return n%2 == 0 })
	assert.Equal(t, []int{2, 4}, matches)
	assert.Equal(t, []int{1, 3, 5}, nonMatches)
}

func TestGroupBy(t *testing.T) {
	type Person struct{ Name, Dept string }
	input := []Person{{Name: "Alice", Dept: "HR"}, {Name: "Bob", Dept: "IT"}, {Name: "Carol", Dept: "HR"}}
	result := GroupBy(input, func(p Person) string { return p.Dept })
	assert.Len(t, result["HR"], 2)
	assert.Len(t, result["IT"], 1)
}

func TestGroupBy_EmptyInput(t *testing.T) {
	result := GroupBy([]int{}, func(n int) int { return n })
	assert.Empty(t, result)
	assert.NotNil(t, result)
}

func BenchmarkDeduplicate(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = i % 100
	}
	for i := 0; i < b.N; i++ {
		Deduplicate(input)
	}
}
