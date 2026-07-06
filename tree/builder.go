package tree

import (
	"cmp"
	"fmt"
	"slices"
)

type Builder[T any, K comparable] struct {
	items    []builderItem[T, K]
	keyFn    func(T) K
	parentFn func(T) (K, bool)
	sortFn   func(a, b T) int
	ops      []operation[T, K]
}

func NewBuilder[T any, K comparable]() *Builder[T, K] {
	return &Builder[T, K]{}
}

func (b *Builder[T, K]) KeyBy(fn func(T) K) *Builder[T, K] {
	b.keyFn = fn
	return b
}

func (b *Builder[T, K]) ParentBy(fn func(T) (K, bool)) *Builder[T, K] {
	b.parentFn = fn
	return b
}

func (b *Builder[T, K]) SortFunc(fn func(a, b T) int) *Builder[T, K] {
	b.sortFn = fn
	return b
}

func (b *Builder[T, K]) Add(items ...T) *Builder[T, K] {
	for _, v := range items {
		b.items = append(b.items, builderItem[T, K]{value: v, insert: len(b.items)})
	}
	return b
}

func (b *Builder[T, K]) Update(key K, fn func(*T)) *Builder[T, K] {
	b.ops = append(b.ops, updateOperation(key, fn))
	return b
}

func (b *Builder[T, K]) Remove(key K) *Builder[T, K] {
	b.ops = append(b.ops, removeOperation[T](key))
	return b
}

func (b *Builder[T, K]) Move(key, parent K) *Builder[T, K] {
	b.ops = append(b.ops, moveOperation[T](key, parent))
	return b
}

func (b *Builder[T, K]) Validate() []error {
	_, err := b.build()
	if err == nil {
		return nil
	}
	return []error{err}
}

func (b *Builder[T, K]) Build() (*Tree[T, K], error) {
	return b.build()
}

func (b *Builder[T, K]) Clone() *Builder[T, K] {
	return &Builder[T, K]{
		items:    append([]builderItem[T, K](nil), b.items...),
		keyFn:    b.keyFn,
		parentFn: b.parentFn,
		sortFn:   b.sortFn,
		ops:      append([]operation[T, K](nil), b.ops...),
	}
}

func (b *Builder[T, K]) build() (*Tree[T, K], error) {
	if b.keyFn == nil {
		return nil, ErrKeyNotSet
	}

	items := append([]builderItem[T, K](nil), b.items...)
	for i := range items {
		if items[i].hasParent || b.parentFn == nil {
			continue
		}
		parent, ok := b.parentFn(items[i].value)
		items[i].parent = parent
		items[i].hasParent = ok
	}

	state := builderState[T, K]{items: items, keyFn: b.keyFn}
	for _, op := range b.ops {
		if err := op(&state); err != nil {
			return nil, err
		}
	}
	return buildTree(state.items, b.keyFn, b.sortFn)
}

type builderItem[T any, K comparable] struct {
	value      T
	parent     K
	hasParent  bool
	insert     int
	hasDeleted bool
}

type operation[T any, K comparable] func(*builderState[T, K]) error

type builderState[T any, K comparable] struct {
	items []builderItem[T, K]
	keyFn func(T) K
}

func updateOperation[T any, K comparable](key K, fn func(*T)) operation[T, K] {
	return func(state *builderState[T, K]) error {
		if fn == nil {
			return fmt.Errorf("%w: nil update function", ErrInvalidInput)
		}
		idx := state.index(key)
		if idx < 0 {
			return fmt.Errorf("%w: key=%v", ErrKeyNotFound, key)
		}
		fn(&state.items[idx].value)
		return nil
	}
}

func removeOperation[T any, K comparable](key K) operation[T, K] {
	return func(state *builderState[T, K]) error {
		if state.index(key) < 0 {
			return fmt.Errorf("%w: key=%v", ErrKeyNotFound, key)
		}

		children := state.childIndex()
		toRemove := make(map[K]struct{})
		stack := append([]K(nil), children[key]...)
		for len(stack) > 0 {
			k := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if _, ok := toRemove[k]; ok {
				continue
			}
			toRemove[k] = struct{}{}
			stack = append(stack, children[k]...)
		}
		toRemove[key] = struct{}{}
		for i := range state.items {
			if _, ok := toRemove[state.keyFn(state.items[i].value)]; ok {
				state.items[i].hasDeleted = true
			}
		}
		return nil
	}
}

