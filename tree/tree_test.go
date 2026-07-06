package tree

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"
)

type item struct {
	ID     int
	Parent int
	Name   string
	Sort   int
}

func TestBuildAndQuery(t *testing.T) {
	tr, err := testBuilder().Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	roots := tr.Roots()
	if len(roots) != 1 || roots[0].Key != 1 {
		t.Fatalf("Roots() = %+v", roots)
	}

	node, ok := tr.Get(2)
	if !ok || node.Key != 2 || node.Parent == nil || *node.Parent != 1 || node.Depth != 1 {
		t.Fatalf("Get(2) = %+v, %v", node, ok)
	}

	children, ok := tr.Children(1)
	if !ok || keys(children) != "3,2" {
		t.Fatalf("Children(1) = %+v, %v", children, ok)
	}
}

func TestBuildAllowsParentAfterChildAndKeepsInputOrder(t *testing.T) {
	tr, err := NewBuilder[item, int]().
		KeyBy(func(v item) int { return v.ID }).
		ParentBy(func(v item) (int, bool) { return v.Parent, v.Parent != 0 }).
		Add(
			item{ID: 2, Parent: 1, Name: "child b"},
			item{ID: 3, Parent: 1, Name: "child a"},
			item{ID: 1, Name: "root"},
		).
		Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	children, ok := tr.Children(1)
	if !ok || keys(children) != "2,3" {
		t.Fatalf("Children(1) = %+v, %v", children, ok)
	}
}

func TestBuilderPendingEdits(t *testing.T) {
	tr, err := testBuilder().
		Update(2, func(v *item) { v.Name = "changed" }).
		Move(3, 2).
		Remove(4).
		Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	node, _ := tr.Get(2)
	if node.Value.Name != "changed" {
		t.Fatalf("updated value = %+v", node.Value)
	}
	children, _ := tr.Children(2)
	if len(children) != 1 || children[0].Key != 3 {
		t.Fatalf("Children(2) = %+v", children)
	}
	if _, ok := tr.Get(4); ok {
		t.Fatalf("removed key still exists")
	}
}

func TestBuildValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		b    *Builder[item, int]
		want error
	}{
		{
			name: "missing key function",
			b:    NewBuilder[item, int]().Add(item{ID: 1}),
			want: ErrKeyNotSet,
		},
		{
			name: "duplicate key",
			b: NewBuilder[item, int]().
				KeyBy(func(v item) int { return v.ID }).
				Add(item{ID: 1}, item{ID: 1}),
			want: ErrDuplicateKey,
		},
		{
			name: "missing parent",
			b: NewBuilder[item, int]().
				KeyBy(func(v item) int { return v.ID }).
				ParentBy(func(v item) (int, bool) { return v.Parent, v.Parent != 0 }).
				Add(item{ID: 1, Parent: 9}),
			want: ErrMissingParent,
		},
		{
			name: "move cycle",
			b:    testBuilder().Move(1, 4),
			want: ErrCycle,
		},
		{
			name: "update missing",
			b:    testBuilder().Update(99, func(*item) {}),
			want: ErrKeyNotFound,
		},
		{
			name: "update nil function",
			b:    testBuilder().Update(1, nil),
			want: ErrInvalidInput,
		},
		{
			name: "remove missing",
			b:    testBuilder().Remove(99),
			want: ErrKeyNotFound,
		},
		{
			name: "move missing key",
			b:    testBuilder().Move(99, 1),
			want: ErrKeyNotFound,
		},
		{
			name: "move missing parent",
			b:    testBuilder().Move(2, 99),
			want: ErrMissingParent,
		},
		{
			name: "move self parent",
			b:    testBuilder().Move(2, 2),
			want: ErrCycle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.b.Build()
			if !errors.Is(err, tt.want) {
				t.Fatalf("Build() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestClonedViews(t *testing.T) {
	tr, err := testBuilder().Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	roots := tr.Roots()
	roots[0].Value.Name = "mutated"
	roots[0].Children = nil

	root, _ := tr.Get(1)
	if root.Value.Name == "mutated" || len(root.Children) == 0 {
		t.Fatalf("query leaked mutable internals: %+v", root)
	}

	_, err = tr.Walk(func(node Node[item, int], parent *Node[item, int]) bool {
		node.Value.Name = "walk-mutated"
		if node.Key == 2 && (parent == nil || parent.Key != 1) {
			t.Fatalf("parent for node 2 = %+v", parent)
		}
		return true
	})
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}
	root, _ = tr.Get(1)
	if root.Value.Name == "walk-mutated" {
		t.Fatalf("walk leaked mutable internals")
	}
}

func TestBuilderClone(t *testing.T) {
	b := testBuilder()
	clone := b.Clone().Add(item{ID: 9, Parent: 99})
	if errs := clone.Validate(); len(errs) == 0 {
		t.Fatalf("Validate() expected missing parent for cloned builder addition")
	}

	if _, err := b.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
}

func TestNilAndMissingQueries(t *testing.T) {
	var tr *Tree[item, int]
	if _, ok := tr.Get(1); ok {
		t.Fatalf("nil Get() ok = true")
	}
	if got := tr.Roots(); got != nil {
		t.Fatalf("nil Roots() = %v", got)
	}
	if _, ok := tr.Children(1); ok {
		t.Fatalf("nil Children() ok = true")
	}
	ok, err := tr.Walk(func(Node[item, int], *Node[item, int]) bool {
		return true
	})
	if !ok || err != nil {
		t.Fatalf("nil Walk() = %v, %v", ok, err)
	}

	tr, err = testBuilder().Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if _, ok := tr.Get(99); ok {
		t.Fatalf("Get(missing) ok = true")
	}
	if _, ok := tr.Children(99); ok {
		t.Fatalf("Children(missing) ok = true")
	}
	if _, err := tr.Walk(nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Walk(nil) error = %v", err)
	}
}

func TestRemoveSubtree(t *testing.T) {
	tr, err := testBuilder().Remove(2).Build()
	if err != nil {
		t.Fatalf("Build(remove subtree) error = %v", err)
	}
	if _, ok := tr.Get(2); ok {
		t.Fatalf("removed key 2 still exists")
	}
	if _, ok := tr.Get(4); ok {
		t.Fatalf("removed child key 4 still exists")
	}
}

func TestTreeConcurrentReaders(t *testing.T) {
	tr, err := testBuilder().Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			_, _ = tr.Get(2)
			_, _ = tr.Children(1)
			_ = tr.Roots()
			_, _ = tr.Walk(func(Node[item, int], *Node[item, int]) bool { return true })
		})
	}
	wg.Wait()
}

func testBuilder() *Builder[item, int] {
	return NewBuilder[item, int]().
		KeyBy(func(v item) int { return v.ID }).
		ParentBy(func(v item) (int, bool) { return v.Parent, v.Parent != 0 }).
		SortFunc(func(a, b item) int { return a.Sort - b.Sort }).
		Add(
			item{ID: 1, Name: "root", Sort: 1},
			item{ID: 2, Parent: 1, Name: "child b", Sort: 2},
			item{ID: 3, Parent: 1, Name: "child a", Sort: 1},
			item{ID: 4, Parent: 2, Name: "leaf", Sort: 1},
		)
}

func keys(nodes []Node[item, int]) string {
	parts := make([]string, 0, len(nodes))
	for _, n := range nodes {
		parts = append(parts, strconv.Itoa(n.Key))
	}
	return strings.Join(parts, ",")
}
