package main

import (
	"testing"
)

func TestBasic(t *testing.T) {
	// Initalize the node
	node := BNode(make([]byte, BTREE_PAGE_SIZE))

	// Set the header
	node.setHeader(BNODE_LEAF, 2) // A Leaf node with 2 keys

	// Add a key-value pair to the node
	nodeAppendKV(node, 0, 0, []byte("k1"), []byte("hi"))
	// First '0' is index, second '0' is 'ptr' and unused for leaf nodes

	nodeAppendKV(node, 1, 0, []byte("k2"), []byte("hello"))
}
