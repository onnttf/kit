package tree

import (
	"errors"
	"strings"
	"testing"
)

type testItem struct {
	ID       int
	ParentID int
	Sort     int
	Name     string
}

func newBuilder() *Builder[testItem] {
	return NewBuilder[testItem]().
		ParentBy(func(d testItem) int { return d.ParentID }).
		SortBy(func(d testItem) int { return d.Sort })
}

func newItems() []testItem {
	return []testItem{
		{ID: 1, ParentID: 1, Sort: 1, Name: "root"},
		{ID: 2, ParentID: 1, Sort: 2, Name: "child-A"},
		{ID: 3, ParentID: 2, Sort: 1, Name: "grandchild-1"},
		{ID: 4, ParentID: 2, Sort: 2, Name: "grandchild-2"},
		{ID: 5, ParentID: 1, Sort: 3, Name: "child-B"},
	}
}

func TestNewBuilder(t *testing.T) {
	b := NewBuilder[testItem]()
	if b == nil {
		t.Fatal("NewBuilder returned nil")
	}
}

func TestKeyByNilIsNoop(t *testing.T) {
	b := NewBuilder[testItem]()
	b2 := b.KeyBy(nil)
	if b2 != b {
		t.Error("KeyBy(nil) should return same builder")
	}
}

func TestParentByNilIsNoop(t *testing.T) {
	b := NewBuilder[testItem]()
	b2 := b.ParentBy(nil)
	if b2 != b {
		t.Error("ParentBy(nil) should return same builder")
	}
}

func TestSortByNilIsNoop(t *testing.T) {
	b := NewBuilder[testItem]()
	b2 := b.SortBy(nil)
	if b2 != b {
		t.Error("SortBy(nil) should return same builder")
	}
}

func TestBuildNilKeyBy(t *testing.T) {
	b := NewBuilder[testItem]().WithItems(newItems())
	_, _, err := b.Build()
	if err == nil {
		t.Error("expected error when KeyBy is not set")
	}
}

func TestBuildEmpty(t *testing.T) {
	roots, nodeMap, _ := newBuilder().WithItems(nil).Build()
	if len(roots) != 0 {
		t.Errorf("expected 0 roots, got %d", len(roots))
	}
	if len(nodeMap) != 0 {
		t.Errorf("expected empty nodeMap, got %d", len(nodeMap))
	}
}

func TestWithItemsNil(t *testing.T) {
	roots, nodeMap, _ := newBuilder().WithItems(nil).Build()
	if len(roots) != 0 || len(nodeMap) != 0 {
		t.Error("WithItems(nil) should produce empty tree")
	}
}

func TestBuildWithItems(t *testing.T) {
	roots, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).Build()
	if len(roots) != 1 {
		t.Errorf("expected 1 root, got %d", len(roots))
	}
	if len(nodeMap) != 5 {
		t.Errorf("expected 5 nodes in map, got %d", len(nodeMap))
	}
	if len(nodeMap[2].Children) != 2 {
		t.Errorf("node 2 should have 2 children, got %d", len(nodeMap[2].Children))
	}
	if len(nodeMap[1].Children) != 2 {
		t.Errorf("node 1 should have 2 children, got %d", len(nodeMap[1].Children))
	}
}

func TestBuildNilParentBy(t *testing.T) {
	b := NewBuilder[testItem]().
		KeyBy(func(d testItem) int { return d.ID }).
		WithItems(newItems())
	roots, nodeMap, _ := b.Build()
	if len(roots) != 5 {
		t.Errorf("without ParentBy all items should be roots, got %d", len(roots))
	}
	if len(nodeMap) != 5 {
		t.Errorf("expected 5 nodes in map, got %d", len(nodeMap))
	}
}

func TestAddItem(t *testing.T) {
	b := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems())
	b.AddItem(testItem{ID: 6, ParentID: 2, Sort: 3, Name: "grandchild-3"})
	roots, nodeMap, _ := b.Build()

	if len(nodeMap) != 6 {
		t.Errorf("expected 6 nodes, got %d", len(nodeMap))
	}
	if nodeMap[6] == nil {
		t.Fatal("node 6 not found in nodeMap")
	}
	if nodeMap[6].Item.Name != "grandchild-3" {
		t.Errorf("unexpected name %q", nodeMap[6].Item.Name)
	}
	if len(nodeMap[2].Children) != 3 {
		t.Errorf("node 2 should have 3 children after add, got %d", len(nodeMap[2].Children))
	}
	_ = roots
}

