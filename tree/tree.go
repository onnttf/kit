package tree

import (
	"errors"
	"fmt"
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
	b.roots = b.roots[:0]
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
// Caller must hold mu.
func (b *Builder[T, K]) sortNodes(nodes []*Node[T, K]) {
	if len(nodes) <= 1 {
		return
	}
	sort.SliceStable(nodes, func(i, j int) bool {
		ki := b.keyFn(nodes[i].Item)
		kj := b.keyFn(nodes[j].Item)
		return b.index[ki].sortVal < b.index[kj].sortVal
	})
	for _, n := range nodes {
		b.sortNodes(n.Children)
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

// isDescendant reports whether potentialDesc is a descendant of ancestorKey.
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

// Placeholder stubs — implemented in later tasks.
// These exist only to allow the package to compile during TDD.

func (b *Builder[T, K]) WithItems(items []T) *Builder[T, K]           { return b }
func (b *Builder[T, K]) AddItem(item T) *Builder[T, K]                { return b }
func (b *Builder[T, K]) RemoveItem(key K) *Builder[T, K]              { return b }
func (b *Builder[T, K]) MoveItem(key, newParentKey K) *Builder[T, K]  { return b }
func (b *Builder[T, K]) UpdateItem(key K, fn func(*T)) *Builder[T, K] { return b }
func (b *Builder[T, K]) Filter(predicate func(T) bool) *Builder[T, K] { return b }
func (b *Builder[T, K]) Transform(fn func(*T)) *Builder[T, K]         { return b }
func (b *Builder[T, K]) Build() ([]*Node[T, K], map[K]*Node[T, K])    { return nil, nil }
func (b *Builder[T, K]) Clone() *Builder[T, K]                        { return b }
func (b *Builder[T, K]) Validate() []error                            { return nil }
func (b *Builder[T, K]) Statistics() map[string]any                   { return nil }

// Suppress unused import errors during stub phase.
var _ = fmt.Sprintf
var _ = errors.New
