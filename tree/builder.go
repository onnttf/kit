package tree

import (
	"cmp"
	"fmt"
	"slices"
	"sync"
)

type item[T any, K comparable] struct {
	data        T
	parentKey   K
	hasParent   bool
	insertOrder int
}

type Builder[T any, K comparable] struct {
	mu    sync.RWMutex
	items []*item[T, K]

	insertCtr int

	keyFn        func(T) K
	parentFn     func(T) (K, bool)
	sortFn       func(T) int
	sortCmpFn    func(T, T) int
	sortByInsert bool

	dirty  bool
	cached *Tree[T, K]
}

func NewBuilder[T any, K comparable]() *Builder[T, K] {
	return &Builder[T, K]{dirty: true}
}

func (b *Builder[T, K]) resolveParent(n *item[T, K], selfKey K) (K, bool) {
	if n.hasParent {
		if n.parentKey == selfKey {
			var zero K
			return zero, false
		}
		return n.parentKey, true
	}
	if b.parentFn != nil {
		pk, ok := b.parentFn(n.data)
		if ok && pk == selfKey {
			var zero K
			return zero, false
		}
		return pk, ok
	}
	return n.parentKey, n.hasParent
}

func (b *Builder[T, K]) ensureTree() (*Tree[T, K], error) {
	b.mu.RLock()
	if !b.dirty && b.cached != nil {
		t := b.cached
		b.mu.RUnlock()
		return t, nil
	}
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.keyFn == nil {
		return nil, ErrKeyNotSet
	}

	if !b.dirty && b.cached != nil {
		return b.cached, nil
	}

	t, err := b.buildTree()
	if err != nil {
		return nil, err
	}

	b.cached = t
	b.dirty = false
	return t, nil
}

func (b *Builder[T, K]) buildTree() (*Tree[T, K], error) {
	count := len(b.items)

	keys := make([]K, count)
	keyIndex := make(map[K]int, count)

	for i, n := range b.items {
		k := b.keyFn(n.data)
		if _, ok := keyIndex[k]; ok {
			return nil, fmt.Errorf("%w: %v", ErrDuplicateKey, k)
		}
		keys[i] = k
		keyIndex[k] = i
	}

	parentKeys := make([]K, count)
	hasParents := make([]bool, count)
	sortVals := make([]int, count)

	for i, n := range b.items {
		pk, has := b.resolveParent(n, keys[i])
		parentKeys[i] = pk
		hasParents[i] = has

		if b.sortCmpFn == nil {
			if b.sortFn != nil {
				sortVals[i] = b.sortFn(n.data)
			} else {
				sortVals[i] = n.insertOrder
			}
		}
	}

	if err := validateTree(keys, parentKeys, hasParents, keyIndex); err != nil {
		return nil, err
	}

	cache := make(map[K]*Node[T], count)
	nodes := make([]*Node[T], count)
	nodeSort := make(map[*Node[T]]int, count)

	for i, n := range b.items {
		out := &Node[T]{Item: n.data}
		nodes[i] = out
		cache[keys[i]] = out
		nodeSort[out] = sortVals[i]
	}

	parentIdx := make(map[K]K, count)
	var roots []*Node[T]

	for i := range b.items {
		out := nodes[i]

		if !hasParents[i] {
			roots = append(roots, out)
			continue
		}

		pk := parentKeys[i]
		parentIdx[keys[i]] = pk
		cache[pk].Children = append(cache[pk].Children, out)
	}

	if b.sortCmpFn != nil {
		sortForestWithCmp(roots, b.sortCmpFn)
	} else {
		sortForest(roots, nodeSort)
	}

	assignLevels(roots, 1)

	return &Tree[T, K]{
		roots:      roots,
		cache:      cache,
		parentIdx:  parentIdx,
		keyFn:      b.keyFn,
		subtreeIdx: make(map[K]*Tree[T, K]),
	}, nil
}

func assignLevels[T any](nodes []*Node[T], level int) {
	for _, n := range nodes {
		n.Level = level
		if len(n.Children) > 0 {
			assignLevels(n.Children, level+1)
		}
	}
}

func (b *Builder[T, K]) AddItem(v T) {
	b.AddItemWithParent(v, *new(K))
}

