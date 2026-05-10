package tree

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTree_Walk_Level(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root"},
		{ID: 2, Name: "Child1", ParentID: 1},
		{ID: 3, Name: "Child2", ParentID: 1},
		{ID: 4, Name: "Grandchild", ParentID: 2},
	})
	tree, err := b.Build()
	require.NoError(t, err)

	var levels []int
	tree.Walk(func(n *Node[TestItem], _ *Node[TestItem]) bool {
		levels = append(levels, n.Level)
		return true
	})

	assert.Equal(t, []int{1, 2, 3, 2}, levels)
}

func TestBuilder_Build_SetsNodeLevel(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root"},
		{ID: 2, Name: "Child1", ParentID: 1},
		{ID: 3, Name: "Child2", ParentID: 1},
		{ID: 4, Name: "Grandchild", ParentID: 2},
	})
	tree, err := b.Build()
	require.NoError(t, err)

	roots := tree.Roots()
	require.Len(t, roots, 1)
	assert.Equal(t, 1, roots[0].Level)
	assert.Equal(t, 2, roots[0].Children[0].Level)
	assert.Equal(t, 2, roots[0].Children[1].Level)
	assert.Equal(t, 3, roots[0].Children[0].Children[0].Level)
}

func TestTree_Walk_Parent(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root"},
		{ID: 2, Name: "Child", ParentID: 1},
		{ID: 3, Name: "Grandchild", ParentID: 2},
	})
	tree, err := b.Build()
	require.NoError(t, err)

	var parentNames []string
	tree.Walk(func(_ *Node[TestItem], parent *Node[TestItem]) bool {
		if parent == nil {
			parentNames = append(parentNames, "")
		} else {
			parentNames = append(parentNames, parent.Item.Name)
		}
		return true
	})

	assert.Equal(t, []string{"", "Root", "Child"}, parentNames)
}

func TestTree_Walk_Stop(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).WithItems([]TestItem{
		{ID: 1, Name: "A"},
		{ID: 2, Name: "B"},
		{ID: 3, Name: "C"},
	})
	tree, err := b.Build()
	require.NoError(t, err)

	var names []string
	tree.Walk(func(n *Node[TestItem], _ *Node[TestItem]) bool {
		names = append(names, n.Item.Name)
		return n.Item.Name != "B"
	})

	assert.Equal(t, []string{"A", "B"}, names)
}

func TestTree_Walk_ClonePreservesLevel(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root"},
		{ID: 2, Name: "Child", ParentID: 1},
	})
	tree, err := b.Build()
	require.NoError(t, err)

	tree.Walk(func(*Node[TestItem], *Node[TestItem]) bool {
		return true
	})

	cloned := tree.Clone()
	var levels []int
	cloned.Walk(func(n *Node[TestItem], _ *Node[TestItem]) bool {
		levels = append(levels, n.Level)
		return true
	})

	assert.Equal(t, []int{1, 2}, levels)
}

func TestTree_Map_PreservesLevel(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root"},
		{ID: 2, Name: "Child", ParentID: 1},
	})
	tree, err := b.Build()
	require.NoError(t, err)

	mapped := tree.Map(func(c TestItem) TestItem { return c }, func(c TestItem) int { return c.ID })

	assert.Len(t, mapped.Roots(), 1)
	assert.Equal(t, 1, mapped.Roots()[0].Level)
	assert.Equal(t, 2, mapped.Roots()[0].Children[0].Level)
}

func TestTree_Map_RebuildsParentIndex(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root"},
		{ID: 2, Name: "Child", ParentID: 1},
	})
	tree, err := b.Build()
	require.NoError(t, err)

	mapped := tree.Map(func(item TestItem) TestItem {
		item.ID += 10
		return item
	}, keyFn)

	parent, ok := mapped.ParentOf(12)
	require.True(t, ok)
	assert.Equal(t, 11, parent)
}

func TestTree_Filter_RebuildsConsistentTree(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root"},
		{ID: 2, Name: "KeepChild", ParentID: 1},
		{ID: 3, Name: "DropChild", ParentID: 1},
		{ID: 4, Name: "KeepGrandchild", ParentID: 3},
	})
	tree, err := b.Build()
	require.NoError(t, err)

	filtered := tree.Filter(func(node *Node[TestItem]) bool {
		return node.Item.ID == 2 || node.Item.ID == 4
	})

	assert.Equal(t, 2, filtered.Len())
	assert.True(t, filtered.ContainsKey(2))
	assert.True(t, filtered.ContainsKey(4))
	assert.False(t, filtered.ContainsKey(1))
	_, ok := filtered.ParentOf(4)
	assert.False(t, ok)

	roots := filtered.Roots()
	require.Len(t, roots, 2)
	assert.Equal(t, []int{2, 4}, []int{roots[0].Item.ID, roots[1].Item.ID})
}

func TestTree_PathTo(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root"},
		{ID: 2, Name: "Child", ParentID: 1},
		{ID: 3, Name: "Grandchild", ParentID: 2},
	})
	tree, err := b.Build()
	require.NoError(t, err)

	path, ok := tree.PathTo(3)
	require.True(t, ok)
	assert.Equal(t, []int{1, 2, 3}, []int{path[0].Item.ID, path[1].Item.ID, path[2].Item.ID})

	_, ok = tree.PathTo(999)
	assert.False(t, ok)
}

func TestTree_Descendants(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root"},
		{ID: 2, Name: "Child", ParentID: 1},
		{ID: 3, Name: "Grandchild", ParentID: 2},
		{ID: 4, Name: "Sibling", ParentID: 1},
	})
	tree, err := b.Build()
	require.NoError(t, err)

	descendants, ok := tree.Descendants(1)
	require.True(t, ok)
	assert.Equal(t, []int{2, 3, 4}, []int{
		descendants[0].Item.ID,
		descendants[1].Item.ID,
		descendants[2].Item.ID,
	})

	descendants, ok = tree.Descendants(4)
	require.True(t, ok)
	assert.Empty(t, descendants)

	_, ok = tree.Descendants(999)
	assert.False(t, ok)
}

func TestTree_Subtree_SetsRootLevel(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root"},
		{ID: 2, Name: "Child", ParentID: 1},
		{ID: 3, Name: "Grandchild", ParentID: 2},
	})
	tree, err := b.Build()
	require.NoError(t, err)

	subtree, ok := tree.Subtree(2)
	require.True(t, ok)
	assert.Len(t, subtree.Roots(), 1)
	assert.Equal(t, 1, subtree.Roots()[0].Level)
	assert.Equal(t, 2, subtree.Roots()[0].Children[0].Level)
}

func TestTree_Subtree_ConcurrentReaders(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root"},
		{ID: 2, Name: "Child", ParentID: 1},
		{ID: 3, Name: "Grandchild", ParentID: 2},
	})
	tree, err := b.Build()
	require.NoError(t, err)

	var wg sync.WaitGroup
	errCh := make(chan error, 25)
	for range 25 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			subtree, ok := tree.Subtree(2)
			if !ok {
				errCh <- assert.AnError
				return
			}
			if subtree.Len() != 2 {
				errCh <- assert.AnError
			}
		}()
	}
	wg.Wait()
	close(errCh)
	assert.Empty(t, errCh)
}
