package tree

import (
	"fmt"
	"sort"
	"sync"
)

// Node defines a single node in a tree with a unique key, parent reference, sort order, and children
type Node struct {
	NodeKey       string  // Unique identifier for the node
	ParentNodeKey string  // Key of the parent node, empty for root nodes
	Sort          int     // Sort order among siblings
	Children      []*Node // Child nodes, built automatically
}

// TreeBuilder defines a builder for constructing and managing tree structures with automatic relationship handling.
// TreeBuilder is safe for concurrent use by multiple goroutines.
type TreeBuilder struct {
	mu        sync.RWMutex     // Protects concurrent access to all fields
	nodeMap   map[string]*Node // Maps node keys to nodes
	rootNodes []*Node          // Root nodes with no parent
	dirty     bool             // Indicates if relationships need rebuilding
}

// NewTreeBuilder returns a new TreeBuilder for creating tree structures
func NewTreeBuilder() *TreeBuilder {
	return &TreeBuilder{
		nodeMap:   make(map[string]*Node),
		rootNodes: make([]*Node, 0),
		dirty:     true,
	}
}

// WithNodes returns the TreeBuilder initialized with a cloned slice of nodes to build a tree.
func (tb *TreeBuilder) WithNodes(nodes []*Node) *TreeBuilder {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.nodeMap = make(map[string]*Node, len(nodes))
	tb.rootNodes = make([]*Node, 0, len(nodes)/4)
	tb.dirty = true

	for _, node := range nodes {
		if node == nil {
			continue
		}
		tb.nodeMap[node.NodeKey] = tb.cloneNode(node)
	}

	return tb
}

// AddNode returns the TreeBuilder after adding a node with the specified key, parent, and sort order
func (tb *TreeBuilder) AddNode(nodeKey, parentNodeKey string, sort int) *TreeBuilder {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	node := &Node{
		NodeKey:       nodeKey,
		ParentNodeKey: parentNodeKey,
		Sort:          sort,
		Children:      make([]*Node, 0, 4),
	}
	tb.nodeMap[nodeKey] = node
	tb.dirty = true
	return tb
}

// RemoveNode returns the TreeBuilder after removing a node and its descendants
func (tb *TreeBuilder) RemoveNode(nodeKey string) *TreeBuilder {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.ensureBuiltLocked()
	node := tb.nodeMap[nodeKey]
	if node == nil {
		return tb
	}
	tb.removeNodeRecursively(node)
	tb.dirty = true
	return tb
}

// removeNodeRecursively removes a node and its descendants from the node map
func (tb *TreeBuilder) removeNodeRecursively(node *Node) {
	for _, child := range node.Children {
		tb.removeNodeRecursively(child)
	}
	delete(tb.nodeMap, node.NodeKey)
}

// MoveNode returns the TreeBuilder after moving a node to a new parent.
// It performs cycle detection to prevent creating circular references (A->B->C->A).
// Self-references and moves that would create cycles are silently ignored.
func (tb *TreeBuilder) MoveNode(nodeKey, newParentKey string) *TreeBuilder {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	node := tb.nodeMap[nodeKey]
	if node == nil || nodeKey == newParentKey {
		return tb
	}

	// Check if the move would create a cycle by checking if newParentKey is a descendant of nodeKey
	if tb.isDescendant(nodeKey, newParentKey) {
		// Silently ignore moves that would create cycles
		return tb
	}

	node.ParentNodeKey = newParentKey
	tb.dirty = true
	return tb
}

// isDescendant checks if potentialDescendant is a descendant of ancestorKey.
// This prevents cycle creation when moving nodes.
// Must be called with tb.mu held.
func (tb *TreeBuilder) isDescendant(ancestorKey, potentialDescendant string) bool {
	if ancestorKey == potentialDescendant {
		return true
	}

	current := tb.nodeMap[potentialDescendant]
	visited := make(map[string]bool)

	for current != nil {
		if current.NodeKey == ancestorKey {
			return true
		}
		if visited[current.NodeKey] {
			// Cycle detected in existing tree structure
			return false
		}
		visited[current.NodeKey] = true

		if current.ParentNodeKey == "" || current.ParentNodeKey == current.NodeKey {
			return false
		}
		current = tb.nodeMap[current.ParentNodeKey]
	}

	return false
}

