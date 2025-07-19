package tree

import (
	"reflect"
	"strings"
	"testing"
)

func newNodes() []*Node {
	return []*Node{
		{NodeKey: "1", ParentNodeKey: "1", Sort: 1}, // root
		{NodeKey: "2", ParentNodeKey: "1", Sort: 2},
		{NodeKey: "3", ParentNodeKey: "2", Sort: 1},
		{NodeKey: "4", ParentNodeKey: "2", Sort: 2},
		{NodeKey: "5", ParentNodeKey: "1", Sort: 3},
	}
}

func TestBuildWithNodes(t *testing.T) {
	tb := NewTreeBuilder().WithNodes(newNodes())
	nodeMap, roots := tb.Build()

	if len(roots) != 1 {
		t.Errorf("expected 1 root, got %d", len(roots))
	}
	if _, ok := nodeMap["2"]; !ok {
		t.Error("node 2 not found in nodeMap")
	}
	if len(nodeMap["2"].Children) != 2 {
		t.Errorf("expected node 2 to have 2 children, got %d", len(nodeMap["2"].Children))
	}
}

func TestAddNodeAndBuild(t *testing.T) {
	tb := NewTreeBuilder()
	tb.AddNode("1", "1", 1).AddNode("2", "1", 2).AddNode("3", "2", 3)
	nodeMap, roots := tb.Build()

	if len(roots) != 1 {
		t.Errorf("expected 1 root, got %d", len(roots))
	}
	if _, ok := nodeMap["3"]; !ok {
		t.Error("node 3 not found in nodeMap")
	}
	if len(nodeMap["2"].Children) != 1 {
		t.Errorf("expected node 2 to have 1 child, got %d", len(nodeMap["2"].Children))
	}
}

func TestMoveNode(t *testing.T) {
	tb := NewTreeBuilder().WithNodes(newNodes())
	tb.MoveNode("5", "2")
	nodeMap, _ := tb.Build()

	if nodeMap["5"].ParentNodeKey != "2" {
		t.Error("node 5 parent not updated")
	}
	if len(nodeMap["1"].Children) != 1 {
		t.Errorf("node 1 should have 1 child after move, got %d", len(nodeMap["1"].Children))
	}
	if len(nodeMap["2"].Children) != 3 {
		t.Errorf("node 2 should have 3 children after move, got %d", len(nodeMap["2"].Children))
	}
}

func TestRemoveNode(t *testing.T) {
	tb := NewTreeBuilder().WithNodes(newNodes())
	tb.RemoveNode("2")
	nodeMap, _ := tb.Build()

	if _, ok := nodeMap["2"]; ok {
		t.Error("node 2 was not removed")
	}
	if _, ok := nodeMap["3"]; ok {
		t.Error("descendant node 3 was not removed")
	}
}

func TestUpdateNode(t *testing.T) {
	tb := NewTreeBuilder().WithNodes(newNodes())
	tb.UpdateNode("5", func(n *Node) { n.Sort = 10 })
	nodeMap, _ := tb.Build()

	if nodeMap["5"].Sort != 10 {
		t.Errorf("expected node 5 Sort to be updated to 10, got %d", nodeMap["5"].Sort)
	}
}

func TestFilter(t *testing.T) {
	tb := NewTreeBuilder().WithNodes(newNodes())
	newTb := tb.Filter(func(n *Node) bool { return n.Sort%2 == 1 })
	nodeMap, _ := newTb.Build()

	for _, n := range nodeMap {
		if n.Sort%2 != 1 {
			t.Errorf("expected only odd Sort nodes, found Sort=%d", n.Sort)
		}
	}
}

func TestTransform(t *testing.T) {
	tb := NewTreeBuilder().WithNodes(newNodes())
	tb.Transform(func(n *Node) { n.Sort = 42 })
	nodeMap, _ := tb.Build()
	for _, n := range nodeMap {
		if n.Sort != 42 {
			t.Errorf("Transform failed, expected 42, got %d", n.Sort)
		}
	}
}

func TestClone(t *testing.T) {
	tb := NewTreeBuilder().WithNodes(newNodes())
	clone := tb.Clone()
	clone.UpdateNode("1", func(n *Node) { n.Sort = 100 })

	orig, _ := tb.Build()
	after, _ := clone.Build()
	if orig["1"].Sort == after["1"].Sort {
		t.Error("modifying clone should not affect original")
	}
}