func (b *Builder[T, K]) AddItemWithParent(v T, parentKey K) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var hasParent bool
	if parentKey != *new(K) {
		hasParent = true
	}

	b.items = append(b.items, &item[T, K]{
		data:        v,
		parentKey:   parentKey,
		hasParent:   hasParent,
		insertOrder: b.insertCtr,
	})
	b.insertCtr++
	b.invalidate()
}

func (b *Builder[T, K]) WithItems(items []T) *Builder[T, K] {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, v := range items {
		b.items = append(b.items, &item[T, K]{data: v, insertOrder: b.insertCtr})
		b.insertCtr++
	}
	b.invalidate()
	return b
}

func (b *Builder[T, K]) KeyBy(fn func(T) K) *Builder[T, K] {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.keyFn = fn
	b.invalidate()
	return b
}

func (b *Builder[T, K]) ParentBy(fn func(T) (K, bool)) *Builder[T, K] {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.parentFn = fn
	b.invalidate()
	return b
}

func (b *Builder[T, K]) SortBy(fn func(T) int) *Builder[T, K] {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sortFn = fn
	b.sortByInsert = false
	b.invalidate()
	return b
}

func (b *Builder[T, K]) SortByFunc(fn func(T, T) int) *Builder[T, K] {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sortCmpFn = fn
	b.sortByInsert = false
	b.invalidate()
	return b
}

func (b *Builder[T, K]) invalidate() {
	b.dirty = true
	b.cached = nil
}

func (b *Builder[T, K]) Build() (*Tree[T, K], error) {
	return b.ensureTree()
}

func (b *Builder[T, K]) Clone() *Builder[T, K] {
	b.mu.Lock()
	defer b.mu.Unlock()

	items := make([]*item[T, K], len(b.items))
	for i, n := range b.items {
		items[i] = &item[T, K]{
			data:        n.data,
			parentKey:   n.parentKey,
			hasParent:   n.hasParent,
			insertOrder: n.insertOrder,
		}
	}

	return &Builder[T, K]{
		items:        items,
		insertCtr:    b.insertCtr,
		keyFn:        b.keyFn,
		parentFn:     b.parentFn,
		sortFn:       b.sortFn,
		sortCmpFn:    b.sortCmpFn,
		sortByInsert: b.sortByInsert,
		dirty:        true,
	}
}

func (b *Builder[T, K]) Statistics() (Stats, error) {
	tree, err := b.Build()
	if err != nil {
		return Stats{}, err
	}
	return tree.Stats(), nil
}

func (b *Builder[T, K]) Validate() []error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.keyFn == nil {
		return []error{ErrKeyNotSet}
	}

	var errs []error

	count := len(b.items)
	keys := make([]K, count)
	keyIndex := make(map[K]int, count)
	parentKeys := make([]K, count)
	hasParents := make([]bool, count)

	for i, n := range b.items {
		k := b.keyFn(n.data)
		keys[i] = k

		if _, ok := keyIndex[k]; ok {
			errs = append(errs, fmt.Errorf("%w: %v", ErrDuplicateKey, k))
		}
		keyIndex[k] = i

		pk, has := b.resolveParent(n, k)
		parentKeys[i] = pk
		hasParents[i] = has
	}

	for i, k := range keys {
		if hasParents[i] {
			if _, ok := keyIndex[parentKeys[i]]; !ok {
				errs = append(errs, fmt.Errorf("%w: %v", ErrOrphanedNode, k))
			}
		}
	}

	type color uint8
	const (
		colorWhite color = iota
		colorGray
		colorBlack
	)

	colors := make(map[K]color, len(keys))
	for _, start := range keys {
		if colors[start] != colorWhite {
			continue
		}

		cur := start
		visited := []K{}
		for colors[cur] == colorWhite {
			colors[cur] = colorGray
			visited = append(visited, cur)

			idx, ok := keyIndex[cur]
			if !ok || !hasParents[idx] {
				break
			}

			if colors[parentKeys[idx]] == colorGray {
				errs = append(errs, fmt.Errorf("%w: %v", ErrCycle, parentKeys[idx]))
				break
			}
			cur = parentKeys[idx]
		}

		for _, k := range visited {
			colors[k] = colorBlack
		}
	}

	return errs
}

