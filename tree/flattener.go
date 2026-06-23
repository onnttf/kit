package tree

import "slices"

type Flattener[T any, K comparable] struct {
	keyFn    func(T) K
	parentFn func(T, K) T
}

func NewFlattener[T any, K comparable]() *Flattener[T, K] {
	return &Flattener[T, K]{}
}

func (f *Flattener[T, K]) KeyBy(fn func(T) K) *Flattener[T, K] {
	f.keyFn = fn
	return f
}

func (f *Flattener[T, K]) ParentBy(fn func(T, K) T) *Flattener[T, K] {
	f.parentFn = fn
	return f
}

func (f *Flattener[T, K]) Flatten(roots []*Node[T]) ([]T, error) {
	if f.keyFn == nil {
		return nil, ErrKeyNotSet
	}

	result := make([]T, 0, len(roots)*4)

	type entry struct {
		node      *Node[T]
		parentKey K
		hasParent bool
	}

	stack := make([]entry, 0, len(roots))
	for _, root := range slices.Backward(roots) {
		stack = append(stack, entry{node: root})
	}

	for len(stack) > 0 {
		e := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		item := e.node.Item
		if f.parentFn != nil && e.hasParent {
			item = f.parentFn(item, e.parentKey)
		}
		result = append(result, item)

		nodeKey := f.keyFn(e.node.Item)
		for _, child := range slices.Backward(e.node.Children) {
			stack = append(stack, entry{
				node:      child,
				parentKey: nodeKey,
				hasParent: true,
			})
		}
	}

	return result, nil
}