func TestAddItemPreservesInsertionOrder(t *testing.T) {
	b := NewBuilder[testItem]().
		KeyBy(func(d testItem) int { return d.ID }).
		ParentBy(func(d testItem) int { return d.ParentID }).
		WithItems(nil)

	b.AddItem(testItem{ID: 1, ParentID: 1, Name: "root"})
	b.AddItem(testItem{ID: 2, ParentID: 1, Name: "first"})
	b.AddItem(testItem{ID: 3, ParentID: 1, Name: "second"})

	roots, _, _ := b.Build()
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	children := roots[0].Children
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	if children[0].Item.Name != "first" || children[1].Item.Name != "second" {
		t.Errorf("insertion order not preserved: got %q, %q", children[0].Item.Name, children[1].Item.Name)
	}
}

func TestRemoveItem(t *testing.T) {
	_, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).RemoveItem(2).Build()

	if _, ok := nodeMap[2]; ok {
		t.Error("node 2 should be removed")
	}
	if _, ok := nodeMap[3]; ok {
		t.Error("descendant node 3 should be removed")
	}
	if _, ok := nodeMap[4]; ok {
		t.Error("descendant node 4 should be removed")
	}
	if len(nodeMap) != 2 {
		t.Errorf("expected 2 remaining nodes, got %d", len(nodeMap))
	}
}

func TestRemoveItemNotExist(t *testing.T) {
	_, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).RemoveItem(99).Build()
	if len(nodeMap) != 5 {
		t.Errorf("expected 5 nodes after removing non-existent key, got %d", len(nodeMap))
	}
}

func TestRemoveRoot(t *testing.T) {
	roots, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).RemoveItem(1).Build()
	if len(nodeMap) != 0 {
		t.Errorf("removing root should remove all nodes, got %d", len(nodeMap))
	}
	if len(roots) != 0 {
		t.Errorf("expected 0 roots, got %d", len(roots))
	}
}

func TestMoveItem(t *testing.T) {
	_, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).MoveItem(5, 2).Build()

	if len(nodeMap[1].Children) != 1 {
		t.Errorf("node 1 should have 1 child after move, got %d", len(nodeMap[1].Children))
	}
	if len(nodeMap[2].Children) != 3 {
		t.Errorf("node 2 should have 3 children after move, got %d", len(nodeMap[2].Children))
	}
}

func TestMoveItemSelfIsNoop(t *testing.T) {
	_, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).MoveItem(2, 2).Build()
	if len(nodeMap[1].Children) != 2 {
		t.Errorf("self-move should be no-op, node 1 should have 2 children, got %d", len(nodeMap[1].Children))
	}
}

func TestMoveItemCycleIsNoop(t *testing.T) {
	roots, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).MoveItem(1, 3).Build()
	if len(roots) != 1 {
		t.Errorf("cycle move should be no-op, expected 1 root, got %d", len(roots))
	}
	if len(nodeMap[1].Children) != 2 {
		t.Errorf("cycle move should be no-op, node 1 should have 2 children, got %d", len(nodeMap[1].Children))
	}
}

func TestMoveItemNotExist(t *testing.T) {
	_, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).MoveItem(99, 1).Build()
	if len(nodeMap) != 5 {
		t.Errorf("expected 5 nodes, got %d", len(nodeMap))
	}
}

func TestUpdateItem(t *testing.T) {
	_, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).
		UpdateItem(5, func(d *testItem) { d.Sort = 99 }).
		Build()

	if nodeMap[5].Item.Sort != 99 {
		t.Errorf("expected Sort=99, got %d", nodeMap[5].Item.Sort)
	}
}

func TestUpdateItemNilFnIsNoop(t *testing.T) {
	_, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).UpdateItem(5, nil).Build()
	if nodeMap[5].Item.Sort != 3 {
		t.Errorf("nil fn should not change item, got Sort=%d", nodeMap[5].Item.Sort)
	}
}

func TestUpdateItemNotExist(t *testing.T) {
	_, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).
		UpdateItem(99, func(d *testItem) { d.Sort = 0 }).
		Build()
	if len(nodeMap) != 5 {
		t.Errorf("expected 5 nodes, got %d", len(nodeMap))
	}
}

