package tree

// Stats holds tree statistics.
type Stats struct {
	TotalNodes  int
	RootNodes   int
	MaxDepth    int
	AvgDepth    float64
	LeafNodes   int
	AvgChildren float64
}

// Tree represents an immutable tree structure.
type Tree[T any, K comparable] struct {
	roots      []*Node[T]
	cache      map[K]*Node[T]
	parentIdx  map[K]K
	keyFn      func(T) K
	subtreeIdx map[K]*Tree[T, K]
}

// Len returns the number of nodes in the tree.
func (t *Tree[T, K]) Len() int { return len(t.cache) }

// Empty returns true if the tree has no nodes.
func (t *Tree[T, K]) Empty() bool { return len(t.cache) == 0 }

// ContainsKey returns true if the key exists in the tree.
func (t *Tree[T, K]) ContainsKey(key K) bool {
	_, ok := t.cache[key]
	return ok
}

// Get returns a cloned node by key.
func (t *Tree[T, K]) Get(key K) (*Node[T], bool) {
	n, ok := t.cache[key]
	if !ok {
		return nil, false
	}
	return cloneNode(n), true
}

// Roots returns all root nodes.
func (t *Tree[T, K]) Roots() []*Node[T] {
	out := make([]*Node[T], len(t.roots))
	for i, n := range t.roots {
		out[i] = cloneNode(n)
	}
	return out
}

// ParentOf returns the parent key of a node.
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

// Children returns the direct children of a node.
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

// Orphans returns root nodes that have children but no parent.
func (t *Tree[T, K]) Orphans() []*Node[T] {
	if t.parentIdx == nil {
		return nil
	}

	var orphans []*Node[T]
	for k, n := range t.cache {
		if _, hasParent := t.parentIdx[k]; !hasParent && len(n.Children) > 0 {
			orphans = append(orphans, cloneNode(n))
		}
	}

	return orphans
}

// LeafNodes returns all leaf nodes.
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

// Ancestors returns all ancestors of a node.
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

// Walk traverses the tree in depth-first pre-order.
// fn receives (node, parent). parent is nil for root nodes.
// The tree is expected to already have Level values assigned.
func (t *Tree[T, K]) Walk(fn func(*Node[T], *Node[T]) bool) bool {
	if fn == nil {
		return true
	}

	type entry struct {
		node   *Node[T]
		parent *Node[T]
	}

	stack := make([]entry, 0, len(t.roots))
	for i := len(t.roots) - 1; i >= 0; i-- {
		stack = append(stack, entry{node: t.roots[i]})
	}

	for len(stack) > 0 {
		e := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if !fn(e.node, e.parent) {
			return false
		}

		for i := len(e.node.Children) - 1; i >= 0; i-- {
			stack = append(stack, entry{
				node:   e.node.Children[i],
				parent: e.node,
			})
		}
	}

	return true
}

// Filter returns a new tree with nodes matching the predicate.
func (t *Tree[T, K]) Filter(fn func(*Node[T]) bool) *Tree[T, K] {
	var filteredRoots []*Node[T]
	var filteredCache map[K]*Node[T] = make(map[K]*Node[T])

	var filterNode func(*Node[T])
	filterNode = func(n *Node[T]) {
		if !fn(n) {
			return
		}
		k := t.keyFn(n.Item)
		newNode := cloneNode(n)
		filteredCache[k] = newNode
		filteredRoots = append(filteredRoots, newNode)

		for _, c := range n.Children {
			filterNode(c)
		}
	}

	for _, r := range t.roots {
		filterNode(r)
	}

	assignLevels(filteredRoots, 1)

	return &Tree[T, K]{
		roots:      filteredRoots,
		cache:      filteredCache,
		parentIdx:  t.parentIdx,
		keyFn:      t.keyFn,
		subtreeIdx: make(map[K]*Tree[T, K]),
	}
}

// Map returns a new tree with transformed items.
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
	for i, r := range newRoots {
		k := keyFn(r.Item)
		newCache[k] = r

		var collectCache func(*Node[T])
		collectCache = func(n *Node[T]) {
			k := keyFn(n.Item)
			newCache[k] = n
			for _, c := range n.Children {
				collectCache(c)
			}
		}
		collectCache(newRoots[i])
	}

	return &Tree[T, K]{
		roots:      newRoots,
		cache:      newCache,
		parentIdx:  t.parentIdx,
		keyFn:      keyFn,
		subtreeIdx: make(map[K]*Tree[T, K]),
	}
}

// Clone returns a deep copy of the tree.
func (t *Tree[T, K]) Clone() *Tree[T, K] {
	roots := make([]*Node[T], len(t.roots))
	cache := make(map[K]*Node[T], len(t.cache))
	parentIdx := make(map[K]K, len(t.parentIdx))

	for i, r := range t.roots {
		roots[i] = cloneNode(r)
	}

	var collect func(*Node[T], K)
	collect = func(n *Node[T], pk K) {
		k := t.keyFn(n.Item)
		cache[k] = n
		if pk != *new(K) {
			parentIdx[k] = pk
		}
		for _, c := range n.Children {
			collect(c, k)
		}
	}

	var zeroK K
	for _, r := range t.roots {
		collect(r, zeroK)
	}

	return &Tree[T, K]{
		roots:      roots,
		cache:      cache,
		parentIdx:  parentIdx,
		keyFn:      t.keyFn,
		subtreeIdx: make(map[K]*Tree[T, K]),
	}
}

// Subtree returns a subtree rooted at the given key.
func (t *Tree[T, K]) Subtree(key K) (*Tree[T, K], bool) {
	if _, ok := t.cache[key]; !ok {
		return nil, false
	}

	if st, ok := t.subtreeIdx[key]; ok {
		return st, true
	}

	root := t.cache[key]
	rootCopy := cloneNode(root)

	roots := []*Node[T]{rootCopy}

	assignLevels(roots, 1)

	cache := make(map[K]*Node[T])
	parentIdx := make(map[K]K)

	var buildCache func(*Node[T], K)
	buildCache = func(n *Node[T], parentKey K) {
		k := t.keyFn(n.Item)
		cache[k] = n
		if parentKey != *new(K) {
			parentIdx[k] = parentKey
		}
		for _, c := range n.Children {
			buildCache(c, k)
		}
	}

	var zeroK K
	buildCache(rootCopy, zeroK)

	subtree := &Tree[T, K]{
		roots:      roots,
		cache:      cache,
		parentIdx:  parentIdx,
		keyFn:      t.keyFn,
		subtreeIdx: make(map[K]*Tree[T, K]),
	}

	t.subtreeIdx[key] = subtree
	return subtree, true
}

// ToMap returns all items as a map.
func (t *Tree[T, K]) ToMap() map[K]T {
	result := make(map[K]T, len(t.cache))
	for k, n := range t.cache {
		result[k] = n.Item
	}
	return result
}

// Stats returns tree statistics.
func (t *Tree[T, K]) Stats() Stats {
	if len(t.cache) == 0 {
		return Stats{}
	}

	var total, leaves, maxDepth, totalDepth int
	rootCount := len(t.roots)

	var walk func(*Node[T])
	walk = func(n *Node[T]) {
		total++
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
	if total > rootCount {
		avgChildren = float64(total-rootCount) / float64(rootCount)
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
