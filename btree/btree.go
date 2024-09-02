package btree

import (
	"fmt"
	"slices"
)

type BTree[K, V any] struct {
	cmp func(K, K) int

	root *node[K, V]

	// tee (t) is the "minimal degree" of the B-tree. Every node other than
	// the root must have between t-1 and 2t-1 keys (inclusive).
	// Nodes with 2t-1 keys are considered "full".
	tee int
}

const defaultTee = 10

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

// New creates a new B-tree with the given comparison function, with the
// default branching factor tee. cmp(a, b) should return a negative number when
// a<b, a positive number when a>b and zero when a==b.
func New[K, V any](cmp func(K, K) int) *BTree[K, V] {
	return NewWithTee[K, V](cmp, defaultTee)
}

func NewWithTee[K, V any](cmp func(K, K) int, tee int) *BTree[K, V] {
	return &BTree[K, V]{
		cmp: cmp,
		root: &node[K, V]{
			leaf: true,
		},
		tee: tee,
	}
}

// Get looks for the given key in the tree. It returns the associated value
// and ok=true; otherwise, it returns ok=false.
func (bt *BTree[K, V]) Get(key K) (v V, ok bool) {
	return bt.getFromNode(key, bt.root)
}

// Insert inserts a new key=value pair into the tree. If `key` already exists
// in the tree, its value is replaced with `value`.
func (bt *BTree[K, V]) Insert(key K, value V) {
	//fmt.Printf("inserting %v=%v\n", key, value)
	// If the root node is full, create a new root node with a single child:
	// the old root. Then split.
	if bt.nodeIsFull(bt.root) {
		//fmt.Printf("  root %p is full\n", bt.root)
		oldRoot := bt.root
		bt.root = &node[K, V]{
			leaf:     false,
			children: []*node[K, V]{oldRoot},
		}
		bt.splitChild(bt.root, 0)
	}

	// Here we know for sure that the root is not full.
	bt.insertNonFull(bt.root, nodeKey[K, V]{key: key, value: value})
}

func (bt *BTree[K, V]) getFromNode(key K, n *node[K, V]) (v V, ok bool) {
	kv := nodeKey[K, V]{key: key}
	i, ok := slices.BinarySearchFunc(n.keys, kv, bt.nodeKeyCmp)

	// * If the binary search finds the exact key, we return the value and true.
	// * If the exact key wasn't found:
	//   * If it's a leaf node, the search failed and we return ok=false.
	//   * Otherwise, i tells us the insertion point for key in the keys slice,
	//     meaning that keys[i-i] < key < keys[i]; therefore, we recurse into
	//     children[i], based on the key ordering invariant.
	if ok {
		return n.keys[i].value, true
	}
	if n.leaf {
		return *new(V), false
	}
	return bt.getFromNode(key, n.children[i])
}

// splitChild splits n.children[i] into two children nodes and moves the
// median key of n.children[i] into n. It assumes that n isn't full,
// but n.children[i] is full.
func (bt *BTree[K, V]) splitChild(n *node[K, V], i int) {
	//fmt.Printf("  splitting %p at i=%d\n", n, i)
	// Notation: y is the i-th child of n (the one being split), and z is the
	// new node created to adopt y's t-1 largest keys.
	y := n.children[i]
	//fmt.Printf("  y=%p\n", y)

	if bt.nodeIsFull(n) || !bt.nodeIsFull(y) {
		panic(fmt.Sprintf("expect n to be non-full (got len=%d) and y to be full (got len %d)", len(n.keys), len(y.keys)))
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
	z.keys = make([]nodeKey[K, V], bt.tee-1)
	medianKey := y.keys[bt.tee-1]
	copy(z.keys, y.keys[bt.tee:])
	y.keys = y.keys[:bt.tee-1]

	// Move children to z.
	//
	// The first t children stay with y; the other t children move to z.
	if !y.leaf {
		z.children = make([]*node[K, V], bt.tee)
		copy(z.children, y.children[bt.tee:])
		y.children = y.children[:bt.tee]
	}

	// Place the median key in n
	n.keys = slices.Insert(n.keys, i, medianKey)

	// Place the pointer to z after the pointer to y in n
	n.children = slices.Insert(n.children, i+1, z)
}

// insertNonFull inserts kv into the subtree rooted at n. It assumes n is
// not full.
func (bt *BTree[K, V]) insertNonFull(n *node[K, V], kv nodeKey[K, V]) {
	if bt.nodeIsFull(n) {
		panic("insertNonFull into a full node")
	}
	i, ok := slices.BinarySearchFunc(n.keys, kv, bt.nodeKeyCmp)
	if ok {
		// If this key exists already, replace its value and we're done.
		n.keys[i] = kv
		return
	}

	// The key doesn't exist, and should be inserted at n.keys[i]
	if n.leaf {
		n.keys = slices.Insert(n.keys, i, kv)
	} else {
		// We want to recursively insert kv into n.children[i], but first we have
		// to guarantee that node is not full.
		if bt.nodeIsFull(n.children[i]) {
			bt.splitChild(n, i)
			// We've split n.children[i], and its median key moved up to n.keys[i];
			// compare kv to this key to insert into the proper child.
			if bt.cmp(kv.key, n.keys[i].key) > 0 {
				i++
			}
		}
		bt.insertNonFull(n.children[i], kv)
	}
}

// nodeIsFull says whether n is full, meaning that it has 2t-1 keys.
func (bt *BTree[K, V]) nodeIsFull(n *node[K, V]) bool {
	return len(n.keys) >= 2*bt.tee-1
}

// nodeKeyCmp wraps bt.cmp to work on nodeKey
func (bt *BTree[K, V]) nodeKeyCmp(a, b nodeKey[K, V]) int {
	return bt.cmp(a.key, b.key)
}

// TODO: add deletion
