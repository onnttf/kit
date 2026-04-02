package tree

import (
	"errors"
	"fmt"
	"maps"
	"sort"
	"sync"
)

// Node is the output tree node. It holds the caller's item and its children.
// Build constructs Nodes; callers should not modify them.
//
// Example:
//
//	roots, _ := b.Build()
//	fmt.Println(roots[0].Item.Name)
//	for _, child := range roots[0].Children { ... }
type Node[T any, K comparable] struct {
	Item     T
	Children []*Node[T, K]
}

// node is the internal storage unit for a Builder entry. Not exported.
type node[T any, K comparable] struct {
	item        T
	key         K
	parentKey   K
	hasParent   bool // false when this is a root (ParentBy unset, or parentKey == key)
	sortVal     int
	insertOrder int // position assigned at insertion; used for stable ordering when SortBy is nil
}

// Builder builds a typed tree structure from arbitrary items using injected key
// extraction functions. It is safe for concurrent use.
//
// Example:
//
//	roots, nodeMap := tree.NewBuilder[Dept, int]().
//	    KeyBy(func(d Dept) int { return d.ID }).
//	    ParentBy(func(d Dept) int { return d.ParentID }).
//	    SortBy(func(d Dept) int { return d.Sort }).
//	    WithItems(depts).
//	    Build()
type Builder[T any, K comparable] struct {
	mu    sync.RWMutex
	items []*node[T, K]     // insertion-ordered; primary data source
	index map[K]*node[T, K] // fast lookup by key

	insertCtr int

	keyFn    func(T) K
	parentFn func(T) K
	sortFn   func(T) int

	// parentOverrides stores parent key overrides written by MoveItem.
	// MoveItem cannot mutate T directly, so overrides are tracked here.
	parentOverrides map[K]K

	dirty     bool
	roots     []*Node[T, K]     // cached by buildTree
	nodeCache map[K]*Node[T, K] // cached by buildTree
}

// NewBuilder returns a new Builder ready for configuration.
//
// Example:
//
//	b := tree.NewBuilder[Dept, int]()
func NewBuilder[T any, K comparable]() *Builder[T, K] {
	return &Builder[T, K]{
		items: make([]*node[T, K], 0),
		index: make(map[K]*node[T, K]),
		dirty: true,
	}
}

