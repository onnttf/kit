package tree

// Node is an immutable tree node returned from built trees.
type Node[T any] struct {
	Item     T
	Children []*Node[T]
	Level    int
}

func cloneNode[T any](n *Node[T]) *Node[T] {
	if n == nil {
		return nil
	}

	cp := &Node[T]{Item: n.Item, Level: n.Level}

	if len(n.Children) > 0 {
		cp.Children = make([]*Node[T], len(n.Children))
		for i := range n.Children {
			cp.Children[i] = cloneNode(n.Children[i])
		}
	}

	return cp
}

func collectIndexes[T any, K comparable](
	roots []*Node[T],
	keyFn func(T) K,
	cache map[K]*Node[T],
	parentIdx map[K]K,
) {
	var collect func(*Node[T], K, bool)
	collect = func(n *Node[T], parentKey K, hasParent bool) {
		k := keyFn(n.Item)
		cache[k] = n
		if hasParent {
			parentIdx[k] = parentKey
		}
		for _, child := range n.Children {
			collect(child, k, true)
		}
	}

	var zero K
	for _, root := range roots {
		collect(root, zero, false)
	}
}
