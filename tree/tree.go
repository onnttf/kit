package tree

import (
	"fmt"
	"sort"
)

// Node represents a single node in the tree structure.
// Each node has a unique key, optional parent relationship, sort order, and child nodes.
type Node struct {
	NodeKey       string  // Unique identifier for this node
	ParentNodeKey string  // Key of the parent node, empty string for root nodes
	Sort          int     // Sort order within the same parent level
	Children      []*Node // Child nodes, built automatically when needed
}

// TreeBuilder provides a flexible tree construction mechanism.
// It allows adding, modifying, and building tree structures with automatic relationship management.
// The builder maintains internal state and rebuilds relationships when necessary for optimal performance.
type TreeBuilder struct {
	nodeMap   map[string]*Node // Maps node keys to node instances
	rootNodes []*Node          // Top-level nodes with no parent
	dirty     bool             // Indicates if relationships need rebuilding
}

// NewTreeBuilder creates a new empty TreeBuilder instance.
// Returns a TreeBuilder ready for adding nodes or importing from a list.
func NewTreeBuilder() *TreeBuilder {
	return &TreeBuilder{
		nodeMap:   make(map[string]*Node),
		rootNodes: make([]*Node, 0),
		dirty:     true,
	}
}

// FromList initializes the TreeBuilder with a flat list of nodes.
// Tree relationships are built automatically when needed.
// Nodes are cloned to prevent external modifications from affecting the internal tree structure.
//
// Parameters:
//   - nodes: Flat list of nodes to import, nil nodes are safely ignored
//
// Returns the TreeBuilder instance for method chaining.
func (tb *TreeBuilder) FromList(nodes []*Node) *TreeBuilder {
	tb.nodeMap = make(map[string]*Node, len(nodes))
	tb.rootNodes = make([]*Node, 0, len(nodes)/4) // Estimate ~25% will be root nodes
	tb.dirty = true

	for _, node := range nodes {
		if node == nil {
			continue
		}
		// Clone the node to prevent external modifications
		clonedNode := &Node{
			NodeKey:       node.NodeKey,
			ParentNodeKey: node.ParentNodeKey,
			Sort:          node.Sort,
			Children:      make([]*Node, 0, 4), // Pre-allocate for typical branching factor
		}
		tb.nodeMap[node.NodeKey] = clonedNode
	}

	return tb
}

