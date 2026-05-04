package tree

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
