package tree

import "errors"

var (
	// ErrDuplicateKey reports two items resolved to the same key.
	ErrDuplicateKey = errors.New("duplicate key")
	// ErrKeyNotSet reports an operation that requires Builder.KeyBy.
	ErrKeyNotSet = errors.New("key function not set")
	// ErrOrphanedNode reports a parent key that does not exist in the tree.
	ErrOrphanedNode = errors.New("orphaned node")
	// ErrCycle reports a parent relationship cycle.
	ErrCycle = errors.New("cycle detected")
	// ErrKeyNotFound reports that an operation referenced a missing key.
	ErrKeyNotFound = errors.New("key not found")
	// ErrInvalidMove reports an invalid tree move request.
	ErrInvalidMove = errors.New("invalid move")
)