func TestUpdateItemKeyChange(t *testing.T) {
	b := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems())
	b.UpdateItem(5, func(d *testItem) { d.ID = 50 })
	_, nodeMap, _ := b.Build()

	if _, ok := nodeMap[5]; ok {
		t.Error("old key 5 should no longer be in nodeMap")
	}
	if nodeMap[50] == nil {
		t.Error("new key 50 should be in nodeMap")
	}
}

func TestFilter(t *testing.T) {
	_, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).
		Filter(func(d testItem) bool { return d.Sort%2 == 1 }).
		Build()

	for _, n := range nodeMap {
		if n.Item.Sort%2 != 1 {
			t.Errorf("filter failed: found Sort=%d", n.Item.Sort)
		}
	}
}

func TestFilterEmpty(t *testing.T) {
	roots, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).
		Filter(func(d testItem) bool { return false }).
		Build()
	if len(nodeMap) != 0 {
		t.Errorf("expected empty nodeMap, got %d", len(nodeMap))
	}
	if len(roots) != 0 {
		t.Errorf("expected empty roots, got %d", len(roots))
	}
}

func TestFilterInheritsKeyFunctions(t *testing.T) {
	filtered := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).
		Filter(func(d testItem) bool { return d.ID <= 2 })
	roots, nodeMap, _ := filtered.Build()

	if len(nodeMap) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodeMap))
	}
	_ = roots
}

func TestFilterNilPredicateRetainsAll(t *testing.T) {
	_, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).Filter(nil).Build()
	if len(nodeMap) != 5 {
		t.Errorf("nil predicate should retain all nodes, got %d", len(nodeMap))
	}
}

func TestTransform(t *testing.T) {
	_, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).
		Transform(func(d *testItem) { d.Sort = 42 }).
		Build()
	for _, n := range nodeMap {
		if n.Item.Sort != 42 {
			t.Errorf("Transform failed: expected Sort=42, got %d", n.Item.Sort)
		}
	}
}

func TestTransformNilIsNoop(t *testing.T) {
	_, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).Transform(nil).Build()
	if len(nodeMap) != 5 {
		t.Errorf("nil transform should be no-op, got %d", len(nodeMap))
	}
}

func TestCloneIndependence(t *testing.T) {
	original := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems())
	cloned := original.Clone()

	cloned.UpdateItem(1, func(d *testItem) { d.Sort = 999 })

	_, origMap, _ := original.Build()
	_, cloneMap, _ := cloned.Build()

	if origMap[1].Item.Sort == cloneMap[1].Item.Sort {
		t.Error("clone mutation should not affect original")
	}
	if cloneMap[1].Item.Sort != 999 {
		t.Errorf("clone should have Sort=999, got %d", cloneMap[1].Item.Sort)
	}
}

func TestCloneSharesFunctions(t *testing.T) {
	original := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems())
	cloned := original.Clone()

	_, cloneMap, _ := cloned.Build()
	if len(cloneMap) != 5 {
		t.Errorf("clone should have 5 nodes, got %d", len(cloneMap))
	}
}

func TestClonePreservesOverrides(t *testing.T) {
	original := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).MoveItem(5, 2)
	cloned := original.Clone()

	_, origMap, _ := original.Build()
	_, cloneMap, _ := cloned.Build()

	if len(origMap[2].Children) != len(cloneMap[2].Children) {
		t.Errorf("clone should preserve MoveItem overrides: orig=%d clone=%d",
			len(origMap[2].Children), len(cloneMap[2].Children))
	}
}

func TestValidateValid(t *testing.T) {
	errs := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).Validate()
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid tree, got %v", errs)
	}
}

func TestValidateCycle(t *testing.T) {
	items := []testItem{
		{ID: 10, ParentID: 20, Name: "A"},
		{ID: 20, ParentID: 10, Name: "B"},
	}
	b := NewBuilder[testItem]().
		KeyBy(func(d testItem) int { return d.ID }).
		ParentBy(func(d testItem) int { return d.ParentID }).
		WithItems(items)
	errs := b.Validate()
	if len(errs) == 0 {
		t.Error("expected cycle error, got none")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "cycle") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error mentioning 'cycle', got %v", errs)
	}
}

