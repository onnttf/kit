package tree

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidInput  = errors.New("tree: invalid input")
	ErrKeyNotSet     = errors.New("tree: key function is not set")
	ErrDuplicateKey  = errors.New("tree: duplicate key")
	ErrMissingParent = errors.New("tree: missing parent")
	ErrCycle         = errors.New("tree: cycle detected")
	ErrKeyNotFound   = errors.New("tree: key not found")
)

type Node[T any, K comparable] struct {
	Key      K
	Parent   *K
	Value    T
	Children []Node[T, K]
	Depth    int
}

type Tree[T any, K comparable] struct {
	roots []Node[T, K]
	index map[K]*Node[T, K]
}

func (t *Tree[T, K]) Get(key K) (Node[T, K], bool) {
	if t == nil {
		return Node[T, K]{}, false
	}
	n, ok := t.index[key]
	if !ok {
		return Node[T, K]{}, false
	}
	return cloneNode(n), true
}

func (t *Tree[T, K]) Roots() []Node[T, K] {
	if t == nil {
		return nil
	}
	return cloneNodes(t.roots)
}

func (t *Tree[T, K]) Children(key K) ([]Node[T, K], bool) {
	if t == nil {
		return nil, false
	}
	n, ok := t.index[key]
	if !ok {
		return nil, false
	}
	return cloneNodes(n.Children), true
}

func (t *Tree[T, K]) Walk(fn func(node Node[T, K], parent *Node[T, K]) bool) (bool, error) {
	if fn == nil {
		return false, fmt.Errorf("%w: nil walk function", ErrInvalidInput)
	}
	if t == nil {
		return true, nil
	}
	var walk func(nodes []Node[T, K], parent *Node[T, K]) bool
	walk = func(nodes []Node[T, K], parent *Node[T, K]) bool {
		for i := range nodes {
			node := cloneNode(&nodes[i])
			var parentCopy *Node[T, K]
			if parent != nil {
				cp := cloneNode(parent)
				parentCopy = &cp
			}
			if !fn(node, parentCopy) {
				return false
			}
			if !walk(nodes[i].Children, &nodes[i]) {
				return false
			}
		}
		return true
	}
	return walk(t.roots, nil), nil
}

func newTree[T any, K comparable](roots []Node[T, K]) *Tree[T, K] {
	t := &Tree[T, K]{
		roots: roots,
		index: make(map[K]*Node[T, K]),
	}
	var collect func(nodes []Node[T, K])
	collect = func(nodes []Node[T, K]) {
		for i := range nodes {
			n := &nodes[i]
			t.index[n.Key] = n
			collect(n.Children)
		}
	}
	collect(t.roots)
	return t
}

func cloneNode[T any, K comparable](n *Node[T, K]) Node[T, K] {
	if n == nil {
		return Node[T, K]{}
	}
	out := Node[T, K]{
		Key:    n.Key,
		Value:  n.Value,
		Depth:  n.Depth,
		Parent: cloneKey(n.Parent),
	}
	out.Children = make([]Node[T, K], len(n.Children))
	for i := range n.Children {
		child := cloneNode(&n.Children[i])
		out.Children[i] = child
	}
	return out
}

func cloneNodes[T any, K comparable](nodes []Node[T, K]) []Node[T, K] {
	out := make([]Node[T, K], len(nodes))
	for i := range nodes {
		out[i] = cloneNode(&nodes[i])
	}
	return out
}

func cloneKey[K comparable](p *K) *K {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}
