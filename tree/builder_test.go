package tree

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	assert.NotNil(t, b)
}

func TestBuilder_KeyBy(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	result := b.KeyBy(keyFn)
	assert.Same(t, b, result)
	assert.NotNil(t, b.keyFn)
}

func TestBuilder_ParentBy(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	result := b.ParentBy(parentFn)
	assert.Same(t, b, result)
}

func TestBuilder_SortBy(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	result := b.SortBy(sortFn)
	assert.Same(t, b, result)
}

func TestBuilder_WithItems(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	items := []TestItem{{ID: 1, Name: "Root"}, {ID: 2, Name: "Child"}}
	result := b.WithItems(items)
	assert.Same(t, b, result)
	assert.Len(t, b.items, 2)
}

func TestBuilder_AddItem(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.AddItem(TestItem{ID: 1, Name: "Item"})
	assert.Len(t, b.items, 1)
}

func TestBuilder_AddItemWithParent(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.AddItemWithParent(TestItem{ID: 2, Name: "Child"}, 1)
	require.Len(t, b.items, 1)
	assert.Equal(t, 2, b.items[0].data.ID)
	assert.True(t, b.items[0].hasParent)
}

func TestBuilder_Filter(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).WithItems([]TestItem{{ID: 1}, {ID: 2}})
	nb := b.Filter(func(c TestItem) bool { return c.ID == 1 })
	assert.Len(t, nb.items, 1)
}

func TestBuilder_Transform(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).WithItems([]TestItem{{ID: 1, Name: "Original"}})
	b.Transform(func(c *TestItem) { c.Name = "Transformed" })
	assert.Equal(t, "Transformed", b.items[0].data.Name)
}

func TestBuilder_Find(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).WithItems([]TestItem{{ID: 1}, {ID: 2}})
	node, err := b.Find(func(c TestItem) bool { return c.ID == 2 })
	require.NoError(t, err)
	assert.NotNil(t, node)
}

func TestBuilder_Contains(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).WithItems([]TestItem{{ID: 1}})
	found, err := b.Contains(1)
	require.NoError(t, err)
	assert.True(t, found)

	found, err = b.Contains(999)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestBuilder_Build_DuplicateKey(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).WithItems([]TestItem{{ID: 1}, {ID: 1}})
	_, err := b.Build()
	assert.ErrorIs(t, err, ErrDuplicateKey)
}

func TestBuilder_Build_NilKeyFn(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	_, err := b.Build()
	assert.ErrorIs(t, err, ErrKeyNotSet)
}

func TestBuilder_Clone(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).WithItems([]TestItem{{ID: 1}})
	clone := b.Clone()
	assert.NotSame(t, b, clone)
	assert.Len(t, clone.items, 1)
}

func TestBuilder_Statistics(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root", ParentID: 1},
		{ID: 2, Name: "Child", ParentID: 1},
	})
	stats, err := b.Statistics()
	require.NoError(t, err)
	t.Logf("Stats: %+v", stats)
	assert.Equal(t, 2, stats.TotalNodes)
	assert.Equal(t, 1, stats.RootNodes)
	assert.Equal(t, 2, stats.MaxDepth)
	assert.Equal(t, 1, stats.LeafNodes)
}

func TestBuilder_Statistics_Empty(t *testing.T) {
	b := NewBuilder[TestItem, int]().KeyBy(keyFn)
	stats, err := b.Statistics()
	require.NoError(t, err)
	assert.Equal(t, 0, stats.TotalNodes)
}

func TestBuilder_Map(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).WithItems([]TestItem{{ID: 1}})
	nb := b.Map(func(c TestItem) TestItem { return c }, func(c TestItem) int { return c.ID })
	stats, err := nb.Statistics()
	require.NoError(t, err)
	assert.Equal(t, 1, stats.TotalNodes)
}

func TestBuilder_Validate_Valid(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).
		ParentBy(parentFn).
		WithItems([]TestItem{
			{ID: 1, Name: "Root"},
			{ID: 2, Name: "Child", ParentID: 1},
		})
	errs := b.Validate()
	assert.Empty(t, errs)
}

func TestBuilder_Validate_OrphanedNode(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).
		ParentBy(parentFn).
		WithItems([]TestItem{
			{ID: 1, Name: "Root"},
			{ID: 2, Name: "Orphan", ParentID: 999},
		})
	errs := b.Validate()
	assert.NotEmpty(t, errs)
	assert.Len(t, errs, 1)
}

