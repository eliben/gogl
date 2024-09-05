package btree

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

// push returns a new path with an appended pathPart that has the
// given parent & childIndex.
func (tp treePath[K, V]) push(parent *node[K, V], childIndex int) treePath[K, V] {
	return append(tp, pathPart[K, V]{
		parent:     parent,
		childIndex: childIndex,
	})
}

// pop returns the path without its last element. It panics if the path is
// empty.
func (tp treePath[K, V]) pop() treePath[K, V] {
	return tp[:len(tp)-1]
}

type pathPart[K, V any] struct {
	parent     *node[K, V]
	childIndex int
}
