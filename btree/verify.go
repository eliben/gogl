package btree

import (
	"errors"
	"fmt"
)

// verify checks B-tree invariants on bt and returns an error combining all
// the problems encountered. Returns nil if bt is ok.
func (bt *BTree[K, V]) verify() error {
	var errs []error

	for n := range bt.nodesPreOrder() {
		if err := bt.verifyNode(n); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (bt *BTree[K, V]) verifyNode(n *node[K, V]) error {
	if len(n.keys) > 2*bt.tee-1 {
		return fmt.Errorf("node %p has too many keys: %d", n, len(n.keys))
	}
	if n != bt.root && len(n.keys) < bt.tee-1 {
		return fmt.Errorf("node %p has too few keys: %d", n, len(n.keys))
	}

	if !n.leaf && len(n.children) != len(n.keys)+1 {
		return fmt.Errorf("internal node %p has %d keys but %d children", n, len(n.keys), len(n.children))
	}

	return nil
}
