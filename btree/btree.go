package btree

import (
	"fmt"
	"iter"
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

// node is a node in the BTree.
//
// Key ordering invariants (defining n=len(keys)):
//
// keys[i].key <= keys[i+1].key for each i in 0..n-2
//
// There are n+1 elements in children (0..n)
// if kj is any key in children[j], then:  keys[j-1].key <= kj <= keys[j].key
// Boundaries: k0 <= keys[0].key   and   keys[n-1].key <= kn
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
	// If the root node is full, create a new root node with a single child:
	// the old root. Then split.
	if bt.nodeIsFull(bt.root) {
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

// Delete deletes a key and its associated value from the tree. If key
// is not found in the tree, Delete is a no-op.
func (bt *BTree[K, V]) Delete(key K) {
	var emptyPath treePath[K, V]
	n, idx, path := bt.findNodeForDeletion(bt.root, key, emptyPath)

	if n == nil {
		return
	} else if n.leaf {
		// Deletion from a leaf.
		n.keys = slices.Delete(n.keys, idx, idx+1)
	} else {
		// Deletion from an internal node.
		// n.keys[idx] is the key we want to replace; find the rightmost descendant
		// in the left subtree of this key, and swap its largest key for n.keys[idx]
		pathWithn := append(path, pathPart[K, V]{parent: n, childIndex: idx})
		d, dpath := bt.rightmostDescendant(n.children[idx], pathWithn)
		n.keys[idx] = d.keys[len(d.keys)-1]
		d.keys = d.keys[:len(d.keys)-1]

		// Set n and path for rebalancing.
		n, path = d, dpath
	}

	if n != bt.root {
		bt.rebalance(n, path)
	}
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
	// Notation: y is the i-th child of n (the one being split), and z is the
	// new node created to adopt y's t-1 largest keys.
	y := n.children[i]

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

// nodesPreOrder returns an iterator over the nodes of bt in pre-order.
func (bt *BTree[K, V]) nodesPreOrder() iter.Seq[*node[K, V]] {
	return func(yield func(*node[K, V]) bool) {
		bt.pushPreOrder(yield, bt.root)
	}
}

// pushPreOrder is a recursive push iterator helper for nodesPreorder.
func (bt *BTree[K, V]) pushPreOrder(yield func(*node[K, V]) bool, n *node[K, V]) bool {
	if !yield(n) {
		return false
	}
	for _, c := range n.children {
		if !bt.pushPreOrder(yield, c) {
			return false
		}
	}
	return true
}

// treePath represents a path taken in the tree to get to a specific node.
// If we're currently in node c, we can find its parents and siblings: c is
// implicitly on the top of the path stack. For every node c at stack[j],
// its parent is stack[j-1].parent and c is in that parent's children at
// index stack[j-1].childIndex
type treePath[K, V any] []pathPart[K, V]

// last returns the destructured last element in the path. It panics if
// the path is empty.
func (tp treePath[K, V]) last() (*node[K, V], int) {
	lp := tp[len(tp)-1]
	return lp.parent, lp.childIndex
}

// TODO: have add() and pop() methods here and use them

type pathPart[K, V any] struct {
	parent     *node[K, V]
	childIndex int
}

// findNodeForDeletion finds the node that holds key K, starting at n.
// It returns the found node along with the index of the found key and the
// node's treePath (that doesn't include the node itself). If the key isn't
// found in n or its descendants, the first returned value is nil.
func (bt *BTree[K, V]) findNodeForDeletion(n *node[K, V], key K, path treePath[K, V]) (*node[K, V], int, treePath[K, V]) {
	kv := nodeKey[K, V]{key: key}
	i, ok := slices.BinarySearchFunc(n.keys, kv, bt.nodeKeyCmp)

	// * If the binary search finds the exact key, we return this node and the
	//   found index.
	// * If the exact key wasn't found:
	//   * If it's a leaf node, the search failed.
	//   * Otherwise, i tells us where to recurse into n's children.
	if ok {
		return n, i, path
	}
	if n.leaf {
		return nil, 0, nil
	}

	newPath := append(path, pathPart[K, V]{
		parent:     n,
		childIndex: i,
	})
	return bt.findNodeForDeletion(n.children[i], key, newPath)
}

func (bt *BTree[K, V]) rightmostDescendant(n *node[K, V], path treePath[K, V]) (*node[K, V], treePath[K, V]) {
	if n.leaf {
		return n, path
	}

	last := len(n.children) - 1
	return bt.rightmostDescendant(n.children[last], append(path, pathPart[K, V]{parent: n, childIndex: last}))
}

// rebalance performs B-Tree rebalancing when n doesn't have enough keys after
// deletion.
//
// It currently follows the Deletion algorithm described on Wikipedia.
func (bt *BTree[K, V]) rebalance(n *node[K, V], path treePath[K, V]) {
	if n == bt.root {
		panic("rebalance on root")
	}

	// We only rebalance if n's key count falls below the minimum.
	if len(n.keys) >= bt.tee-1 {
		return
	}

	parent, childIndex := path.last()
	var rightSibling, leftSibling *node[K, V]
	if len(parent.children) > childIndex+1 {
		rightSibling = parent.children[childIndex+1]
	}
	if childIndex > 0 {
		leftSibling = parent.children[childIndex-1]
	}

	// If n's right sibling exists and has enough elements (at least the minimum
	// plus one), rotate left.
	if rightSibling != nil && len(rightSibling.keys) >= bt.tee {
		// 1. Copy the separator from the parent to the end of n
		n.keys = append(n.keys, parent.keys[childIndex])

		// 2. Replace the separator in the parent with the first key of the
		//    right sibling.
		parent.keys[childIndex] = rightSibling.keys[0]
		rightSibling.keys = rightSibling.keys[1:]

		// ... for internal nodes, move the child pointer from the sibling to n
		if !n.leaf {
			n.children = append(n.children, rightSibling.children[0])
			rightSibling.children = rightSibling.children[1:]
		}

		// 3. The tree is now balanced
		return
	}

	// Otherwise, if n's left sibling exists and has enough elements (at least
	// the minimum plus one), rotate right.
	if leftSibling != nil && len(leftSibling.keys) >= bt.tee {
		// 1. Copy the separator from the parent to the start of n
		n.keys = slices.Insert(n.keys, 0, parent.keys[childIndex-1])

		// 2. Replace the separator in the parent with the last key of the
		//    left sibling.
		parent.keys[childIndex-1] = leftSibling.keys[len(leftSibling.keys)-1]
		leftSibling.keys = leftSibling.keys[:len(leftSibling.keys)-1]

		// ... for internal nodes, move the child pointer from the sibling to n
		if !n.leaf {
			n.children = slices.Insert(n.children, 0, leftSibling.children[len(leftSibling.children)-1])
			leftSibling.children = leftSibling.children[len(leftSibling.children)-1:]
		}

		// 3. The tree is now balanced
		return
	}

	// Otherwise, if both siblings have the only the minimum number of elements,
	// then merge with a sibling sandwiching their separator taken off from their
	// parent.
	// Both the n and the sibling have at most tee-1 elements each, so we can
	// safely merge them along with the separator into a single node with no
	// more than 2*tee-1 keys.
	var mergedNode *node[K, V]
	if leftSibling == nil {
		// Merge rightSibling into n

		// 1. Copy the separator to the end of the left node.
		n.keys = append(n.keys, parent.keys[childIndex])

		// 2. Move all elements from the right node to the left node.
		n.keys = append(n.keys, rightSibling.keys...)
		if !n.leaf {
			n.children = append(n.children, rightSibling.children...)
		}

		// 3. Remove the separator from the parent along with its empty right
		//    child.
		parent.keys = slices.Delete(parent.keys, childIndex, childIndex+1)
		parent.children = slices.Delete(parent.children, childIndex+1, childIndex+2)
		mergedNode = n
	} else {
		// Merge n into leftSibling

		// 1. Copy the separator to the end of the left node.
		leftSibling.keys = append(leftSibling.keys, parent.keys[childIndex-1])

		// 2. Move all elements from the right node to the left node.
		leftSibling.keys = append(leftSibling.keys, n.keys...)
		if !n.leaf {
			leftSibling.children = append(leftSibling.children, n.children...)
		}

		// 3. Remove the separator from the parent along with its empty right
		//    child.
		parent.keys = slices.Delete(parent.keys, childIndex-1, childIndex)
		parent.children = slices.Delete(parent.children, childIndex, childIndex+1)
		mergedNode = leftSibling
	}

	if parent == bt.root {
		if len(parent.keys) == 0 {
			// If the parent is the root and now has no elements, then free it and
			// make the merged node the new root.
			bt.root = mergedNode
		}
	} else {
		bt.rebalance(parent, path[:len(path)-1])
	}
}
