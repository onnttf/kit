package tree

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync"
)

var (
	ErrDuplicateKey = errors.New("duplicate key")
	ErrKeyNotSet    = errors.New("key function not set")
	ErrOrphanedNode = errors.New("orphaned node")
	ErrCycle        = errors.New("cycle detected")
)

// Node is the output tree node. It holds the caller's item and its children.
// Build constructs Nodes; callers should not modify them.
//
// Example:
//
//	roots, _ := b.Build()
//	fmt.Println(roots[0].Item.Name)
//	for _, child := range roots[0].Children { ... }
type Node[T any] struct {
	Item     T
	Children []*Node[T]
}

// node is the internal storage unit for a Builder entry. Not exported.
type node[T any] struct {
	item        T
	key         any
	parentKey   any
	hasParent   bool
	sortVal     int
	insertOrder int
}

// Builder builds a typed tree structure from arbitrary items. It is safe for
// concurrent use. Keys are provided by the caller via KeyBy.
//
// Example:
//
//	roots, nodeMap := tree.NewBuilder[Dept]().
//	    KeyBy(func(d Dept) int { return d.ID }).
//	    ParentBy(func(d Dept) int { return d.ParentID }).
//	    SortBy(func(d Dept) int { return d.Sort }).
//	    WithItems(depts).
//	    Build()
type Builder[T any] struct {
	mu    sync.RWMutex
	items []*node[T]
	index map[any]*node[T]

	insertCtr int

	keyFn    func(T) any
	parentFn func(T) any
	sortFn   func(T) int

	parentOverrides map[any]any

	dirty     bool
	roots     []*Node[T]
	nodeCache map[any]*Node[T]
}

// NewBuilder returns a new Builder ready for configuration.
//
// Example:
//
//	b := tree.NewBuilder[Dept]()
func NewBuilder[T any]() *Builder[T] {
	return &Builder[T]{
		items: make([]*node[T], 0),
		index: make(map[any]*node[T]),
		dirty: true,
	}
}

