package tree

import "slices"

type Stats struct {
	TotalNodes  int
	RootNodes   int
	MaxDepth    int
	AvgDepth    float64
	LeafNodes   int
	AvgChildren float64
}

// Tree is a built tree indexed by key.
type Tree[T any, K comparable] struct {
	roots     []*Node[T]
	cache     map[K]*Node[T]
	parentIdx map[K]K
	keyFn     func(T) K
}

func (t *Tree[T, K]) Len() int { return len(t.cache) }

func (t *Tree[T, K]) Empty() bool { return len(t.cache) == 0 }

func (t *Tree[T, K]) ContainsKey(key K) bool {
	_, ok := t.cache[key]
	return ok
}

// Get returns a copy of the node for key.
func (t *Tree[T, K]) Get(key K) (*Node[T], bool) {
	n, ok := t.cache[key]
	if !ok {
		return nil, false
	}
	return cloneNode(n), true
}

// Roots returns copies of the root nodes.
func (t *Tree[T, K]) Roots() []*Node[T] {
	out := make([]*Node[T], len(t.roots))
	for i, n := range t.roots {
		out[i] = cloneNode(n)
	}
	return out
}

func (t *Tree[T, K]) ParentOf(key K) (K, bool) {
	if _, ok := t.cache[key]; !ok {
		var zero K
		return zero, false
	}
	if t.parentIdx == nil {
		var zero K
		return zero, false
	}
	pk, ok := t.parentIdx[key]
	return pk, ok
}

// Children returns copies of the direct children for key.
func (t *Tree[T, K]) Children(key K) ([]*Node[T], bool) {
	n, ok := t.cache[key]
	if !ok {
		return nil, false
	}
	children := make([]*Node[T], len(n.Children))
	for i, c := range n.Children {
		children[i] = cloneNode(c)
	}
	return children, true
}

// LeafNodes returns copies of all leaf nodes.
func (t *Tree[T, K]) LeafNodes() []*Node[T] {
	var leaves []*Node[T]
	t.Walk(func(n *Node[T], _ *Node[T]) bool {
		if len(n.Children) == 0 {
			leaves = append(leaves, cloneNode(n))
		}
		return true
	})
	return leaves
}

// Ancestors returns copies of ancestors from parent to root.
func (t *Tree[T, K]) Ancestors(key K) ([]*Node[T], bool) {
	if _, ok := t.cache[key]; !ok {
		return nil, false
	}

	var ancestors []*Node[T]
	cur := key
	for {
		pk, ok := t.parentIdx[cur]
		if !ok {
			break
		}
		if n, ok := t.cache[pk]; ok {
			ancestors = append(ancestors, cloneNode(n))
		}
		cur = pk
	}

	return ancestors, len(ancestors) > 0
}

// PathTo returns copies of nodes from a root to key.
func (t *Tree[T, K]) PathTo(key K) ([]*Node[T], bool) {
	n, ok := t.cache[key]
	if !ok {
		return nil, false
	}

	ancestors, _ := t.Ancestors(key)
	path := make([]*Node[T], 0, len(ancestors)+1)
	for _, ancestor := range slices.Backward(ancestors) {
		path = append(path, ancestor)
	}
	path = append(path, cloneNode(n))
	return path, true
}

// Descendants returns copies of all descendants in depth-first pre-order.
func (t *Tree[T, K]) Descendants(key K) ([]*Node[T], bool) {
	n, ok := t.cache[key]
	if !ok {
		return nil, false
	}

	descendants := make([]*Node[T], 0)
	stack := make([]*Node[T], 0, len(n.Children))
	for _, child := range slices.Backward(n.Children) {
		stack = append(stack, child)
	}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		descendants = append(descendants, cloneNode(cur))
		for _, child := range slices.Backward(cur.Children) {
			stack = append(stack, child)
		}
	}
	return descendants, true
}

// Walk visits nodes in depth-first pre-order until fn returns false.
func (t *Tree[T, K]) Walk(fn func(*Node[T], *Node[T]) bool) bool {
	if fn == nil {
		return true
	}

	type entry struct {
		node   *Node[T]
		parent *Node[T]
	}

	stack := make([]entry, 0, len(t.roots))
	for _, root := range slices.Backward(t.roots) {
		stack = append(stack, entry{node: root})
	}

	for len(stack) > 0 {
		e := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if !fn(e.node, e.parent) {
			return false
		}

		for _, child := range slices.Backward(e.node.Children) {
			stack = append(stack, entry{
				node:   child,
				parent: e.node,
			})
		}
	}

	return true
}