// UpdateNode returns the TreeBuilder after applying a transformation to a specific node
func (tb *TreeBuilder) UpdateNode(nodeKey string, transformer func(*Node)) *TreeBuilder {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if node := tb.nodeMap[nodeKey]; node != nil {
		transformer(node)
		tb.dirty = true
	}
	return tb
}

// Filter returns a new TreeBuilder with nodes matching the predicate, preserving relationships.
// Uses iterative approach to avoid stack overflow on deep trees.
func (tb *TreeBuilder) Filter(predicate func(*Node) bool) *TreeBuilder {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	tb.ensureBuiltLocked()
	newBuilder := NewTreeBuilder()

	// Use explicit stack to avoid recursion depth issues
	type stackItem struct {
		node *Node
	}
	stack := make([]stackItem, 0, len(tb.rootNodes))

	for _, root := range tb.rootNodes {
		stack = append(stack, stackItem{node: root})
	}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if predicate(current.node) {
			newBuilder.nodeMap[current.node.NodeKey] = tb.cloneNode(current.node)
		}

		// Add children to stack
		for i := len(current.node.Children) - 1; i >= 0; i-- {
			stack = append(stack, stackItem{node: current.node.Children[i]})
		}
	}

	newBuilder.dirty = true
	return newBuilder
}

// Transform returns the TreeBuilder after applying a transformation to all nodes
func (tb *TreeBuilder) Transform(transformer func(*Node)) *TreeBuilder {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	for _, node := range tb.nodeMap {
		transformer(node)
	}
	tb.dirty = true
	return tb
}

// Build returns the node map and sorted root nodes, ensuring relationships are updated
func (tb *TreeBuilder) Build() (map[string]*Node, []*Node) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.ensureBuiltLocked()
	return tb.nodeMap, tb.rootNodes
}

// Clone returns a new TreeBuilder with a deep copy of the current tree
func (tb *TreeBuilder) Clone() *TreeBuilder {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	tb.ensureBuiltLocked()
	newBuilder := NewTreeBuilder()
	newBuilder.nodeMap = make(map[string]*Node, len(tb.nodeMap))

	for key, node := range tb.nodeMap {
		newBuilder.nodeMap[key] = tb.cloneNode(node)
	}
	newBuilder.dirty = true
	return newBuilder
}

// Validate returns errors for issues like cycles or orphaned nodes in the tree
func (tb *TreeBuilder) Validate() []error {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	tb.ensureBuiltLocked()
	var errors []error

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(*Node) bool
	hasCycle = func(node *Node) bool {
		visited[node.NodeKey] = true
		recStack[node.NodeKey] = true

		for _, child := range node.Children {
			if !visited[child.NodeKey] {
				if hasCycle(child) {
					return true
				}
			} else if recStack[child.NodeKey] {
				return true
			}
		}

		recStack[node.NodeKey] = false
		return false
	}

	for _, node := range tb.nodeMap {
		if !visited[node.NodeKey] && hasCycle(node) {
			errors = append(errors, fmt.Errorf("cycle detected in tree starting from node %s", node.NodeKey))
		}
	}

	reachable := make(map[string]bool)
	var markReachable func(*Node)
	markReachable = func(node *Node) {
		reachable[node.NodeKey] = true
		for _, child := range node.Children {
			markReachable(child)
		}
	}

	for _, root := range tb.rootNodes {
		markReachable(root)
	}

	for key := range tb.nodeMap {
		if !reachable[key] {
			errors = append(errors, fmt.Errorf("orphaned node found: %s", key))
		}
	}

	return errors
}

