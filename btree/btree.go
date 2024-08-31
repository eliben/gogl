package btree

import "slices"

type BTree[K, V any] struct {
	cmp func(K, K) int

	root *node[K, V]
}

// tee (t) is the "minimal degree" of the B-tree. Every node other than
// the root must have between t-1 and 2t-1 keys (inclusive).
// Nodes with 2t-1 keys are considered "full".
const tee = 10
const teeFull = tee*2 - 1

type node[K, V any] struct {
	// Key ordering invariants (defining n=len(keys)):
	//
	// keys[i].key <= keys[i+1].key for each i in 0..n-2
	//
	// There are n+1 elements in children (0..n)
	// if kj is any key in children[j], then:  keys[j-1].key <= kj <= keys[j].key
	// Boundaries: k0 <= keys[0].key   and   keys[n-1].key <= kn
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
		// This is a leaf node and the key wasn't found yet; return ok=false.
		return *new(V), false
	}

	// define the relationship between the two slices in a docstring in node
	// Because of the first loop, i is the first index where key <= n.keys[i]
	// (it could also be len(n.keys)), so using the key ordering invariant we
	// recurse into children[i]
	return bt.getFromNode(key, n.children[i])
}

// splitChild splits n.children[i] into two children nodes and moves the
// median key of n.children[i] into n. It assumes that n isn't full,
// but n.children[i] is full.
func (bt *BTree[K, V]) splitChild(n *node[K, V], i int) {
	// Notation: y is the i-th child of n (the one being split), and z is the
	// new node created to adopt y's t-1 largest keys.
	y := n.children[i]
	if len(n.keys) == teeFull || len(y.keys) != teeFull {
		panic("expect n to be non-full and y to be full")
	}
	z := &node[K, V]{
		leaf: y.leaf,
	}

	// Move keys to z.
	// Before the move, y.keys:
	//
	//   k[0]  k[1] ... k[t-2]  k[t-1]  k[t]  k[t+1] ... k[2t-2]
	//
	// k[t-1] is the median key -- it will move to n.
	// k[0]..k[t-2] will stay with y
	// k[t]..k[2t-2] will move to z
	z.keys = make([]nodeKey[K, V], tee-1)
	medianKey := y.keys[tee-1]
	copy(z.keys, y.keys[tee:])
	y.keys = y.keys[:tee-1]

	// Move children to z.
	//
	// The first t children stay with y; the other t children move to z.
	if !y.leaf {
		z.children = make([]*node[K, V], tee)
		copy(z.children, y.children[tee:])
		y.children = y.children[:tee]
	}

	// Place the median key in n
	n.keys = slices.Insert(n.keys, i, medianKey)

	// Place the pointer to z after the pointer to y in n
	n.children = slices.Insert(n.children, i+1, z)
}
