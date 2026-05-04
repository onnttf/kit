package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNode(t *testing.T) {
	node := &Node[TestItem]{Item: TestItem{ID: 1}}
	assert.Equal(t, 1, node.Item.ID)
}

func TestCloneNode(t *testing.T) {
	child := &Node[TestItem]{Item: TestItem{ID: 2}, Level: 2}
	parent := &Node[TestItem]{
		Item:     TestItem{ID: 1},
		Children: []*Node[TestItem]{child},
		Level:    1,
	}

	cloned := cloneNode(parent)

	assert.Equal(t, parent.Item, cloned.Item)
	assert.Equal(t, parent.Level, cloned.Level)
	assert.Len(t, cloned.Children, 1)
	assert.Equal(t, child.Item, cloned.Children[0].Item)
	assert.Equal(t, child.Level, cloned.Children[0].Level)
	assert.NotSame(t, parent, cloned)
	assert.NotSame(t, child, cloned.Children[0])
}

func TestCloneNodeView_Nil(t *testing.T) {
	assert.Nil(t, cloneNode[*Node[TestItem]](nil))
}
