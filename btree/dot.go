package btree

import (
	"fmt"
	"iter"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

const prefix = `digraph g {
node [shape = record,height=0.1];
`

// renderDotToImage generates a dot graph for GraphViz from bt, and invokes
// dot to create an image from it; the output image file is provided.
// This function will panic if it's unable to invoke the `dot` command-line
// tool or if that tool fails for some reason.
func (bt *BTree[K, V]) renderDotToImage(outfilename string) {
	ds := bt.renderDot()
	absPath, err := filepath.Abs(outfilename)
	if err != nil {
		panic(err)
	}

	dotCmd := exec.Command("dot", "-Tpng", "-o", absPath)
	dotIn, _ := dotCmd.StdinPipe()
	if err := dotCmd.Start(); err != nil {
		panic(err)
	}
	dotIn.Write([]byte(ds))
	dotIn.Close()
	if err := dotCmd.Wait(); err != nil {
		panic(err)
	}

	log.Println("renderDotToImage wrote", absPath)
}

// renderDot generates a dot graph for Graphviz from bt, and returns it as
// a string.
func (bt *BTree[K, V]) renderDot() string {
	// nodeNames maps a node pointer to a unique name like "nodeN" where N
	// is auto-incremented any time a new node pointer is encountered.
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

	var sb strings.Builder
	sb.WriteString(prefix)
	for n := range bt.nodesPreOrder() {
		var labelParts []string
		// Build label for this node: alternating child nodes with key nodes.
		for i := 0; i < len(n.keys)+1; i++ {
			labelParts = append(labelParts, fmt.Sprintf("<f%d>", i))
			if i < len(n.keys) {
				labelParts = append(labelParts, fmt.Sprintf("|%v|", n.keys[i].key))
			}
		}
		sb.WriteString(fmt.Sprintf(`%s[label = "%s"];`+"\n", node2name(n), strings.Join(labelParts, "")))

		// Print the links from this node to other nodes.
		for i := 0; i < len(n.children); i++ {
			link := fmt.Sprintf("<f%d>", i)
			sb.WriteString(fmt.Sprintf(`"%s":%s -> "%s"`+"\n", node2name(n), link, node2name(n.children[i])))
		}
	}
	sb.WriteString("}\n")

	return sb.String()
}

// nodesPreOrder returns an iterator over the nodes of bt in pre-order.
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