// AddNode inserts a new node with the specified key, parent relationship, and sort order.
// If a node with the same key already exists, it will be replaced.
//
// Parameters:
//   - nodeKey: Unique identifier for the new node
//   - parentNodeKey: Key of the parent node, empty string for root node
//   - sort: Sort order within the same parent level (lower values appear first)
//
// Returns the TreeBuilder instance for method chaining.
func (tb *TreeBuilder) AddNode(nodeKey, parentNodeKey string, sort int) *TreeBuilder {
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

// RemoveNode deletes a node and all its descendants from the tree.
// The operation is performed recursively to maintain tree integrity.
// If the specified node doesn't exist, the operation is safely ignored.
//
// Parameters:
//   - nodeKey: Key of the node to remove
//
// Returns the TreeBuilder instance for method chaining.
func (tb *TreeBuilder) RemoveNode(nodeKey string) *TreeBuilder {
	tb.ensureBuilt()
	node := tb.nodeMap[nodeKey]
	if node == nil {
		return tb // Node doesn't exist, nothing to remove
	}
	tb.removeNodeRecursively(node)
	tb.dirty = true
	return tb
}

// removeNodeRecursively deletes a node and all its descendants from the nodeMap.
// This ensures no orphaned references remain in the tree structure.
func (tb *TreeBuilder) removeNodeRecursively(node *Node) {
	// Recursively remove all children first
	for _, child := range node.Children {
		tb.removeNodeRecursively(child)
	}
	// Remove this node from the map
	delete(tb.nodeMap, node.NodeKey)
}

// MoveNode changes a node's parent to the specified newParentKey.
// Self-references are ignored to prevent invalid tree structures.
// If the node doesn't exist, the operation is safely ignored.
//
// Parameters:
//   - nodeKey: Key of the node to move
//   - newParentKey: Key of the new parent node, empty string to make it a root node
//
// Returns the TreeBuilder instance for method chaining.
func (tb *TreeBuilder) MoveNode(nodeKey, newParentKey string) *TreeBuilder {
	node := tb.nodeMap[nodeKey]
	if node == nil || nodeKey == newParentKey {
		return tb // Node doesn't exist or trying to move to itself
	}
	node.ParentNodeKey = newParentKey
	tb.dirty = true
	return tb
}

// UpdateNode applies a transformation function to a specific node.
// The transformer function receives the node instance and can modify its properties.
// If the node doesn't exist, the operation is safely ignored.
//
// Parameters:
//   - nodeKey: Key of the node to update
//   - transformer: Function that receives and can modify the node
//
// Returns the TreeBuilder instance for method chaining.
func (tb *TreeBuilder) UpdateNode(nodeKey string, transformer func(*Node)) *TreeBuilder {
	if node := tb.nodeMap[nodeKey]; node != nil {
		transformer(node)
		tb.dirty = true
	}
	return tb
}

// Filter creates a new TreeBuilder containing only nodes that match the predicate.
// The predicate function is applied to each node; matching nodes are included in the result.
// Parent-child relationships are preserved in the filtered tree, but orphaned nodes
// (whose parents don't match the predicate) will become root nodes in the new tree.
//
// Parameters:
//   - predicate: Function that determines whether a node should be included
//
// Returns a new TreeBuilder containing only matching nodes.
func (tb *TreeBuilder) Filter(predicate func(*Node) bool) *TreeBuilder {
	tb.ensureBuilt()
	newBuilder := NewTreeBuilder()

	// Recursively traverse and filter nodes
	var addNodeIfMatch func(*Node)
	addNodeIfMatch = func(node *Node) {
		if predicate(node) {
			newBuilder.nodeMap[node.NodeKey] = &Node{
				NodeKey:       node.NodeKey,
				ParentNodeKey: node.ParentNodeKey,
				Sort:          node.Sort,
				Children:      make([]*Node, 0),
			}
		}
		for _, child := range node.Children {
			addNodeIfMatch(child)
		}
	}

	for _, root := range tb.rootNodes {
		addNodeIfMatch(root)
	}

	newBuilder.dirty = true
	return newBuilder
}

// Transform applies a transformation function to all nodes in the tree.
// The transformer function receives each node instance and can modify its properties.
// This is useful for bulk updates like changing sort orders or updating node data.
//
// Parameters:
//   - transformer: Function that receives and can modify each node
//
// Returns the TreeBuilder instance for method chaining.
func (tb *TreeBuilder) Transform(transformer func(*Node)) *TreeBuilder {
	for _, node := range tb.nodeMap {
		transformer(node)
	}
	tb.dirty = true
	return tb
}

// Build finalizes the tree structure and returns the node map and root nodes.
// This method ensures all relationships are built and sorted before returning.
// The returned structures are safe to use and reflect the current tree state.
//
// Returns:
//   - map[string]*Node: All nodes indexed by their NodeKey
//   - []*Node: Root nodes (nodes with no parent) sorted by their Sort field
func (tb *TreeBuilder) Build() (map[string]*Node, []*Node) {
	tb.ensureBuilt()
	return tb.nodeMap, tb.rootNodes
}

// Clone creates a deep copy of the current tree structure.
// The cloned tree is completely independent and can be modified without affecting the original.
// All node relationships and sort orders are preserved in the clone.
//
// Returns a new TreeBuilder that is an exact copy of the current tree.
func (tb *TreeBuilder) Clone() *TreeBuilder {
	tb.ensureBuilt()
	newBuilder := NewTreeBuilder()
	newBuilder.nodeMap = make(map[string]*Node, len(tb.nodeMap))

	// Deep copy each node
	for key, node := range tb.nodeMap {
		newBuilder.nodeMap[key] = &Node{
			NodeKey:       node.NodeKey,
			ParentNodeKey: node.ParentNodeKey,
			Sort:          node.Sort,
			Children:      make([]*Node, 0, len(node.Children)),
		}
	}
	newBuilder.dirty = true
	return newBuilder
}

// Validate checks the tree structure for common issues and returns a list of errors.
// This method performs comprehensive validation to ensure tree integrity.
//
// Checks performed:
//   - Cycle detection: Ensures no circular references exist in the tree
//   - Orphan detection: Identifies nodes whose parents don't exist
//
// Returns a slice of errors describing any issues found. An empty slice indicates a valid tree.
func (tb *TreeBuilder) Validate() []error {
	tb.ensureBuilt()
	var errors []error

	// Detect cycles using DFS with recursion stack tracking
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
				return true // Back edge found - cycle detected
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

	// Detect orphaned nodes (nodes not reachable from any root)
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

// Statistics computes and returns basic metrics about the tree structure.
// This is useful for monitoring tree complexity and performance characteristics.
//
// Returned metrics include:
//   - total_nodes: Total number of nodes in the tree
//   - root_nodes: Number of root nodes (nodes with no parent)
//   - max_depth: Maximum depth of the tree (distance from root to deepest leaf)
//   - leaf_nodes: Number of leaf nodes (nodes with no children)
//   - avg_children_per_node: Average number of children per node
//
// Returns a map containing the computed statistics.
func (tb *TreeBuilder) Statistics() map[string]interface{} {
	tb.ensureBuilt()
	stats := make(map[string]interface{})

	stats["total_nodes"] = len(tb.nodeMap)
	stats["root_nodes"] = len(tb.rootNodes)
	stats["max_depth"] = tb.getMaxDepth()
	stats["leaf_nodes"] = len(tb.getLeafNodes())

	// Calculate average children per node
	totalChildren := 0
	for _, node := range tb.nodeMap {
		totalChildren += len(node.Children)
	}
	if len(tb.nodeMap) > 0 {
		stats["avg_children_per_node"] = float64(totalChildren) / float64(len(tb.nodeMap))
	}

	return stats
}

// getMaxDepth computes the maximum depth of the tree using iterative DFS.
// Depth is measured from root nodes (depth 1) to the deepest leaf.
// Returns 0 if the tree is empty (no root nodes).
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

	// Initialize stack with root nodes at depth 1
	for _, node := range tb.rootNodes {
		stack = append(stack, nodeWithDepth{node: node, depth: 1})
	}

	// Process nodes using iterative DFS
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if current.depth > maxDepth {
			maxDepth = current.depth
		}

		// Add children to stack with incremented depth
		for _, child := range current.node.Children {
			stack = append(stack, nodeWithDepth{node: child, depth: current.depth + 1})
		}
	}

	return maxDepth
}