// KeyBy sets the function used to extract the unique key from each item.
// Passing nil is a no-op.
//
// Example:
//
//	b.KeyBy(func(d Dept) int { return d.ID })
func (b *Builder[T, K]) KeyBy(fn func(T) K) *Builder[T, K] {
	if fn == nil {
		return b
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.keyFn = fn
	b.dirty = true
	return b
}

// ParentBy sets the function used to extract the parent key from each item.
// When not set, all items are treated as root nodes. Passing nil is a no-op.
//
// Example:
//
//	b.ParentBy(func(d Dept) int { return d.ParentID })
func (b *Builder[T, K]) ParentBy(fn func(T) K) *Builder[T, K] {
	if fn == nil {
		return b
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.parentFn = fn
	b.dirty = true
	return b
}

// SortBy sets the function used to determine sibling sort order.
// When not set, siblings retain their insertion order. Passing nil is a no-op.
//
// Example:
//
//	b.SortBy(func(d Dept) int { return d.Sort })
func (b *Builder[T, K]) SortBy(fn func(T) int) *Builder[T, K] {
	if fn == nil {
		return b
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sortFn = fn
	b.dirty = true
	return b
}

// ensureBuiltLocked rebuilds the tree if dirty. Caller must hold mu.
func (b *Builder[T, K]) ensureBuiltLocked() {
	if b.dirty {
		b.buildTree()
		b.dirty = false
	}
}

// buildTree recomputes node metadata, constructs output Nodes,
// links parent-child relationships, and sorts siblings.
// Caller must hold mu (write lock).
func (b *Builder[T, K]) buildTree() {
	// Phase 1: recompute metadata and rebuild index
	b.index = make(map[K]*node[T, K], len(b.items))
	for _, n := range b.items {
		if b.keyFn != nil {
			n.key = b.keyFn(n.item)
			b.index[n.key] = n
		}
		if b.parentFn != nil && b.keyFn != nil {
			n.parentKey = b.parentFn(n.item)
			n.hasParent = n.parentKey != n.key
		} else {
			var zero K
			n.parentKey = zero
			n.hasParent = false
		}
		if b.sortFn != nil {
			n.sortVal = b.sortFn(n.item)
		} else {
			n.sortVal = n.insertOrder
		}
	}

	// Phase 2: construct output Node wrappers
	b.nodeCache = make(map[K]*Node[T, K], len(b.items))
	for _, n := range b.items {
		if b.keyFn != nil {
			b.nodeCache[n.key] = &Node[T, K]{Item: n.item}
		}
	}

	// Phase 3: link parent-child relationships
	b.roots = nil
	for _, n := range b.items {
		if b.keyFn == nil {
			continue
		}
		outNode := b.nodeCache[n.key]

		effectiveParent, isRoot := b.effectiveParent(n)
		if isRoot {
			b.roots = append(b.roots, outNode)
		} else if parent, ok := b.nodeCache[effectiveParent]; ok {
			parent.Children = append(parent.Children, outNode)
		}
		// else: orphan — not reachable from any root
	}

	// Phase 4: sort siblings when SortBy is set.
	// When SortBy is nil, insertion order is preserved naturally.
	if b.sortFn != nil {
		b.sortNodes(b.roots)
	}
}

// effectiveParent returns the effective parent key for n, accounting for
// parentOverrides. Reports isRoot=true when the node should be a root.
// Caller must hold mu.
func (b *Builder[T, K]) effectiveParent(n *node[T, K]) (key K, isRoot bool) {
	if b.parentOverrides != nil {
		if pk, ok := b.parentOverrides[n.key]; ok {
			return pk, false
		}
	}
	if n.hasParent {
		return n.parentKey, false
	}
	var zero K
	return zero, true
}

// sortNodes sorts nodes and their descendants by sortVal using a stable sort.
// Uses an iterative post-order traversal to avoid stack overflow on deep trees.
// Caller must hold mu.
func (b *Builder[T, K]) sortNodes(roots []*Node[T, K]) {
	// Collect all node slices that need sorting via iterative DFS.
	type frame struct{ children []*Node[T, K] }
	stack := []frame{{roots}}
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if len(top.children) <= 1 {
			continue
		}
		sort.SliceStable(top.children, func(i, j int) bool {
			ki := b.keyFn(top.children[i].Item)
			kj := b.keyFn(top.children[j].Item)
			return b.index[ki].sortVal < b.index[kj].sortVal
		})
		for _, n := range top.children {
			if len(n.Children) > 1 {
				stack = append(stack, frame{n.Children})
			}
		}
	}
}

// maxDepth returns the maximum depth of the tree using iterative DFS.
// Caller must hold mu and have called ensureBuiltLocked.
func (b *Builder[T, K]) maxDepth() int {
	if len(b.roots) == 0 {
		return 0
	}
	type entry struct {
		n     *Node[T, K]
		depth int
	}
	stack := make([]entry, 0, len(b.roots))
	for _, r := range b.roots {
		stack = append(stack, entry{r, 1})
	}
	max := 0
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if cur.depth > max {
			max = cur.depth
		}
		for _, child := range cur.n.Children {
			stack = append(stack, entry{child, cur.depth + 1})
		}
	}
	return max
}

// leafCount returns the number of nodes with no children.
// Caller must hold mu and have called ensureBuiltLocked.
func (b *Builder[T, K]) leafCount() int {
	count := 0
	for _, n := range b.nodeCache {
		if len(n.Children) == 0 {
			count++
		}
	}
	return count
}

// isDescendant reports whether potentialDesc is a descendant of ancestorKey,
// including the case where potentialDesc == ancestorKey (a node is considered
// a descendant of itself, which prevents self-moves in MoveItem).
// Caller must hold mu.
func (b *Builder[T, K]) isDescendant(ancestorKey, potentialDesc K) bool {
	if ancestorKey == potentialDesc {
		return true
	}
	visited := make(map[K]bool)
	cur := potentialDesc
	for {
		if visited[cur] {
			return false
		}
		visited[cur] = true

		n := b.index[cur]
		if n == nil {
			return false
		}
		effectiveParent, isRoot := b.effectiveParent(n)
		if isRoot {
			return false
		}
		if effectiveParent == ancestorKey {
			return true
		}
		cur = effectiveParent
	}
}

// WithItems replaces all current items with the provided slice.
// Passing nil clears all items.
//
// Example:
//
//	b.WithItems(depts)
func (b *Builder[T, K]) WithItems(items []T) *Builder[T, K] {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.items = make([]*node[T, K], 0, len(items))
	b.index = make(map[K]*node[T, K], len(items))
	b.parentOverrides = nil
	b.insertCtr = 0
	b.dirty = true

	for _, item := range items {
		n := &node[T, K]{item: item, insertOrder: b.insertCtr}
		b.insertCtr++
		b.items = append(b.items, n)
		if b.keyFn != nil {
			n.key = b.keyFn(item)
			b.index[n.key] = n
		}
	}
	return b
}

// AddItem adds a single item to the builder.
//
// Example:
//
//	b.AddItem(Dept{ID: 3, ParentID: 1, Name: "HR"})
func (b *Builder[T, K]) AddItem(item T) *Builder[T, K] {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := &node[T, K]{item: item, insertOrder: b.insertCtr}
	b.insertCtr++
	b.items = append(b.items, n)
	if b.keyFn != nil {
		n.key = b.keyFn(item)
		b.index[n.key] = n
	}
	b.dirty = true
	return b
}

// RemoveItem removes the item with the given key and all its descendants.
// If the key does not exist, this is a no-op.
//
// Example:
//
//	b.RemoveItem(2) // removes node 2 and all its children
func (b *Builder[T, K]) RemoveItem(key K) *Builder[T, K] {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.ensureBuiltLocked()

	cached, ok := b.nodeCache[key]
	if !ok {
		return b
	}

	// Collect the key and all descendant keys via iterative DFS.
	toRemove := make(map[K]bool)
	stack := []*Node[T, K]{cached}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		k := b.keyFn(cur.Item)
		toRemove[k] = true
		stack = append(stack, cur.Children...)
	}

	// Remove from index and overrides.
	for k := range toRemove {
		delete(b.index, k)
		if b.parentOverrides != nil {
			delete(b.parentOverrides, k)
		}
	}

	// Compact items slice. Nil out the tail so removed *node pointers can be GC'd.
	kept := b.items[:0]
	for _, n := range b.items {
		if !toRemove[n.key] {
			kept = append(kept, n)
		}
	}
	clear(b.items[len(kept):])
	b.items = kept
	b.dirty = true
	return b
}