func TestBuilder_Validate_NilKeyFn(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	errs := b.Validate()
	assert.Len(t, errs, 1)
}

func TestBuilder_Validate_MultipleErrors(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "a"},
		{ID: 1, Name: "b"},
		{ID: 2, ParentID: 99},
	})
	errs := b.Validate()
	assert.GreaterOrEqual(t, len(errs), 2)
}

func TestBuilder_Filter_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		items  []TestItem
		filter func(TestItem) bool
		want   int
	}{
		{
			name:   "empty result",
			items:  []TestItem{{ID: 1}, {ID: 2}},
			filter: func(TestItem) bool { return false },
			want:   0,
		},
		{
			name:   "all pass",
			items:  []TestItem{{ID: 1}, {ID: 2}},
			filter: func(TestItem) bool { return true },
			want:   2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder[TestItem, int]().KeyBy(keyFn)
			nb := b.WithItems(tt.items).Filter(tt.filter)
			assert.Len(t, nb.items, tt.want)
		})
	}
}

func TestBuilder_Map_EdgeCases(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).WithItems([]TestItem{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}})
	nb := b.Map(
		func(c TestItem) TestItem { c.Name = c.Name + "!"; return c },
		func(c TestItem) int { return c.ID },
	)
	assert.Len(t, nb.items, 2)
}

func TestBuilder_UpdateItem_KeyChange(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).
		WithItems([]TestItem{{ID: 1, Name: "A"}, {ID: 2, ParentID: 1}})

	err := b.UpdateItem(1, func(item *TestItem) { item.Name = "Updated" })
	assert.NoError(t, err)

	node, err := b.Find(func(item TestItem) bool { return item.ID == 1 })
	require.NoError(t, err)
	assert.Equal(t, "Updated", node.Item.Name)
}

func TestBuilder_UpdateItem_KeyConflictAfterChange(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).
		WithItems([]TestItem{{ID: 1, Name: "A"}, {ID: 2, Name: "B", ParentID: 1}})

	err := b.UpdateItem(1, func(item *TestItem) { item.ID = 2 })
	assert.ErrorIs(t, err, ErrDuplicateKey)
}

