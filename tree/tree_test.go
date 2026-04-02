package tree

import (
	"testing"
)

// testItem is the shared fixture type for all tests.
type testItem struct {
	ID       int
	ParentID int
	Sort     int
	Name     string
}

// newBuilder returns a fully configured Builder for testItem.
func newBuilder() *Builder[testItem, int] {
	return NewBuilder[testItem, int]().
		KeyBy(func(d testItem) int { return d.ID }).
		ParentBy(func(d testItem) int { return d.ParentID }).
		SortBy(func(d testItem) int { return d.Sort })
}

// newItems returns a standard 5-node fixture:
//
//	1 (root, self-ref ParentID=1)
//	├── 2 (Sort=2)
//	│   ├── 3 (Sort=1)
//	│   └── 4 (Sort=2)
//	└── 5 (Sort=3)
func newItems() []testItem {
	return []testItem{
		{ID: 1, ParentID: 1, Sort: 1, Name: "root"},
		{ID: 2, ParentID: 1, Sort: 2, Name: "child-A"},
		{ID: 3, ParentID: 2, Sort: 1, Name: "grandchild-1"},
		{ID: 4, ParentID: 2, Sort: 2, Name: "grandchild-2"},
		{ID: 5, ParentID: 1, Sort: 3, Name: "child-B"},
	}
}

func TestNewBuilder(t *testing.T) {
	b := NewBuilder[testItem, int]()
	if b == nil {
		t.Fatal("NewBuilder returned nil")
	}
}

func TestKeyByNilIsNoop(t *testing.T) {
	b := NewBuilder[testItem, int]()
	b2 := b.KeyBy(nil)
	if b2 != b {
		t.Error("KeyBy(nil) should return same builder")
	}
}

func TestParentByNilIsNoop(t *testing.T) {
	b := NewBuilder[testItem, int]()
	b2 := b.ParentBy(nil)
	if b2 != b {
		t.Error("ParentBy(nil) should return same builder")
	}
}

func TestSortByNilIsNoop(t *testing.T) {
	b := NewBuilder[testItem, int]()
	b2 := b.SortBy(nil)
	if b2 != b {
		t.Error("SortBy(nil) should return same builder")
	}
}