// getLeafNodes returns all leaf nodes (nodes with no children) in the tree.
// Uses iterative DFS to traverse the entire tree structure efficiently.
func (tb *TreeBuilder) getLeafNodes() []*Node {
	var leaves []*Node
	stack := make([]*Node, 0, len(tb.rootNodes))

	// Initialize stack with root nodes
	stack = append(stack, tb.rootNodes...)

	// Process nodes using iterative DFS
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if len(current.Children) == 0 {
			leaves = append(leaves, current) // This is a leaf node
		} else {
			stack = append(stack, current.Children...) // Add children to stack
		}
	}

	return leaves
}

// ensureBuilt builds the tree structure if it needs to be updated.
// This implements efficient rebuilding - relationships are only reconstructed when necessary.
func (tb *TreeBuilder) ensureBuilt() {
	if tb.dirty {
		tb.buildRelationshipsAndSort()
		tb.dirty = false
	}
}

// buildRelationshipsAndSort reconstructs parent-child relationships and sorts nodes.
// This method handles the core tree building logic.
//
// Root node criteria:
//   - ParentNodeKey is empty string, OR
//   - ParentNodeKey equals NodeKey (self-reference treated as root)
//
// Orphaned nodes (nodes with non-existent parents) are not added to any parent's
// children list or to the root nodes list, making them detectable via Validate().
func (tb *TreeBuilder) buildRelationshipsAndSort() {
	tb.rootNodes = make([]*Node, 0)

	// Clear existing children relationships
	for _, node := range tb.nodeMap {
		node.Children = make([]*Node, 0, 4)
	}

	// Build parent-child relationships
	for _, node := range tb.nodeMap {
		// Check if this is a root node
		if node.ParentNodeKey == "" || node.ParentNodeKey == node.NodeKey {
			tb.rootNodes = append(tb.rootNodes, node)
			continue
		}

		// Add to parent's children if parent exists
		if parent, exists := tb.nodeMap[node.ParentNodeKey]; exists {
			parent.Children = append(parent.Children, node)
		}
		// Note: If parent doesn't exist, node becomes orphaned
	}

	// Sort all nodes by their Sort field
	tb.sortNodesRecursively(tb.rootNodes)
}

// sortNodesRecursively sorts nodes by their Sort field and recursively sorts all descendants.
// Uses in-place sorting for memory efficiency. Nodes with the same sort value maintain
// their relative order (stable sort).
func (tb *TreeBuilder) sortNodesRecursively(nodes []*Node) {
	if len(nodes) <= 1 {
		return // No sorting needed for 0 or 1 nodes
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Sort < nodes[j].Sort
	})

	// Recursively sort children of each node
	for _, node := range nodes {
		if len(node.Children) > 1 {
			tb.sortNodesRecursively(node.Children)
		}
	}
}
