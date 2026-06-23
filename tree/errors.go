package tree

import "errors"

var (
	ErrDuplicateKey = errors.New("tree: duplicate key")

	ErrKeyNotSet = errors.New("tree: key function not set")

	ErrOrphanedNode = errors.New("tree: orphaned node")

	ErrCycle = errors.New("tree: cycle detected")

	ErrKeyNotFound = errors.New("tree: key not found")

	ErrInvalidMove = errors.New("tree: invalid move")

	ErrItemNotFound = errors.New("tree: item not found")

	ErrNilCallback = errors.New("tree: callback is nil")
)
