package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	assert.Equal(t, "duplicate key", ErrDuplicateKey.Error())
	assert.Equal(t, "key function not set", ErrKeyNotSet.Error())
	assert.Equal(t, "orphaned node", ErrOrphanedNode.Error())
	assert.Equal(t, "cycle detected", ErrCycle.Error())
}