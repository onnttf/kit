package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStats(t *testing.T) {
	b := NewBuilder[TestItem, int]()
	b.KeyBy(keyFn).ParentBy(parentFn).WithItems([]TestItem{
		{ID: 1, Name: "Root", ParentID: 1},
		{ID: 2, Name: "Child1", ParentID: 1},
		{ID: 3, Name: "Child2", ParentID: 1},
		{ID: 4, Name: "Grandchild", ParentID: 2},
	})
	tree, err := b.Build()
	require.NoError(t, err)

	stats := tree.Stats()

	assert.Equal(t, 4, stats.TotalNodes)
	assert.Equal(t, 1, stats.RootNodes)
	assert.Equal(t, 3, stats.MaxDepth)
	assert.Equal(t, 2, stats.LeafNodes)
	assert.Greater(t, stats.AvgChildren, 0.0)
	assert.Greater(t, stats.AvgDepth, 0.0)
}

func TestStats_Empty(t *testing.T) {
	b := NewBuilder[TestItem, int]().KeyBy(keyFn)
	tree, err := b.Build()
	require.NoError(t, err)

	stats := tree.Stats()
	assert.Equal(t, 0, stats.TotalNodes)
	assert.Equal(t, 0, stats.RootNodes)
	assert.Equal(t, 0, stats.MaxDepth)
	assert.Equal(t, 0, stats.LeafNodes)
}

func TestBuilder_Statistics_EmptyTree(t *testing.T) {
	b := NewBuilder[TestItem, int]().KeyBy(keyFn).WithItems([]TestItem{})
	stats, err := b.Statistics()
	require.NoError(t, err)
	assert.Equal(t, 0, stats.TotalNodes)
	assert.Equal(t, 0, stats.RootNodes)
	assert.Equal(t, 0, stats.MaxDepth)
	assert.Equal(t, 0, stats.LeafNodes)
}