func TestValidateOrphan(t *testing.T) {
	items := []testItem{
		{ID: 1, ParentID: 1, Name: "root"},
		{ID: 2, ParentID: 99, Name: "orphan"},
	}
	b := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(items)
	errs := b.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "orphaned") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected orphan error, got %v", errs)
	}
}

func TestValidateNilKeyBy(t *testing.T) {
	b := NewBuilder[testItem]().WithItems(newItems())
	errs := b.Validate()
	if len(errs) == 0 {
		t.Error("expected error when KeyBy is not set")
	}
	if !errors.Is(errs[0], ErrKeyNotSet) {
		t.Errorf("expected ErrKeyNotSet, got %v", errs[0])
	}
}

func TestStatistics(t *testing.T) {
	stats := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).Statistics()

	if stats["totalNodes"] != 5 {
		t.Errorf("expected totalNodes=5, got %v", stats["totalNodes"])
	}
	if stats["rootNodes"] != 1 {
		t.Errorf("expected rootNodes=1, got %v", stats["rootNodes"])
	}
	if stats["maxDepth"] != 3 {
		t.Errorf("expected maxDepth=3, got %v", stats["maxDepth"])
	}
	if stats["leafNodes"] != 3 {
		t.Errorf("expected leafNodes=3, got %v", stats["leafNodes"])
	}
	avg, ok := stats["avgChildrenPerNode"].(float64)
	if !ok || avg <= 0 {
		t.Errorf("expected avgChildrenPerNode > 0, got %v", stats["avgChildrenPerNode"])
	}
}

func TestStatisticsEmpty(t *testing.T) {
	stats := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(nil).Statistics()
	if stats["totalNodes"] != 0 {
		t.Errorf("expected totalNodes=0, got %v", stats["totalNodes"])
	}
	if stats["maxDepth"] != 0 {
		t.Errorf("expected maxDepth=0, got %v", stats["maxDepth"])
	}
}

func TestStableSortEqualSortVal(t *testing.T) {
	items := []testItem{
		{ID: 1, ParentID: 1, Sort: 0, Name: "root"},
		{ID: 2, ParentID: 1, Sort: 5, Name: "first"},
		{ID: 3, ParentID: 1, Sort: 5, Name: "second"},
		{ID: 4, ParentID: 1, Sort: 5, Name: "third"},
	}
	roots, _, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(items).Build()
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	children := roots[0].Children
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}
	want := []string{"first", "second", "third"}
	for i, child := range children {
		if child.Item.Name != want[i] {
			t.Errorf("position %d: want %q, got %q", i, want[i], child.Item.Name)
		}
	}
}

func TestInsertionOrderWithoutSortBy(t *testing.T) {
	b := NewBuilder[testItem]().
		KeyBy(func(d testItem) int { return d.ID }).
		ParentBy(func(d testItem) int { return d.ParentID }).
		WithItems(nil)
	b.AddItem(testItem{ID: 1, ParentID: 1, Name: "root"})
	b.AddItem(testItem{ID: 2, ParentID: 1, Name: "alpha"})
	b.AddItem(testItem{ID: 3, ParentID: 1, Name: "beta"})
	b.AddItem(testItem{ID: 4, ParentID: 1, Name: "gamma"})

	roots, _, _ := b.Build()
	children := roots[0].Children
	want := []string{"alpha", "beta", "gamma"}
	for i, child := range children {
		if child.Item.Name != want[i] {
			t.Errorf("position %d: want %q, got %q", i, want[i], child.Item.Name)
		}
	}
}

func TestSelfReferenceRoot(t *testing.T) {
	items := []testItem{
		{ID: 10, ParentID: 10, Name: "self-ref-root"},
		{ID: 20, ParentID: 10, Name: "child"},
	}
	roots, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(items).Build()
	if len(roots) != 1 {
		t.Errorf("expected 1 root, got %d", len(roots))
	}
	if roots[0].Item.ID != 10 {
		t.Errorf("expected root ID=10, got %d", roots[0].Item.ID)
	}
	if len(nodeMap[10].Children) != 1 {
		t.Errorf("expected 1 child of root, got %d", len(nodeMap[10].Children))
	}
}

