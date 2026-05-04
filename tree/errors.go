package tree

import "errors"

var (
	ErrDuplicateKey = errors.New("duplicate key")
	ErrKeyNotSet    = errors.New("key function not set")
	ErrOrphanedNode = errors.New("orphaned node")
	ErrCycle        = errors.New("cycle detected")
)