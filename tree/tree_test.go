package tree

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Category struct {
	ID       int
	Name     string
	ParentID int
	Sort     int
}

func getCategoryKey(c Category) any { return c.ID }
func getCategoryParent(c Category) any {
	if c.ParentID == 0 {
		return c.ID
	}
	return c.ParentID
}
func getCategorySort(c Category) int { return c.Sort }

func TestNewBuilder(t *testing.T) {
	b := NewBuilder[Category]()
	assert.NotNil(t, b)
}

func TestBuilder_KeyBy(t *testing.T) {
	b := NewBuilder[Category]()
	result := b.KeyBy(getCategoryKey)
	assert.Same(t, b, result)
	assert.NotNil(t, b.keyFn)
}

func TestBuilder_ParentBy(t *testing.T) {
	b := NewBuilder[Category]()
	result := b.ParentBy(getCategoryParent)
	assert.Same(t, b, result)
}

func TestBuilder_SortBy(t *testing.T) {
	b := NewBuilder[Category]()
	result := b.SortBy(getCategorySort)
	assert.Same(t, b, result)
}

func TestBuilder_WithItems(t *testing.T) {
	b := NewBuilder[Category]()
	items := []Category{{ID: 1, Name: "Root"}, {ID: 2, Name: "Child"}}
	result := b.WithItems(items)
	assert.Same(t, b, result)
	assert.Len(t, b.items, 2)
}

func TestBuilder_AddItem(t *testing.T) {
	b := NewBuilder[Category]()
	b.AddItem(Category{ID: 1, Name: "Item"})
	assert.Len(t, b.items, 1)
}

func TestBuilder_Filter(t *testing.T) {
	b := NewBuilder[Category]()
	b.KeyBy(getCategoryKey).WithItems([]Category{{ID: 1}, {ID: 2}})
	nb := b.Filter(func(c Category) bool { return c.ID == 1 })
	assert.Len(t, nb.items, 1)
}

func TestBuilder_Transform(t *testing.T) {
	b := NewBuilder[Category]()
	b.KeyBy(getCategoryKey).WithItems([]Category{{ID: 1, Name: "Original"}})
	b.Transform(func(c *Category) { c.Name = "Transformed" })
	assert.Equal(t, "Transformed", b.items[0].item.Name)
}

func TestBuilder_Find(t *testing.T) {
	b := NewBuilder[Category]()
	b.KeyBy(getCategoryKey).WithItems([]Category{{ID: 1}, {ID: 2}})
	node := b.Find(func(c Category) bool { return c.ID == 2 })
	assert.NotNil(t, node)
}

func TestBuilder_Contains(t *testing.T) {
	b := NewBuilder[Category]()
	b.KeyBy(getCategoryKey).WithItems([]Category{{ID: 1}})
	assert.True(t, b.Contains(1))
	assert.False(t, b.Contains(999))
}

func TestBuilder_Build_DuplicateKey(t *testing.T) {
	b := NewBuilder[Category]()
	b.KeyBy(getCategoryKey).WithItems([]Category{{ID: 1}, {ID: 1}})
	_, _, err := b.Build()
	assert.ErrorIs(t, err, ErrDuplicateKey)
}

func TestBuilder_Build_NilKeyFn(t *testing.T) {
	b := NewBuilder[Category]()
	roots, _, err := b.Build()
	assert.ErrorIs(t, err, ErrKeyNotSet)
	assert.Nil(t, roots)
}

func TestBuilder_Clone(t *testing.T) {
	b := NewBuilder[Category]()
	b.KeyBy(getCategoryKey).WithItems([]Category{{ID: 1}})
	clone := b.Clone()
	assert.NotSame(t, b, clone)
	assert.Len(t, clone.items, 1)
}

func TestBuilder_Statistics(t *testing.T) {
	b := NewBuilder[Category]()
	b.KeyBy(getCategoryKey).WithItems([]Category{{ID: 1}, {ID: 2}})
	stats := b.Statistics()
	assert.Equal(t, 2, stats["totalNodes"])
}

func TestBuilder_Statistics_Empty(t *testing.T) {
	b := NewBuilder[Category]()
	stats := b.Statistics()
	assert.Equal(t, 0, stats["totalNodes"])
}

func TestErrors(t *testing.T) {
	assert.Equal(t, "duplicate key", ErrDuplicateKey.Error())
	assert.Equal(t, "key function not set", ErrKeyNotSet.Error())
	assert.Equal(t, "orphaned node", ErrOrphanedNode.Error())
	assert.Equal(t, "cycle detected", ErrCycle.Error())
}

func TestNode(t *testing.T) {
	node := &Node[Category]{Item: Category{ID: 1}}
	assert.Equal(t, 1, node.Item.ID)
}

func TestBuilder_Map(t *testing.T) {
	b := NewBuilder[Category]()
	b.KeyBy(getCategoryKey).WithItems([]Category{{ID: 1}})
	nb := b.Map(func(c Category) Category { return c }, func(c Category) any { return c.ID })
	stats := nb.Statistics()
	assert.Equal(t, 1, stats["totalNodes"])
}

func TestBuilder_MoveItem(t *testing.T) {
	b := NewBuilder[Category]()
	b.KeyBy(getCategoryKey).WithItems([]Category{{ID: 1}, {ID: 2}})
	result := b.MoveItem(1, 2)
	assert.Same(t, b, result)
}

func TestBuilder_Validate_Valid(t *testing.T) {
	b := NewBuilder[Category]()
	b.KeyBy(getCategoryKey).
		ParentBy(getCategoryParent).
		WithItems([]Category{
			{ID: 1, Name: "Root"},
			{ID: 2, Name: "Child", ParentID: 1},
		})
	errs := b.Validate()
	assert.Empty(t, errs)
}

func TestBuilder_Validate_OrphanedNode(t *testing.T) {
	b := NewBuilder[Category]()
	b.KeyBy(getCategoryKey).
		ParentBy(getCategoryParent).
		WithItems([]Category{
			{ID: 1, Name: "Root"},
			{ID: 2, Name: "Orphan", ParentID: 999},
		})
	errs := b.Validate()
	assert.NotEmpty(t, errs)
	assert.True(t, errors.Is(errs[0], ErrOrphanedNode))
}

func TestBuilder_Validate_NilKeyFn(t *testing.T) {
	b := NewBuilder[Category]()
	errs := b.Validate()
	assert.Len(t, errs, 1)
}