// MoveItem moves the item with the given key under a new parent.
// Self-moves, moves to a non-existent parent, and moves that would create
// cycles are silently ignored.
// MoveItem does not modify the item's fields; it stores an internal override.
//
// Example:
//
//	b.MoveItem(5, 2) // move item with key 5 under parent with key 2
func (b *Builder[T, K]) MoveItem(key, newParentKey K) *Builder[T, K] {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.ensureBuiltLocked() // populate hasParent/parentKey before cycle detection

	if _, ok := b.index[key]; !ok {
		return b
	}
	if _, ok := b.index[newParentKey]; !ok {
		return b
	}
	if key == newParentKey {
		return b
	}
	if b.isDescendant(key, newParentKey) {
		return b
	}

	if b.parentOverrides == nil {
		b.parentOverrides = make(map[K]K)
	}
	b.parentOverrides[key] = newParentKey
	b.dirty = true
	return b
}

// UpdateItem applies fn to the item identified by key.
// If the key function returns a new key after fn runs, the index is updated.
// Passing nil fn is a no-op.
//
// Example:
//
//	b.UpdateItem(1, func(d *Dept) { d.Name = "Engineering" })
func (b *Builder[T, K]) UpdateItem(key K, fn func(*T)) *Builder[T, K] {
	if fn == nil {
		return b
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	n := b.index[key]
	if n == nil {
		return b
	}

	fn(&n.item)

	if b.keyFn != nil {
		newKey := b.keyFn(n.item)
		if newKey != n.key {
			delete(b.index, n.key)
			n.key = newKey
			b.index[newKey] = n
		}
	}
	b.dirty = true
	return b
}

// Filter returns a new Builder containing only items for which predicate returns true.
// The new Builder inherits the key extraction functions from the original.
// Passing nil predicate retains all items.
//
// Example:
//
//	active := b.Filter(func(d Dept) bool { return d.Active })
func (b *Builder[T, K]) Filter(predicate func(T) bool) *Builder[T, K] {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.ensureBuiltLocked()

	nb := &Builder[T, K]{
		keyFn:    b.keyFn,
		parentFn: b.parentFn,
		sortFn:   b.sortFn,
		items:    make([]*node[T, K], 0),
		index:    make(map[K]*node[T, K]),
		dirty:    true,
	}

	stack := make([]*Node[T, K], 0, len(b.roots))
	stack = append(stack, b.roots...)

	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if predicate == nil || predicate(cur.Item) {
			if b.keyFn != nil {
				k := b.keyFn(cur.Item)
				if orig := b.index[k]; orig != nil {
					copied := *orig
					nb.items = append(nb.items, &copied)
					nb.index[k] = &copied
					if b.parentOverrides != nil {
						if pk, ok := b.parentOverrides[k]; ok {
							if nb.parentOverrides == nil {
								nb.parentOverrides = make(map[K]K)
							}
							nb.parentOverrides[k] = pk
						}
					}
				}
			}
		}

		for i := len(cur.Children) - 1; i >= 0; i-- {
			stack = append(stack, cur.Children[i])
		}
	}

	nb.insertCtr = len(nb.items)
	return nb
}