// Statistics returns metrics about the tree, including node count and depth
func (tb *TreeBuilder) Statistics() map[string]any {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	tb.ensureBuiltLocked()
	stats := make(map[string]any)

	stats["total_nodes"] = len(tb.nodeMap)
	stats["root_nodes"] = len(tb.rootNodes)
	stats["max_depth"] = tb.getMaxDepth()
	stats["leaf_nodes"] = len(tb.getLeafNodes())

	totalChildren := 0
	for _, node := range tb.nodeMap {
		totalChildren += len(node.Children)
	}
	if len(tb.nodeMap) > 0 {
		stats["avg_children_per_node"] = float64(totalChildren) / float64(len(tb.nodeMap))
	}

	return stats
}

// getMaxDepth computes the maximum depth of the tree using iterative DFS
func (tb *TreeBuilder) getMaxDepth() int {
	if len(tb.rootNodes) == 0 {
		return 0
	}

	type nodeWithDepth struct {
		node  *Node
		depth int
	}

	maxDepth := 0
	stack := make([]nodeWithDepth, 0, len(tb.rootNodes))

	for _, node := range tb.rootNodes {
		stack = append(stack, nodeWithDepth{node: node, depth: 1})
	}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if current.depth > maxDepth {
			maxDepth = current.depth
		}

		for _, child := range current.node.Children {
			stack = append(stack, nodeWithDepth{node: child, depth: current.depth + 1})
		}
	}

	return maxDepth
}

// getLeafNodes retrieves leaf nodes (nodes with no children) using iterative DFS
func (tb *TreeBuilder) getLeafNodes() []*Node {
	var leaves []*Node
	stack := make([]*Node, 0, len(tb.rootNodes))

	stack = append(stack, tb.rootNodes...)

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if len(current.Children) == 0 {
			leaves = append(leaves, current)
		} else {
			stack = append(stack, current.Children...)
		}
	}

	return leaves
}

// ensureBuiltLocked builds the tree structure if relationships need updating.
// Must be called with tb.mu held (either read or write lock).
func (tb *TreeBuilder) ensureBuiltLocked() {
	if tb.dirty {
		tb.buildRelationshipsAndSort()
		tb.dirty = false
	}
}

// cloneNode creates a deep copy of a node without its children relationships.
// The Children slice is initialized empty, ready to be rebuilt.
func (tb *TreeBuilder) cloneNode(node *Node) *Node {
	return &Node{
		NodeKey:       node.NodeKey,
		ParentNodeKey: node.ParentNodeKey,
		Sort:          node.Sort,
		Children:      make([]*Node, 0, 4),
	}
}

// buildRelationshipsAndSort constructs parent-child relationships and sorts nodes by sort order.
// Optimized to use single pass and reuse underlying arrays where possible.
// Must be called with tb.mu held (write lock).
func (tb *TreeBuilder) buildRelationshipsAndSort() {
	tb.rootNodes = tb.rootNodes[:0] // Reuse underlying array

	// Reset children slices, reusing underlying arrays
	for _, node := range tb.nodeMap {
		node.Children = node.Children[:0]
	}

	// Single pass: build relationships and identify roots
	for _, node := range tb.nodeMap {
		if node.ParentNodeKey == "" || node.ParentNodeKey == node.NodeKey {
			tb.rootNodes = append(tb.rootNodes, node)
		} else if parent, exists := tb.nodeMap[node.ParentNodeKey]; exists {
			parent.Children = append(parent.Children, node)
		}
	}

	tb.sortNodesRecursively(tb.rootNodes)
}

// sortNodesRecursively sorts nodes by sort order and their descendants recursively
func (tb *TreeBuilder) sortNodesRecursively(nodes []*Node) {
	if len(nodes) <= 1 {
		return
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Sort < nodes[j].Sort
	})

	for _, node := range nodes {
		if len(node.Children) > 1 {
			tb.sortNodesRecursively(node.Children)
		}
	}
}
