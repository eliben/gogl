package btree

import (
	"errors"
	"fmt"
	"slices"
)

// verify checks B-tree invariants on bt and returns an error combining all
// the problems encountered. Returns nil if bt is ok.
func (bt *BTree[K, V]) verify() error {
	var errs []error

	// Verify invariants on each node separately
	for n := range bt.nodesPreOrder() {
		if err := bt.verifyNode(n); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	// Recursive visit to calculate height of all leaf nodes in the tree.
	var heights []int
	var visit func(n *node[K, V], h int)
	visit = func(n *node[K, V], h int) {
		if n.leaf {
			heights = append(heights, h)
			return
		}

		for _, c := range n.children {
			visit(c, h+1)
		}
	}
	visit(bt.root, 0)

	cp := slices.Compact(heights)
	if len(cp) != 1 {
		return fmt.Errorf("different leaf heights: %v", cp)
	}

	return nil
}

func (bt *BTree[K, V]) verifyNode(n *node[K, V]) error {
	if len(n.keys) > 2*bt.tee-1 {
		return fmt.Errorf("node %p: has too many keys: %d", n, len(n.keys))
	}
	if n != bt.root && len(n.keys) < bt.tee-1 {
		return fmt.Errorf("node %p: has too few keys: %d", n, len(n.keys))
	}

	if n == bt.root && len(n.keys) < 1 {
		return fmt.Errorf("node %p: root has 0 keys", n)
	}

	if !slices.IsSortedFunc(n.keys, bt.nodeKeyCmp) {
		return fmt.Errorf("node %p: keys not sorted", n)
	}

	if !n.leaf {
		if len(n.children) != len(n.keys)+1 {
			return fmt.Errorf("internal node %p: %d keys but %d children", n, len(n.keys), len(n.children))
		}

		for j, k := range n.keys {
			// Verify ordering invariant (see type node's comment).

			// 1: ensure that kj <= keys[j].key
			for ci, kj := range n.children[j].keys {
				if bt.nodeKeyCmp(kj, k) > 0 {
					return fmt.Errorf("node %p: key %v of child [%d] >= key %v of node[%d]", n, kj.key, j, k.key, ci)
				}
			}

			// 2: ensure that keys[j-1].key <= kj
			for ci, kj := range n.children[j+1].keys {
				if bt.nodeKeyCmp(k, kj) > 0 {
					return fmt.Errorf("node %p: key %v of child [%d] <= key %v of node[%d]", n, kj.key, j, k.key, ci)
				}
			}
		}
	}

	return nil
}