func TestValidate_CycleAndOrphan(t *testing.T) {
	// Cycle
	nodes := []*Node{
		{NodeKey: "1", ParentNodeKey: "2", Sort: 1},
		{NodeKey: "2", ParentNodeKey: "1", Sort: 2},
	}
	tb := NewTreeBuilder().WithNodes(nodes)
	errs := tb.Validate()
	hasCycle := false
	for _, err := range errs {
		if err.Error() == "cycle detected in tree starting from node 1" || err.Error() == "cycle detected in tree starting from node 2" {
			hasCycle = true
		}
	}
	if !hasCycle {
		t.Error("expected cycle detected error")
	}

	// Orphan
	nodes = []*Node{
		{NodeKey: "1", ParentNodeKey: "1", Sort: 1},
		{NodeKey: "2", ParentNodeKey: "3", Sort: 2},
	}
	tb = NewTreeBuilder().WithNodes(nodes)
	errs = tb.Validate()
	hasOrphan := false
	t.Log("nodeMap:", tb.nodeMap)
	t.Log("rootNodes:", tb.rootNodes)
	t.Log("Validate errs:")
	for _, err := range errs {
		if strings.Contains(err.Error(), "orphaned node found") && strings.Contains(err.Error(), "2") {
			hasOrphan = true
		}
	}
	if !hasOrphan {
		t.Error("expected orphan detected error")
	}
}

func TestStatistics(t *testing.T) {
	tb := NewTreeBuilder().WithNodes(newNodes())
	stats := tb.Statistics()

	if stats["total_nodes"] != 5 {
		t.Errorf("expected total_nodes 5, got %v", stats["total_nodes"])
	}
	if stats["root_nodes"] != 1 {
		t.Errorf("expected root_nodes 1, got %v", stats["root_nodes"])
	}
	if stats["max_depth"] != 3 {
		t.Errorf("expected max_depth 3, got %v", stats["max_depth"])
	}
	if stats["leaf_nodes"] != 3 {
		t.Errorf("expected leaf_nodes 3, got %v", stats["leaf_nodes"])
	}
	if avg, ok := stats["avg_children_per_node"].(float64); !ok || avg <= 0.0 {
		t.Errorf("expected avg_children_per_node > 0, got %v", stats["avg_children_per_node"])
	}
}

func TestOrderStableSort(t *testing.T) {
	nodes := []*Node{
		{NodeKey: "1", ParentNodeKey: "1", Sort: 1},
		{NodeKey: "2", ParentNodeKey: "1", Sort: 2},
		{NodeKey: "3", ParentNodeKey: "1", Sort: 2},
		{NodeKey: "4", ParentNodeKey: "1", Sort: 3},
	}
	tb := NewTreeBuilder().WithNodes(nodes)
	_, roots := tb.Build()
	children := roots[0].Children
	if len(children) != 3 {
		t.Errorf("expected 3 children under root, got %d", len(children))
	}
	// Children with same Sort should keep input order (2 before 3)
	idx2, idx3 := -1, -1
	for i, n := range children {
		if n.NodeKey == "2" {
			idx2 = i
		}
		if n.NodeKey == "3" {
			idx3 = i
		}
	}
	if idx2 == -1 || idx3 == -1 || idx2 > idx3 {
		t.Errorf("stable sort failed, node 2 should be before node 3")
	}
}

func TestEmptyTree(t *testing.T) {
	tb := NewTreeBuilder()
	nodeMap, roots := tb.Build()
	if len(nodeMap) != 0 {
		t.Errorf("expected empty nodeMap, got %d", len(nodeMap))
	}
	if len(roots) != 0 {
		t.Errorf("expected empty roots, got %d", len(roots))
	}
}

func TestWithNodesNilNodes(t *testing.T) {
	nodes := []*Node{
		{NodeKey: "1", ParentNodeKey: "1", Sort: 1},
		nil,
		{NodeKey: "2", ParentNodeKey: "1", Sort: 2},
	}
	tb := NewTreeBuilder().WithNodes(nodes)
	nodeMap, roots := tb.Build()
	if len(nodeMap) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodeMap))
	}
	if len(roots) != 1 {
		t.Errorf("expected 1 root, got %d", len(roots))
	}
}

func TestFilterEmpty(t *testing.T) {
	tb := NewTreeBuilder().WithNodes(newNodes())
	newTb := tb.Filter(func(n *Node) bool { return false })
	nodeMap, roots := newTb.Build()
	if len(nodeMap) != 0 {
		t.Errorf("expected filtered nodeMap to be empty, got %d", len(nodeMap))
	}
	if len(roots) != 0 {
		t.Errorf("expected filtered roots to be empty, got %d", len(roots))
	}
}

func TestTransformNoop(t *testing.T) {
	tb := NewTreeBuilder().WithNodes(newNodes())
	tb.Transform(func(n *Node) {})
	nodeMap, roots := tb.Build()
	tb2 := NewTreeBuilder().WithNodes(newNodes())
	nodeMap2, roots2 := tb2.Build()
	if !reflect.DeepEqual(nodeMap, nodeMap2) || !reflect.DeepEqual(roots, roots2) {
		t.Errorf("noop Transform changed the tree")
	}
}
