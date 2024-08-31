package btree

import (
	"fmt"
	"iter"
	"strings"
)

const prefix = `digraph g {
node [shape = record,height=0.1];
`

func (bt *BTree[K, V]) renderDot() string {
	nodeNames := make(map[*node[K, V]]string)
	nodeNumber := 0

	node2name := func(n *node[K, V]) string {
		if name, ok := nodeNames[n]; ok {
			return name
		}
		nodeNames[n] = fmt.Sprintf("node%d", nodeNumber)
		nodeNumber++
		return nodeNames[n]
	}

	// TODO: continue here, emit node contents followed by child pointers
	var sb strings.Builder
	sb.WriteString(prefix)
	for n := range bt.nodesPreOrder() {
		sb.WriteString(fmt.Sprintf("%s\n", node2name(n)))
	}
	sb.WriteString("}\n")

	return sb.String()
}

func (bt *BTree[K, V]) nodesPreOrder() iter.Seq[*node[K, V]] {
	return func(yield func(*node[K, V]) bool) {
		bt.pushPreOrder(yield, bt.root)
	}
}

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