func TestBuilder_IsDescendant(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, ParentID: 1},
		{ID: 2, ParentID: 1},
		{ID: 3, ParentID: 2},
	})

	_, err := b.Build()
	require.NoError(t, err)

	t.Run("self is descendant", func(t *testing.T) {
		ok, err := b.IsDescendant(1, 1)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("direct child", func(t *testing.T) {
		ok, err := b.IsDescendant(1, 2)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("grandchild", func(t *testing.T) {
		ok, err := b.IsDescendant(1, 3)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("not descendant", func(t *testing.T) {
		ok, err := b.IsDescendant(3, 1)
		require.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestBuilder_RemoveItem(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root", ParentID: 1},
		{ID: 2, Name: "Child", ParentID: 1},
		{ID: 3, Name: "Grandchild", ParentID: 2},
		{ID: 4, Name: "Sibling", ParentID: 1},
	})

	require.NoError(t, b.RemoveItem(2))
	tree, err := b.Build()
	require.NoError(t, err)
	assert.Equal(t, 2, tree.Len())

	children, err := b.ChildrenOf(1)
	require.NoError(t, err)
	assert.Len(t, children, 1)
	assert.Equal(t, 4, children[0].Item.ID)

	assert.Error(t, b.RemoveItem(999))
}

func TestBuilder_MoveItem(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root", ParentID: 1},
		{ID: 2, Name: "Child", ParentID: 1},
		{ID: 3, Name: "Grandchild", ParentID: 2},
		{ID: 4, Name: "OtherRoot", ParentID: 4},
	})

	err := b.MoveItem(1, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "self-move not allowed")

	err = b.MoveItem(1, 3)
	assert.ErrorIs(t, err, ErrCycle)

	require.NoError(t, b.MoveItem(3, 4))
	tree, err := b.Build()
	require.NoError(t, err)
	assert.Equal(t, 4, tree.Len())

	children, err := b.ChildrenOf(4)
	require.NoError(t, err)
	assert.Len(t, children, 1)
	assert.Equal(t, 3, children[0].Item.ID)
}

func TestBuilder_MoveItem_Errors(t *testing.T) {
	b := NewBuilder[TestItem, int]().KeyBy(keyFn).ParentBy(parentFn)
	b.WithItems([]TestItem{{ID: 1, Name: "Root", ParentID: 1}})

	err := b.MoveItem(999, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")

	err = b.MoveItem(1, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parent key not found")
}

func TestBuilder_DepthAndChildrenOf(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root", ParentID: 1},
		{ID: 2, Name: "Child", ParentID: 1},
	})

	d, err := b.Depth(2)
	require.NoError(t, err)
	assert.Equal(t, 2, d)

	children, err := b.ChildrenOf(1)
	require.NoError(t, err)
	assert.Len(t, children, 1)
	assert.Equal(t, 2, children[0].Item.ID)
	assert.Equal(t, 2, children[0].Level)

	_, err = b.Depth(999)
	assert.Error(t, err)

	_, err = b.ChildrenOf(999)
	assert.Error(t, err)
}

func TestBuilder_SortBy_Effect(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).SortBy(sortFn).
		WithItems([]TestItem{
			{ID: 1, Name: "C", Sort: 3},
			{ID: 2, Name: "A", Sort: 1, ParentID: 1},
			{ID: 3, Name: "B", Sort: 2, ParentID: 1},
		})
	tree, err := b.Build()
	require.NoError(t, err)

	var names []string
	tree.Walk(func(n *Node[TestItem], _ *Node[TestItem]) bool {
		names = append(names, n.Item.Name)
		return true
	})
	assert.Equal(t, []string{"C", "A", "B"}, names)
}

func TestBuilder_ConcurrentWrites(t *testing.T) {
	b := NewBuilder[TestItem, int]().KeyBy(keyFn)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			b.AddItem(TestItem{ID: id})
		}(i)
	}
	wg.Wait()
	assert.Equal(t, 100, len(b.items))
}

func TestBuilder_Concurrent_BuildWhileAdd(t *testing.T) {
	b := NewBuilder[TestItem, int]().KeyBy(keyFn).ParentBy(parentFn)

	b.WithItems([]TestItem{{ID: 1, Name: "Root", ParentID: 1}})

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 2; i <= 50; i++ {
			b.AddItem(TestItem{ID: i, ParentID: 1})
		}
	}()

	go func() {
		defer wg.Done()
		for i := 51; i <= 100; i++ {
			b.AddItem(TestItem{ID: i, ParentID: 1})
		}
	}()

	wg.Wait()
	tree, err := b.Build()
	require.NoError(t, err)
	assert.Equal(t, 100, tree.Len())
}

func TestBuilder_ConcurrentBuildDuringAdd(t *testing.T) {
	b := NewBuilder[TestItem, int]().KeyBy(keyFn).ParentBy(parentFn)
	b.WithItems([]TestItem{{ID: 1, Name: "Root", ParentID: 1}})

	var wg sync.WaitGroup
	errCh := make(chan error, 20)

	for i := 2; i <= 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			b.AddItem(TestItem{ID: id, ParentID: 1})
		}(i)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := b.Build()
			if err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		require.NoError(t, err)
	}

	tree, err := b.Build()
	require.NoError(t, err)
	assert.Equal(t, 50, tree.Len())
}

func TestBuilder_SortByFunc(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).SortByFunc(func(a, b TestItem) int {
		return strings.Compare(a.Name, b.Name)
	}).WithItems([]TestItem{
		{ID: 1, Name: "C"},
		{ID: 2, Name: "A", ParentID: 1},
		{ID: 3, Name: "B", ParentID: 1},
	})
	tree, err := b.Build()
	require.NoError(t, err)

	var names []string
	tree.Walk(func(n *Node[TestItem], _ *Node[TestItem]) bool {
		names = append(names, n.Item.Name)
		return true
	})
	assert.Equal(t, []string{"C", "A", "B"}, names)
}

func TestBuilder_Subtree(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root", ParentID: 1},
		{ID: 2, Name: "Child", ParentID: 1},
		{ID: 3, Name: "Grandchild", ParentID: 2},
	})

	subtree, err := b.Subtree(1)
	require.NoError(t, err)
	assert.Equal(t, 3, subtree.Len())

	_, err = b.Subtree(999)
	assert.Error(t, err)
}

func TestBuilder_Orphans(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root", ParentID: 1},
		{ID: 3, Name: "Root2", ParentID: 3},
	})

	orphans, err := b.Orphans()
	require.NoError(t, err)
	assert.Len(t, orphans, 0)
}
