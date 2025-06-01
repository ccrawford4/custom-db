package customdb

import (
	"bytes"
	"testing"
)

/*  kvs:
*  [
*  		[[]byte("k1"), []byte("val1")],
*  		[[]byte("k2"), []byte("val2")],
*  ]
 */
func createLeafNode(size uint16, kvs [][][]byte) BNode {
	node := BNode(make([]byte, BTREE_PAGE_SIZE*size))
	node.setHeader(BNODE_LEAF, uint16(len(kvs)))

	for i := range kvs {
		entry := kvs[i]
		key := entry[0]
		val := entry[1]

		nodeAppendKV(node, uint16(i), 0, key, val)
	}

	return node
}

func TestConstructBTreeNode(t *testing.T) {
	kvs := [][][]byte{
		{[]byte("k1"), []byte("hi")},
		{[]byte("k2"), []byte("hello")},
	}
	node := createLeafNode(uint16(1), kvs)
	if node.btype() != BNODE_LEAF {
		t.Errorf("[ERROR]: Expected BNODE_LEAF for the type")
	}
	key1 := node.getKey(0)
	val1 := node.getVal(0)
	if !bytes.Equal([]byte("k1"), key1) {
		t.Errorf("[ERROR]: Expected %s, Received %s", "k1", key1)
	}
	if !bytes.Equal([]byte("hi"), val1) {
		t.Errorf("[ERROR]: Expected %s, Received %s", "hi", val1)
	}
	key2 := node.getKey(1)
	val2 := node.getVal(1)
	if !bytes.Equal([]byte("k2"), key2) {
		t.Errorf("[ERROR]: Expected %s, Received %s", "k2", key2)
	}
	if !bytes.Equal([]byte("hello"), val2) {
		t.Errorf("[ERROR]: Expected %s, Received %s", "hello", val2)
	}
}