func TestKeyByAfterWithItems(t *testing.T) {
	b := NewBuilder[testItem]().
		KeyBy(func(d testItem) int { return d.ID }).
		ParentBy(func(d testItem) int { return d.ParentID }).
		SortBy(func(d testItem) int { return d.Sort }).
		WithItems(newItems())
	roots, nodeMap, _ := b.Build()
	if len(roots) != 1 {
		t.Errorf("expected 1 root, got %d", len(roots))
	}
	if len(nodeMap) != 5 {
		t.Errorf("expected 5 nodes, got %d", len(nodeMap))
	}
}

func TestDuplicateKey(t *testing.T) {
	_, err := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems([]testItem{
		{ID: 1, ParentID: 1},
		{ID: 1, ParentID: 1},
	}).Build()
	if err == nil {
		t.Error("expected error on duplicate key")
	}
}

func TestAddItemDuplicateKey(t *testing.T) {
	b := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems([]testItem{{ID: 1, ParentID: 1}})
	_, err := b.AddItem(testItem{ID: 1, ParentID: 1}).Build()
	if err == nil {
		t.Error("expected error on duplicate key")
	}
}

func TestMap(t *testing.T) {
	_, nodeMap, _ := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).
		Map(func(d testItem) testItem {
			d.Name = strings.ToUpper(d.Name)
			return d
		}, func(d testItem) int { return d.ID }).
		Build()

	for _, n := range nodeMap {
		if n.Item.Name != strings.ToUpper(n.Item.Name) {
			t.Errorf("Map failed: expected uppercase name, got %q", n.Item.Name)
		}
	}
}

func TestMapPreservesTreeStructure(t *testing.T) {
	b := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems())
	mapped := b.Map(func(d testItem) testItem { return d }, func(d testItem) int { return d.ID })
	roots, nodeMap, _ := mapped.Build()

	if len(roots) != 1 {
		t.Errorf("expected 1 root, got %d", len(roots))
	}
	if len(nodeMap) != 5 {
		t.Errorf("expected 5 nodes, got %d", len(nodeMap))
	}
	if len(nodeMap[2].Children) != 2 {
		t.Errorf("expected node 2 to have 2 children, got %d", len(nodeMap[2].Children))
	}
}

func TestFind(t *testing.T) {
	b := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems())
	node := b.Find(func(d testItem) bool { return d.Name == "child-A" })
	if node == nil {
		t.Fatal("Find returned nil")
	}
	if node.Item.ID != 2 {
		t.Errorf("expected to find node with ID=2, got ID=%d", node.Item.ID)
	}
}

func TestFindNotFound(t *testing.T) {
	b := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems())
	node := b.Find(func(d testItem) bool { return d.Name == "nonexistent" })
	if node != nil {
		t.Error("Find should return nil for non-matching predicate")
	}
}

func TestContains(t *testing.T) {
	b := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems())
	if !b.Contains(1) {
		t.Error("Contains(1) should return true")
	}
	if !b.Contains(5) {
		t.Error("Contains(5) should return true")
	}
	if b.Contains(99) {
		t.Error("Contains(99) should return false")
	}
}

func TestDepth(t *testing.T) {
	b := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems())
	if b.Depth(1) != 1 {
		t.Errorf("root depth should be 1, got %d", b.Depth(1))
	}
	if b.Depth(2) != 2 {
		t.Errorf("child depth should be 2, got %d", b.Depth(2))
	}
	if b.Depth(3) != 3 {
		t.Errorf("grandchild depth should be 3, got %d", b.Depth(3))
	}
	if b.Depth(99) != 0 {
		t.Errorf("non-existent node depth should be 0, got %d", b.Depth(99))
	}
}

func TestMapPreservesOverrides(t *testing.T) {
	original := newBuilder().KeyBy(func(d testItem) int { return d.ID }).WithItems(newItems()).MoveItem(5, 2)
	mapped := original.Map(func(d testItem) testItem { return d }, func(d testItem) int { return d.ID })

	_, origMap, _ := original.Build()
	_, mapMap, _ := mapped.Build()

	if len(origMap[2].Children) != len(mapMap[2].Children) {
		t.Errorf("Map should preserve MoveItem overrides: orig=%d mapped=%d",
			len(origMap[2].Children), len(mapMap[2].Children))
	}
}
