package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlattener_Flatten(t *testing.T) {
	root := &Node[TestItem]{Item: TestItem{ID: 1, Name: "Root"}}
	child := &Node[TestItem]{Item: TestItem{ID: 2, Name: "Child"}}
	grandchild := &Node[TestItem]{Item: TestItem{ID: 3, Name: "Grandchild"}}
	root.Children = []*Node[TestItem]{child}
	child.Children = []*Node[TestItem]{grandchild}

	f := NewFlattener[TestItem, int]().KeyBy(keyFn).ParentBy(func(item TestItem, parentKey int) TestItem {
		item.ParentID = parentKey
		return item
	})

	result, err := f.Flatten([]*Node[TestItem]{root})
	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, 1, result[0].ID)
	assert.Equal(t, 2, result[1].ID)
	assert.Equal(t, 3, result[2].ID)
	assert.Equal(t, 0, result[0].ParentID)
	assert.Equal(t, 1, result[1].ParentID)
	assert.Equal(t, 2, result[2].ParentID)
}

func TestFlattener_Flatten_MissingKeyFn(t *testing.T) {
	f := NewFlattener[TestItem, int]()
	_, err := f.Flatten(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "KeyBy not set")
}

func TestFlattener_MultipleRoots(t *testing.T) {
	root1 := &Node[TestItem]{Item: TestItem{ID: 1, Name: "Root1"}}
	child := &Node[TestItem]{Item: TestItem{ID: 2, Name: "Child"}}
	root2 := &Node[TestItem]{Item: TestItem{ID: 3, Name: "Root2"}}
	root1.Children = []*Node[TestItem]{child}

	f := NewFlattener[TestItem, int]().KeyBy(keyFn).ParentBy(func(item TestItem, parentKey int) TestItem {
		item.ParentID = parentKey
		return item
	})

	result, err := f.Flatten([]*Node[TestItem]{root1, root2})
	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, 1, result[0].ID)
	assert.Equal(t, 2, result[1].ID)
	assert.Equal(t, 3, result[2].ID)
	assert.Equal(t, 0, result[0].ParentID)
	assert.Equal(t, 1, result[1].ParentID)
	assert.Equal(t, 0, result[2].ParentID)
}