func moveOperation[T any, K comparable](key, parent K) operation[T, K] {
	return func(state *builderState[T, K]) error {
		idx := state.index(key)
		if idx < 0 {
			return fmt.Errorf("%w: key=%v", ErrKeyNotFound, key)
		}
		if key == parent {
			return fmt.Errorf("%w: key=%v parent=%v", ErrCycle, key, parent)
		}
		if state.index(parent) < 0 {
			return fmt.Errorf("%w: parent=%v", ErrMissingParent, parent)
		}
		state.items[idx].parent = parent
		state.items[idx].hasParent = true
		return nil
	}
}

func (s *builderState[T, K]) index(key K) int {
	for i, item := range s.items {
		if item.hasDeleted {
			continue
		}
		if s.keyFn(item.value) == key {
			return i
		}
	}
	return -1
}

func (s *builderState[T, K]) childIndex() map[K][]K {
	out := make(map[K][]K)
	for _, item := range s.items {
		if item.hasDeleted || !item.hasParent {
			continue
		}
		out[item.parent] = append(out[item.parent], s.keyFn(item.value))
	}
	return out
}

func buildTree[T any, K comparable](
	items []builderItem[T, K],
	keyFn func(T) K,
	sortFn func(a, b T) int,
) (*Tree[T, K], error) {
	nodes := make(map[K]Node[T, K], len(items))
	order := make(map[K]int, len(items))
	parentBy := make(map[K]K, len(items))
	hasParent := make(map[K]bool, len(items))
	childrenBy := make(map[K][]K, len(items))
	var live []builderItem[T, K]

	for _, item := range items {
		if item.hasDeleted {
			continue
		}
		key := keyFn(item.value)
		if _, ok := nodes[key]; ok {
			return nil, fmt.Errorf("%w: key=%v", ErrDuplicateKey, key)
		}

		nodes[key] = Node[T, K]{
			Key:    key,
			Parent: cloneParent(item.parent, item.hasParent),
			Value:  item.value,
		}
		order[key] = item.insert
		live = append(live, item)

		if item.hasParent {
			parentBy[key] = item.parent
			hasParent[key] = true
			childrenBy[item.parent] = append(childrenBy[item.parent], key)
		}
	}

	for key, parent := range parentBy {
		if parent == key {
			return nil, fmt.Errorf("%w: key=%v parent=%v", ErrCycle, key, parent)
		}
		if _, ok := nodes[parent]; !ok {
			return nil, fmt.Errorf("%w: parent=%v", ErrMissingParent, parent)
		}
		if cycleKey, ok := detectCycleFrom(key, parentBy, hasParent); ok {
			return nil, fmt.Errorf("%w: key=%v", ErrCycle, cycleKey)
		}
	}

	var buildNode func(K) Node[T, K]
	buildNode = func(key K) Node[T, K] {
		node := nodes[key]
		for _, childKey := range childrenBy[key] {
			node.Children = append(node.Children, buildNode(childKey))
		}
		return node
	}

	var roots []Node[T, K]
	for _, item := range live {
		key := keyFn(item.value)
		if !item.hasParent {
			roots = append(roots, buildNode(key))
		}
	}

	sortNodes(roots, order, sortFn)
	assignDepth(roots, 0)
	return newTree(roots), nil
}

func cloneParent[K comparable](parent K, ok bool) *K {
	if !ok {
		return nil
	}
	p := parent
	return &p
}

func sortNodes[T any, K comparable](nodes []Node[T, K], order map[K]int, sortFn func(a, b T) int) {
	slices.SortStableFunc(nodes, func(a, b Node[T, K]) int {
		if sortFn != nil {
			if c := sortFn(a.Value, b.Value); c != 0 {
				return c
			}
		}
		return cmp.Compare(order[a.Key], order[b.Key])
	})
	for i := range nodes {
		sortNodes(nodes[i].Children, order, sortFn)
	}
}

func assignDepth[T any, K comparable](nodes []Node[T, K], depth int) {
	for i := range nodes {
		nodes[i].Depth = depth
		for j := range nodes[i].Children {
			parent := nodes[i].Key
			nodes[i].Children[j].Parent = &parent
		}
		assignDepth(nodes[i].Children, depth+1)
	}
}

func detectCycleFrom[K comparable](key K, parentBy map[K]K, hasParent map[K]bool) (K, bool) {
	var zero K
	seen := make(map[K]bool)
	cur := key
	for hasParent[cur] {
		if seen[cur] {
			return cur, true
		}
		seen[cur] = true
		cur = parentBy[cur]
	}
	return zero, false
}