// Transform applies fn to every item in the builder in place.
// Passing nil fn is a no-op.
//
// Example:
//
//	b.Transform(func(d *Dept) { d.Name = strings.ToUpper(d.Name) })
func (b *Builder[T, K]) Transform(fn func(*T)) *Builder[T, K] {
	if fn == nil {
		return b
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, n := range b.items {
		fn(&n.item)
	}
	b.dirty = true
	return b
}

// Build constructs and returns the tree. Returns root nodes in sort order and a
// map for direct key lookup. Returns nil, nil if KeyBy is not set.
// The returned roots slice and nodeMap must not be modified by the caller.
//
// Example:
//
//	roots, nodeMap := b.Build()
func (b *Builder[T, K]) Build() ([]*Node[T, K], map[K]*Node[T, K]) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.keyFn == nil {
		return nil, nil
	}
	b.ensureBuiltLocked()
	return b.roots, b.nodeCache
}

// Clone returns a new Builder with an independent deep copy of all items.
// Key extraction functions are shared (not copied) between original and clone.
//
// Example:
//
//	copy := b.Clone()
func (b *Builder[T, K]) Clone() *Builder[T, K] {
	b.mu.RLock()
	defer b.mu.RUnlock()

	nb := &Builder[T, K]{
		keyFn:     b.keyFn,
		parentFn:  b.parentFn,
		sortFn:    b.sortFn,
		insertCtr: b.insertCtr,
		items:     make([]*node[T, K], len(b.items)),
		index:     make(map[K]*node[T, K], len(b.index)),
		dirty:     true,
	}

	for i, n := range b.items {
		copied := *n
		nb.items[i] = &copied
		if b.keyFn != nil {
			nb.index[copied.key] = nb.items[i]
		}
	}

	if len(b.parentOverrides) > 0 {
		nb.parentOverrides = make(map[K]K, len(b.parentOverrides))
		maps.Copy(nb.parentOverrides, b.parentOverrides)
	}

	return nb
}

// Validate checks the item set for structural problems. It returns errors for
// any cycles or orphaned nodes (nodes whose parent is not in the set).
// Returns a single error if KeyBy has not been set.
//
// Example:
//
//	if errs := b.Validate(); len(errs) != 0 { log.Fatal(errs) }
func (b *Builder[T, K]) Validate() []error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.keyFn == nil {
		return []error{errors.New("key function not set")}
	}

	b.ensureBuiltLocked()

	var errs []error

	// Orphan check: nodes whose effective parent is not in the built set.
	for k := range b.nodeCache {
		n := b.index[k]
		pk, isRoot := b.effectiveParent(n)
		if isRoot {
			continue
		}
		if _, exists := b.nodeCache[pk]; !exists {
			errs = append(errs, fmt.Errorf("orphaned node %v", k))
		}
	}

	// Cycle detection via upward DFS over all nodes in nodeCache.
	// Nodes in a cycle never appear in roots, so we must start from every node.
	type state uint8
	const (
		unvisited state = iota
		inProgress
		done
	)
	visited := make(map[K]state, len(b.nodeCache))

	var visit func(k K) bool
	visit = func(k K) bool {
		switch visited[k] {
		case done:
			return false
		case inProgress:
			return true // back-edge: cycle detected
		}
		visited[k] = inProgress
		if n := b.index[k]; n != nil {
			pk, isRoot := b.effectiveParent(n)
			if !isRoot {
				if _, parentExists := b.nodeCache[pk]; parentExists {
					if visit(pk) {
						visited[k] = done
						return true
					}
				}
			}
		}
		visited[k] = done
		return false
	}

	for k := range b.nodeCache {
		if visited[k] == unvisited {
			if visit(k) {
				errs = append(errs, fmt.Errorf("cycle detected at node %v", k))
			}
		}
	}

	return errs
}

// Statistics returns aggregate metrics about the built tree.
// Keys: "totalNodes", "rootNodes", "maxDepth", "leafNodes", "avgChildrenPerNode".
//
// Example:
//
//	stats := b.Statistics()
//	fmt.Println(stats["totalNodes"], stats["maxDepth"])
func (b *Builder[T, K]) Statistics() map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.keyFn == nil {
		return map[string]any{
			"totalNodes":         0,
			"rootNodes":          0,
			"maxDepth":           0,
			"leafNodes":          0,
			"avgChildrenPerNode": 0.0,
		}
	}

	b.ensureBuiltLocked()

	total := len(b.nodeCache)
	rootCount := len(b.roots)
	depth := b.maxDepth()
	leaves := b.leafCount()

	var avg float64
	if total > 0 {
		// Each non-root node is a child of exactly one parent.
		avg = float64(total-rootCount) / float64(total)
	}

	return map[string]any{
		"totalNodes":         total,
		"rootNodes":          rootCount,
		"maxDepth":           depth,
		"leafNodes":          leaves,
		"avgChildrenPerNode": avg,
	}
}