func (b *Builder[T, K]) Filter(fn func(T) bool) *Builder[T, K] {
	b.mu.Lock()
	defer b.mu.Unlock()

	var filtered []*item[T, K]
	for _, n := range b.items {
		if fn(n.data) {
			filtered = append(filtered, n)
		}
	}

	clone := &Builder[T, K]{
		items:        filtered,
		insertCtr:    b.insertCtr,
		keyFn:        b.keyFn,
		parentFn:     b.parentFn,
		sortFn:       b.sortFn,
		sortCmpFn:    b.sortCmpFn,
		sortByInsert: b.sortByInsert,
		dirty:        true,
	}
	return clone
}

func (b *Builder[T, K]) Map(fn func(T) T, keyFn func(T) K) *Builder[T, K] {
	b.mu.Lock()
	defer b.mu.Unlock()

	mapped := make([]*item[T, K], len(b.items))
	for i, n := range b.items {
		mapped[i] = &item[T, K]{
			data:        fn(n.data),
			parentKey:   n.parentKey,
			hasParent:   n.hasParent,
			insertOrder: n.insertOrder,
		}
	}

	return &Builder[T, K]{
		items:        mapped,
		insertCtr:    b.insertCtr,
		keyFn:        keyFn,
		parentFn:     b.parentFn,
		sortFn:       b.sortFn,
		sortCmpFn:    b.sortCmpFn,
		sortByInsert: b.sortByInsert,
		dirty:        true,
	}
}

func (b *Builder[T, K]) Transform(fn func(*T)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, n := range b.items {
		fn(&n.data)
	}
	b.invalidate()
}

func (b *Builder[T, K]) Find(fn func(T) bool) (*Node[T], error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, n := range b.items {
		if fn(n.data) {
			return &Node[T]{Item: n.data}, nil
		}
	}
	return nil, fmt.Errorf("item not found")
}

func (b *Builder[T, K]) Contains(key K) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.keyFn == nil {
		return false, nil
	}

	for _, n := range b.items {
		if b.keyFn(n.data) == key {
			return true, nil
		}
	}
	return false, nil
}

func (b *Builder[T, K]) UpdateItem(key K, fn func(*T)) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	idx := -1
	for i, n := range b.items {
		if b.keyFn(n.data) == key {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("key not found: %v", key)
	}

	oldItem := b.items[idx].data
	fn(&b.items[idx].data)
	newKey := b.keyFn(b.items[idx].data)

	if newKey != key {
		for i, n := range b.items {
			if i != idx && b.keyFn(n.data) == newKey {
				b.items[idx].data = oldItem
				return fmt.Errorf("%w: %v", ErrDuplicateKey, newKey)
			}
		}
	}

	b.invalidate()
	return nil
}

func (b *Builder[T, K]) ChildrenOf(key K) ([]*Node[T], error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	tree, err := b.buildTree()
	if err != nil {
		return nil, err
	}

	parentNode, ok := tree.cache[key]
	if !ok {
		return nil, fmt.Errorf("key not found: %v", key)
	}

	result := make([]*Node[T], len(parentNode.Children))
	for i, c := range parentNode.Children {
		result[i] = cloneNode(c)
	}
	return result, nil
}

func (b *Builder[T, K]) Depth(key K) (int, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	tree, err := b.buildTree()
	if err != nil {
		return 0, err
	}

	node, ok := tree.cache[key]
	if !ok {
		return 0, fmt.Errorf("key not found: %v", key)
	}

	return node.Level, nil
}

func (b *Builder[T, K]) IsDescendant(ancestor, key K) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	tree, err := b.buildTree()
	if err != nil {
		return false, err
	}

	if _, ok := tree.cache[key]; !ok {
		return false, fmt.Errorf("key not found: %v", key)
	}
	if _, ok := tree.cache[ancestor]; !ok {
		return false, fmt.Errorf("ancestor not found: %v", ancestor)
	}

	if key == ancestor {
		return true, nil
	}

	cur := key
	for {
		pk, ok := tree.parentIdx[cur]
		if !ok {
			return false, nil
		}
		if pk == ancestor {
			return true, nil
		}
		cur = pk
	}
}

func (b *Builder[T, K]) RemoveItem(key K) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	keyIndex := make(map[K]int, len(b.items))
	for i, n := range b.items {
		keyIndex[b.keyFn(n.data)] = i
	}

	if _, ok := keyIndex[key]; !ok {
		return fmt.Errorf("key not found: %v", key)
	}

	keysToRemove := make(map[K]bool)
	var collect func(k K)
	collect = func(k K) {
		keysToRemove[k] = true
		for _, n := range b.items {
			pk, has := b.resolveParent(n, b.keyFn(n.data))
			if has && pk == k {
				collect(b.keyFn(n.data))
			}
		}
	}
	collect(key)

	var remaining []*item[T, K]
	for _, n := range b.items {
		if !keysToRemove[b.keyFn(n.data)] {
			remaining = append(remaining, n)
		}
	}

	b.items = remaining
	b.invalidate()
	return nil
}