func (t *Tree[T, K]) Filter(fn func(*Node[T]) bool) *Tree[T, K] {
	var filteredRoots []*Node[T]
	filteredCache := make(map[K]*Node[T])
	filteredParentIdx := make(map[K]K)

	var filterNode func(*Node[T]) []*Node[T]
	filterNode = func(n *Node[T]) []*Node[T] {
		var children []*Node[T]
		for _, c := range n.Children {
			children = append(children, filterNode(c)...)
		}

		if !fn(n) {
			return children
		}

		newNode := &Node[T]{
			Item:     n.Item,
			Children: children,
			Level:    n.Level,
		}
		return []*Node[T]{newNode}
	}

	for _, r := range t.roots {
		filteredRoots = append(filteredRoots, filterNode(r)...)
	}

	assignLevels(filteredRoots, 1)
	collectIndexes(filteredRoots, t.keyFn, filteredCache, filteredParentIdx)

	return &Tree[T, K]{
		roots:     filteredRoots,
		cache:     filteredCache,
		parentIdx: filteredParentIdx,
		keyFn:     t.keyFn,
	}
}

func (t *Tree[T, K]) Map(fn func(T) T, keyFn func(T) K) *Tree[T, K] {
	newRoots := make([]*Node[T], len(t.roots))

	var mapNode func(*Node[T]) *Node[T]
	mapNode = func(n *Node[T]) *Node[T] {
		newItem := fn(n.Item)
		children := make([]*Node[T], len(n.Children))
		for i, c := range n.Children {
			children[i] = mapNode(c)
		}
		return &Node[T]{Item: newItem, Children: children, Level: n.Level}
	}

	for i, r := range t.roots {
		newRoots[i] = mapNode(r)
	}

	assignLevels(newRoots, 1)

	newCache := make(map[K]*Node[T], len(t.cache))
	newParentIdx := make(map[K]K, len(t.parentIdx))
	collectIndexes(newRoots, keyFn, newCache, newParentIdx)

	return &Tree[T, K]{
		roots:     newRoots,
		cache:     newCache,
		parentIdx: newParentIdx,
		keyFn:     keyFn,
	}
}

// Clone returns a deep copy of the tree structure.
func (t *Tree[T, K]) Clone() *Tree[T, K] {
	roots := make([]*Node[T], len(t.roots))
	cache := make(map[K]*Node[T], len(t.cache))
	parentIdx := make(map[K]K, len(t.parentIdx))

	for i, r := range t.roots {
		roots[i] = cloneNode(r)
	}

	var collect func(*Node[T], K, bool)
	collect = func(n *Node[T], pk K, hasParent bool) {
		k := t.keyFn(n.Item)
		cache[k] = n
		if hasParent {
			parentIdx[k] = pk
		}
		for _, c := range n.Children {
			collect(c, k, true)
		}
	}

	var zero K
	for _, r := range roots {
		collect(r, zero, false)
	}

	return &Tree[T, K]{
		roots:     roots,
		cache:     cache,
		parentIdx: parentIdx,
		keyFn:     t.keyFn,
	}
}

// Subtree returns a tree rooted at a copy of key's node.
func (t *Tree[T, K]) Subtree(key K) (*Tree[T, K], bool) {
	if _, ok := t.cache[key]; !ok {
		return nil, false
	}

	root := t.cache[key]
	rootCopy := cloneNode(root)

	roots := []*Node[T]{rootCopy}

	assignLevels(roots, 1)

	cache := make(map[K]*Node[T])
	parentIdx := make(map[K]K)

	collectIndexes(roots, t.keyFn, cache, parentIdx)

	subtree := &Tree[T, K]{
		roots:     roots,
		cache:     cache,
		parentIdx: parentIdx,
		keyFn:     t.keyFn,
	}

	return subtree, true
}

func (t *Tree[T, K]) ToMap() map[K]T {
	result := make(map[K]T, len(t.cache))
	for k, n := range t.cache {
		result[k] = n.Item
	}
	return result
}

func (t *Tree[T, K]) Stats() Stats {
	if len(t.cache) == 0 {
		return Stats{}
	}

	var total, leaves, maxDepth, totalDepth, totalChildren int
	rootCount := len(t.roots)

	var walk func(*Node[T])
	walk = func(n *Node[T]) {
		total++
		totalChildren += len(n.Children)
		if n.Level > maxDepth {
			maxDepth = n.Level
		}
		totalDepth += n.Level

		if len(n.Children) == 0 {
			leaves++
		}
		for _, c := range n.Children {
			walk(c)
		}
	}

	for _, r := range t.roots {
		walk(r)
	}

	avgChildren := 0.0
	if total > 0 {
		avgChildren = float64(totalChildren) / float64(total)
	}

	avgDepth := 0.0
	if total > 0 {
		avgDepth = float64(totalDepth) / float64(total)
	}

	return Stats{
		TotalNodes:  total,
		RootNodes:   rootCount,
		MaxDepth:    maxDepth,
		LeafNodes:   leaves,
		AvgChildren: avgChildren,
		AvgDepth:    avgDepth,
	}
}