// KeyBy sets the function used to extract the unique key from each item.
// Passing nil clears the key function.
//
// Example:
//
//	b.KeyBy(func(d Dept) int { return d.ID })
func (b *Builder[T]) KeyBy(fn func(T) any) *Builder[T] {
	if fn == nil {
		b.mu.Lock()
		defer b.mu.Unlock()
		b.keyFn = nil
		b.dirty = true
		return b
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.keyFn = fn
	b.dirty = true
	return b
}

// ParentBy sets the function used to extract the parent key from each item.
// When not set, all items are treated as root nodes. Passing nil clears the
// parent function.
//
// Example:
//
//	b.ParentBy(func(d Dept) int { return d.ParentID })
func (b *Builder[T]) ParentBy(fn func(T) any) *Builder[T] {
	if fn == nil {
		b.mu.Lock()
		defer b.mu.Unlock()
		b.parentFn = nil
		b.dirty = true
		return b
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.parentFn = fn
	b.dirty = true
	return b
}

// SortBy sets the function used to determine sibling sort order.
// When not set, siblings retain their insertion order. Passing nil clears the
// sort function.
//
// Example:
//
//	b.SortBy(func(d Dept) int { return d.Sort })
func (b *Builder[T]) SortBy(fn func(T) int) *Builder[T] {
	if fn == nil {
		b.mu.Lock()
		defer b.mu.Unlock()
		b.sortFn = nil
		b.dirty = true
		return b
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sortFn = fn
	b.dirty = true
	return b
}

// ensureBuiltLocked rebuilds the tree if dirty. Caller must hold mu.
func (b *Builder[T]) ensureBuiltLocked() error {
	if b.dirty {
		if err := b.buildTree(); err != nil {
			return err
		}
		b.dirty = false
	}
	return nil
}

// buildTree recomputes node metadata, constructs output Nodes,
// links parent-child relationships, and sorts siblings.
// Returns error on duplicate key.
// Caller must hold mu (write lock).
func (b *Builder[T]) buildTree() error {
	b.index = make(map[any]*node[T], len(b.items))
	for _, n := range b.items {
		if b.keyFn != nil {
			n.key = b.keyFn(n.item)
			if _, exists := b.index[n.key]; exists {
				return ErrDuplicateKey
			}
			b.index[n.key] = n
		}
		if b.parentFn != nil && b.keyFn != nil {
			n.parentKey = b.parentFn(n.item)
			n.hasParent = n.parentKey != n.key
		} else {
			n.parentKey = nil
			n.hasParent = false
		}
		if b.sortFn != nil {
			n.sortVal = b.sortFn(n.item)
		} else {
			n.sortVal = n.insertOrder
		}
	}

	// Phase 2: construct output Node wrappers
	b.nodeCache = make(map[any]*Node[T], len(b.items))
	for _, n := range b.items {
		if b.keyFn != nil {
			b.nodeCache[n.key] = &Node[T]{Item: n.item}
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
	}

	// Phase 4: sort siblings when SortBy is set.
	if b.sortFn != nil {
		b.sortNodes(b.roots)
	}
	return nil
}

// effectiveParent returns the effective parent key for n, accounting for
// parentOverrides. Reports isRoot=true when the node should be a root.
// Caller must hold mu.
func (b *Builder[T]) effectiveParent(n *node[T]) (key any, isRoot bool) {
	if b.parentOverrides != nil {
		if pk, ok := b.parentOverrides[n.key]; ok {
			return pk, false
		}
	}
	if n.hasParent {
		return n.parentKey, false
	}
	return nil, true
}

// sortNodes sorts nodes and their descendants by sortVal using a stable sort.
// Uses an iterative post-order traversal to avoid stack overflow on deep trees.
// Caller must hold mu.
func (b *Builder[T]) sortNodes(roots []*Node[T]) {
	type frame struct{ children []*Node[T] }
	stack := make([]frame, 0, len(roots)*4)
	stack = append(stack, frame{roots})
	keyFn := b.keyFn
	index := b.index
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if len(top.children) <= 1 {
			continue
		}
		slices.SortFunc(top.children, func(a, b *Node[T]) int {
			ka := keyFn(a.Item)
			kb := keyFn(b.Item)
			return cmp.Compare(index[ka].sortVal, index[kb].sortVal)
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
func (b *Builder[T]) maxDepth() int {
	if len(b.roots) == 0 {
		return 0
	}
	type entry struct {
		n     *Node[T]
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
func (b *Builder[T]) leafCount() int {
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
func (b *Builder[T]) isDescendant(ancestorKey, potentialDesc any) bool {
	if ancestorKey == potentialDesc {
		return true
	}
	visited := make(map[any]struct{})
	cur := potentialDesc
	for {
		if _, ok := visited[cur]; ok {
			return false
		}
		visited[cur] = struct{}{}

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
func (b *Builder[T]) WithItems(items []T) *Builder[T] {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.items = make([]*node[T], 0, len(items))
	b.index = make(map[any]*node[T], len(items))
	b.parentOverrides = nil
	b.insertCtr = 0
	b.dirty = true

	for _, item := range items {
		n := &node[T]{item: item, insertOrder: b.insertCtr}
		b.insertCtr++
		if b.keyFn != nil {
			n.key = b.keyFn(item)
			b.index[n.key] = n
		}
		b.items = append(b.items, n)
	}
	return b
}

// AddItem adds a single item to the builder.
//
// Example:
//
//	b.AddItem(Dept{ID: 3, ParentID: 1, Name: "HR"})
func (b *Builder[T]) AddItem(item T) *Builder[T] {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := &node[T]{item: item, insertOrder: b.insertCtr}
	b.insertCtr++
	if b.keyFn != nil {
		n.key = b.keyFn(item)
		b.index[n.key] = n
	}
	b.items = append(b.items, n)
	b.dirty = true
	return b
}

// RemoveItem removes the item with the given key and all its descendants.
// If the key does not exist, this is a no-op.
//
// Example:
//
//	b.RemoveItem(2) // removes node 2 and all its children
func (b *Builder[T]) RemoveItem(key any) *Builder[T] {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.keyFn == nil {
		return b
	}

	if err := b.ensureBuiltLocked(); err != nil {
		return b
	}

	cached, ok := b.nodeCache[key]
	if !ok {
		return b
	}

	toRemove := make(map[any]bool)
	stack := []*Node[T]{cached}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		k := b.keyFn(cur.Item)
		toRemove[k] = true
		stack = append(stack, cur.Children...)
	}

	for k := range toRemove {
		delete(b.index, k)
		if b.parentOverrides != nil {
			delete(b.parentOverrides, k)
		}
	}

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
func (b *Builder[T]) MoveItem(key, newParentKey any) *Builder[T] {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.keyFn == nil {
		return b
	}

	if err := b.ensureBuiltLocked(); err != nil {
		return b
	}

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
		b.parentOverrides = make(map[any]any)
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
func (b *Builder[T]) UpdateItem(key any, fn func(*T)) *Builder[T] {
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
func (b *Builder[T]) Filter(predicate func(T) bool) *Builder[T] {
	b.mu.Lock()
	defer b.mu.Unlock()

	nb := &Builder[T]{
		keyFn:    b.keyFn,
		parentFn: b.parentFn,
		sortFn:   b.sortFn,
		items:    make([]*node[T], 0, len(b.items)),
		index:    make(map[any]*node[T], len(b.items)),
		dirty:    true,
	}

	for _, n := range b.items {
		if predicate == nil || predicate(n.item) {
			copied := *n
			if b.keyFn != nil {
				nb.index[copied.key] = &copied
			}
			nb.items = append(nb.items, &copied)
		}
	}

	if b.parentOverrides != nil && len(nb.items) > 0 {
		nb.parentOverrides = make(map[any]any, len(b.parentOverrides))
		for _, n := range nb.items {
			if pk, ok := b.parentOverrides[n.key]; ok {
				nb.parentOverrides[n.key] = pk
			}
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
func (b *Builder[T]) Transform(fn func(*T)) *Builder[T] {
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

// Map returns a new Builder with all items transformed by fn.
// The new Builder inherits parent and sort functions from the original.
// Key function must be provided for the new type.
//
// Example:
//
//	ids := b.Map(func(d Dept) string { return d.Name })
func (b *Builder[T]) Map(fn func(T) T, keyFn func(T) any) *Builder[T] {
	b.mu.Lock()
	defer b.mu.Unlock()

	nb := &Builder[T]{
		keyFn:    keyFn,
		parentFn: b.parentFn,
		sortFn:   b.sortFn,
		items:    make([]*node[T], 0, len(b.items)),
		index:    make(map[any]*node[T], len(b.items)),
		dirty:    true,
	}

	for _, n := range b.items {
		mapped := fn(n.item)
		newNode := &node[T]{item: mapped, insertOrder: n.insertOrder}
		if keyFn != nil {
			newNode.key = keyFn(mapped)
			nb.index[newNode.key] = newNode
		}
		nb.items = append(nb.items, newNode)
	}

	if b.parentOverrides != nil && len(nb.items) > 0 {
		nb.parentOverrides = make(map[any]any, len(b.parentOverrides))
		for _, n := range nb.items {
			if pk, ok := b.parentOverrides[n.key]; ok {
				nb.parentOverrides[n.key] = pk
			}
		}
	}

	nb.insertCtr = len(nb.items)
	return nb
}

// Find returns the first node matching predicate, or nil if not found.
// Searches in insertion order (not tree order).
//
// Example:
//
//	node := b.Find(func(d Dept) bool { return d.Name == "HR" })
func (b *Builder[T]) Find(predicate func(T) bool) *Node[T] {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.keyFn == nil {
		return nil
	}

	if err := b.ensureBuiltLocked(); err != nil {
		return nil
	}

	for _, n := range b.items {
		if predicate(n.item) {
			return b.nodeCache[n.key]
		}
	}
	return nil
}

// Contains returns true if a node with the given key exists.
//
// Example:
//
//	exists := b.Contains(5)
func (b *Builder[T]) Contains(key any) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.keyFn == nil {
		return false
	}

	if err := b.ensureBuiltLocked(); err != nil {
		return false
	}
	_, ok := b.nodeCache[key]
	return ok
}

// Depth returns the depth of the node at key (root = 1).
// Returns 0 if key does not exist.
//
// Example:
//
//	depth := b.Depth(5)
func (b *Builder[T]) Depth(key any) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.keyFn == nil {
		return 0
	}

	if err := b.ensureBuiltLocked(); err != nil {
		return 0
	}

	target, ok := b.nodeCache[key]
	if !ok {
		return 0
	}

	type entry struct {
		n     *Node[T]
		depth int
	}
	stack := make([]entry, 0, len(b.roots)*2)
	for _, r := range b.roots {
		stack = append(stack, entry{r, 1})
	}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if cur.n == target {
			return cur.depth
		}
		for _, child := range cur.n.Children {
			stack = append(stack, entry{child, cur.depth + 1})
		}
	}
	return 0
}

// Build constructs and returns the tree. Returns root nodes in sort order and a
// map for direct key lookup. Returns nil, nil if KeyBy is not set.
// The returned roots slice and nodeMap must not be modified by the caller.
//
// Example:
//
//	roots, nodeMap, err := b.Build()
func (b *Builder[T]) Build() ([]*Node[T], map[any]*Node[T], error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.keyFn == nil {
		return nil, nil, ErrKeyNotSet
	}
	if err := b.ensureBuiltLocked(); err != nil {
		return nil, nil, err
	}
	return b.roots, b.nodeCache, nil
}

// Clone returns a new Builder with an independent deep copy of all items.
// Key extraction functions are shared (not copied) between original and clone.
//
// Example:
//
//	copy := b.Clone()
func (b *Builder[T]) Clone() *Builder[T] {
	b.mu.RLock()
	defer b.mu.RUnlock()

	nb := &Builder[T]{
		keyFn:     b.keyFn,
		parentFn:  b.parentFn,
		sortFn:    b.sortFn,
		insertCtr: b.insertCtr,
		items:     make([]*node[T], len(b.items)),
		index:     make(map[any]*node[T], len(b.index)),
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
		nb.parentOverrides = make(map[any]any, len(b.parentOverrides))
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
func (b *Builder[T]) Validate() []error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.keyFn == nil {
		return []error{ErrKeyNotSet}
	}

	if err := b.ensureBuiltLocked(); err != nil {
		return []error{err}
	}

	var errs []error

	for k := range b.nodeCache {
		n := b.index[k]
		pk, isRoot := b.effectiveParent(n)
		if isRoot {
			continue
		}
		if _, exists := b.nodeCache[pk]; !exists {
			errs = append(errs, fmt.Errorf("orphaned node: %v", k))
		}
	}

	type state uint8
	const (
		unvisited state = iota
		inProgress
		done
	)
	visited := make(map[any]state, len(b.nodeCache))

	for k := range b.nodeCache {
		if visited[k] != unvisited {
			continue
		}
		stack := []any{k}
		path := make(map[any]struct{})
		for len(stack) > 0 {
			cur := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if _, inPath := path[cur]; inPath {
				errs = append(errs, fmt.Errorf("cycle detected: %v", cur))
				break
			}
			if visited[cur] == done {
				continue
			}
			if visited[cur] == inProgress {
				path[cur] = struct{}{}
			}
			visited[cur] = inProgress

			n := b.index[cur]
			if n == nil {
				visited[cur] = done
				continue
			}
			pk, isRoot := b.effectiveParent(n)
			if isRoot {
				visited[cur] = done
				continue
			}
			if _, parentExists := b.nodeCache[pk]; parentExists {
				stack = append(stack, pk)
			}
		}
		for k := range path {
			visited[k] = done
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
func (b *Builder[T]) Statistics() map[string]any {
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

	if err := b.ensureBuiltLocked(); err != nil {
		return map[string]any{
			"totalNodes":         0,
			"rootNodes":          0,
			"maxDepth":           0,
			"leafNodes":          0,
			"avgChildrenPerNode": 0.0,
		}
	}

	total := len(b.nodeCache)
	rootCount := len(b.roots)
	depth := b.maxDepth()
	leaves := b.leafCount()

	var avg float64
	if total > 0 {
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