func (b *Builder[T, K]) MoveItem(key, newParent K) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if key == newParent {
		return fmt.Errorf("self-move not allowed")
	}

	keyIndex := make(map[K]int, len(b.items))
	for i, n := range b.items {
		k := b.keyFn(n.data)
		keyIndex[k] = i
	}

	if _, ok := keyIndex[key]; !ok {
		return fmt.Errorf("key not found: %v", key)
	}
	if _, ok := keyIndex[newParent]; !ok {
		return fmt.Errorf("parent key not found: %v", newParent)
	}

	ancestors := make(map[K]bool)
	cur := newParent
	for {
		ancestors[cur] = true
		idx, ok := keyIndex[cur]
		if !ok {
			break
		}
		n := b.items[idx]
		pk, has := b.resolveParent(n, b.keyFn(n.data))
		if !has {
			break
		}
		cur = pk
	}

	if ancestors[key] {
		return fmt.Errorf("%w: %v", ErrCycle, key)
	}

	for _, n := range b.items {
		if b.keyFn(n.data) == key {
			n.parentKey = newParent
			n.hasParent = true
			break
		}
	}

	b.invalidate()
	return nil
}

func (b *Builder[T, K]) Subtree(key K) (*Tree[T, K], error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	tree, err := b.buildTree()
	if err != nil {
		return nil, err
	}

	subtree, ok := tree.Subtree(key)
	if !ok {
		return nil, fmt.Errorf("key not found: %v", key)
	}
	return subtree, nil
}

func (b *Builder[T, K]) Orphans() ([]*Node[T], error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	tree, err := b.buildTree()
	if err != nil {
		return nil, err
	}

	return tree.Orphans(), nil
}

func validateTree[K comparable](
	keys []K,
	parentKeys []K,
	hasParents []bool,
	keyIndex map[K]int,
) error {
	for i, k := range keys {
		if hasParents[i] {
			if _, ok := keyIndex[parentKeys[i]]; !ok {
				return fmt.Errorf("%w: %v", ErrOrphanedNode, k)
			}
		}
	}

	type color uint8
	const (
		colorWhite color = iota
		colorGray
		colorBlack
	)

	colors := make(map[K]color, len(keys))

	for _, start := range keys {
		if colors[start] != colorWhite {
			continue
		}

		cur := start
		visited := []K{}
		for colors[cur] == colorWhite {
			colors[cur] = colorGray
			visited = append(visited, cur)

			idx, ok := keyIndex[cur]
			if !ok || !hasParents[idx] {
				break
			}

			if colors[parentKeys[idx]] == colorGray {
				return fmt.Errorf("%w: %v", ErrCycle, parentKeys[idx])
			}
			cur = parentKeys[idx]
		}

		for _, k := range visited {
			colors[k] = colorBlack
		}
	}

	return nil
}

func sortForest[T any](roots []*Node[T], sortVals map[*Node[T]]int) {
	type frame struct{ nodes []*Node[T] }
	stack := []frame{{roots}}

	for len(stack) > 0 {
		f := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if len(f.nodes) > 1 {
			slices.SortStableFunc(f.nodes, func(a, b *Node[T]) int {
				return cmp.Compare(sortVals[a], sortVals[b])
			})
		}

		for _, n := range f.nodes {
			if len(n.Children) > 0 {
				stack = append(stack, frame{n.Children})
			}
		}
	}
}

func sortForestWithCmp[T any](roots []*Node[T], cmpFn func(T, T) int) {
	type frame struct{ nodes []*Node[T] }
	stack := []frame{{roots}}

	for len(stack) > 0 {
		f := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if len(f.nodes) > 1 {
			slices.SortStableFunc(f.nodes, func(a, b *Node[T]) int {
				return cmpFn(a.Item, b.Item)
			})
		}

		for _, n := range f.nodes {
			if len(n.Children) > 0 {
				stack = append(stack, frame{n.Children})
			}
		}
	}
}
