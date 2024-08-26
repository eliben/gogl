package btree

type BTree[K, V any] struct {
	cmp func(K, K) int

	root *node[K, V]
}

// tee (t) is the "minimal degree" of the B-tree. Every node other than
// the root must have between t-1 and 2t-1 keys (inclusive).
// Nodes with 2t-1 keys are considered "full".
const tee = 10

type node[K, V any] struct {
	keys     []nodeKey[K, V]
	children []*node[K, V]
	leaf     bool
}

// nodeKey is a pair of key, value
type nodeKey[K, V any] struct {
	key   K
	value V
}

// New creates a new B-tree with the given comparison function. cmp(a, b)
// should return a negative number when a<b, a positive number when a>b and
// zero when a==b.
func New[K, V any](cmp func(K, K) int) *BTree[K, V] {
	return &BTree[K, V]{
		cmp: cmp,
		root: &node[K, V]{
			leaf: true,
		},
	}
}

// Get looks for the given key in the tree. It returns the associated value
// and ok=true; otherwise, it returns ok=false.
func (bt *BTree[K, V]) Get(key K) (v V, ok bool) {
	return bt.getFromNode(key, bt.root)
}

func (bt *BTree[K, V]) getFromNode(key K, n *node[K, V]) (v V, ok bool) {
	// Find the key in this node's keys, or the place where this key would
	// fit so we can follow a child link.
	// TODO: use slices.BinarySearchFunc?
	i := 0
	for i < len(n.keys) && bt.cmp(key, n.keys[i].key) > 0 {
		i++
	}
	if i < len(n.keys) && bt.cmp(key, n.keys[i].key) == 0 {
		// Found the key in this node!
		return n.keys[i].value, true
	}

	if n.leaf {
		return *new(V), false
	}

	return bt.getFromNode(key, n.children[i])
}
