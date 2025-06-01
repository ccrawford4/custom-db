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

		node.appendKV(uint16(i), 0, key, val)
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

func TestCopyOnWriteBTreeNode(t *testing.T) {
	oldKvs := [][][]byte{
		{[]byte("k1"), []byte("hi")},
		{[]byte("k2"), []byte("hello")},
		{[]byte("k3"), []byte("world")},
	}
	old := createLeafNode(1, oldKvs)

	new := BNode(make([]byte, BTREE_PAGE_SIZE))
	new.setHeader(BNODE_LEAF, 3)

	// Copy-on-write means no inplace updates; updates creates a new node
	new.appendKV(0, 0, old.getKey(0), old.getVal(0))
	new.appendKV(1, 0, []byte("k2"), []byte("updated"))
	new.appendKV(2, 0, old.getKey(2), old.getVal(2))

	// Verify the new node has the updated value
	if !bytes.Equal(new.getKey(1), []byte("k2")) {
		t.Errorf("[ERROR]: Expected k2, got %s", new.getKey(1))
	}
	if !bytes.Equal(new.getVal(1), []byte("updated")) {
		t.Errorf("[ERROR]: Expected updated, got %s", new.getVal(1))
	}

	// Verify the old node is unchanged
	if !bytes.Equal(old.getVal(1), []byte("hello")) {
		t.Errorf("[ERROR]: Old node should be unchanged, expected hello, got %s", old.getVal(1))
	}

	// Verify other keys are copied correctly
	if !bytes.Equal(new.getKey(0), []byte("k1")) {
		t.Errorf("[ERROR]: Expected k1, got %s", new.getKey(0))
	}
	if !bytes.Equal(new.getVal(0), []byte("hi")) {
		t.Errorf("[ERROR]: Expected hi, got %s", new.getVal(0))
	}
	if !bytes.Equal(new.getKey(2), []byte("k3")) {
		t.Errorf("[ERROR]: Expected k3, got %s", new.getKey(2))
	}
	if !bytes.Equal(new.getVal(2), []byte("world")) {
		t.Errorf("[ERROR]: Expected world, got %s", new.getVal(2))
	}
}

func TestCopyOnWriteAndDeleteBTreeNode(t *testing.T) {
	oldKvs := [][][]byte{
		{[]byte("k1"), []byte("hi")},
		{[]byte("k2"), []byte("hello")},
		{[]byte("k3"), []byte("world")},
	}
	old := createLeafNode(1, oldKvs)

	new := BNode(make([]byte, BTREE_PAGE_SIZE))
	new.setHeader(BNODE_LEAF, 2)

	new.appendKV(0, 0, old.getKey(0), old.getVal(0))
	new.appendKV(1, 0, old.getKey(2), old.getVal(2))

	// Verify the new node has the updated value
	if !bytes.Equal(new.getKey(0), []byte("k1")) {
		t.Errorf("[ERROR]: Expected hi, got %s", new.getKey(1))
	}
	if !bytes.Equal(new.getVal(0), []byte("hi")) {
		t.Errorf("[ERROR]: Expected hi, got %s", new.getVal(1))
	}

	// Verify other keys are copied correctly
	if !bytes.Equal(new.getKey(1), []byte("k3")) {
		t.Errorf("[ERROR]: Expected k3, got %s", new.getKey(0))
	}
	if !bytes.Equal(new.getVal(1), []byte("world")) {
		t.Errorf("[ERROR]: Expected world, got %s", new.getVal(0))
	}
}

func TestCopyOnWriteAndNewInsertBTreeNode(t *testing.T) {
	oldKvs := [][][]byte{
		{[]byte("k1"), []byte("hi")},
		{[]byte("k2"), []byte("hello")},
		{[]byte("k3"), []byte("world")},
	}
	old := createLeafNode(1, oldKvs)

	new := BNode(make([]byte, 2*BTREE_PAGE_SIZE)) // larger
	new.setHeader(BNODE_LEAF, 4)

	new.appendKV(0, 0, []byte("a"), []byte("b"))
	new.appendKV(1, 0, old.getKey(0), old.getVal(0))
	new.appendKV(2, 0, old.getKey(1), old.getVal(1))
	new.appendKV(3, 0, old.getKey(2), old.getVal(2))

	// Verify the new node has the inserted key-value pair
	if !bytes.Equal(new.getKey(0), []byte("a")) {
		t.Errorf("[ERROR]: Expected a, got %s", new.getKey(0))
	}
	if !bytes.Equal(new.getVal(0), []byte("b")) {
		t.Errorf("[ERROR]: Expected b, got %s", new.getVal(0))
	}

	// Verify the copied keys and values from old node
	if !bytes.Equal(new.getKey(1), []byte("k1")) {
		t.Errorf("[ERROR]: Expected k1, got %s", new.getKey(1))
	}
	if !bytes.Equal(new.getVal(1), []byte("hi")) {
		t.Errorf("[ERROR]: Expected hi, got %s", new.getVal(1))
	}
	if !bytes.Equal(new.getKey(2), []byte("k2")) {
		t.Errorf("[ERROR]: Expected k2, got %s", new.getKey(2))
	}
	if !bytes.Equal(new.getVal(2), []byte("hello")) {
		t.Errorf("[ERROR]: Expected hello, got %s", new.getVal(2))
	}
	if !bytes.Equal(new.getKey(3), []byte("k3")) {
		t.Errorf("[ERROR]: Expected k3, got %s", new.getKey(3))
	}
	if !bytes.Equal(new.getVal(3), []byte("world")) {
		t.Errorf("[ERROR]: Expected world, got %s", new.getVal(3))
	}
}

func TestLeafUpsertUpdate(t *testing.T) {
	oldKvs := [][][]byte{
		{[]byte("k1"), []byte("val1")},
		{[]byte("k3"), []byte("val3")},
		{[]byte("k5"), []byte("val5")},
	}
	old := createLeafNode(1, oldKvs)

	new := BNode(make([]byte, BTREE_PAGE_SIZE))
	new.leafUpsert(old, []byte("k3"))

	if new.nkeys() != 3 {
		t.Errorf("[ERROR]: Expected 3 keys, got %d", new.nkeys())
	}

	if !bytes.Equal(new.getKey(1), []byte("k3")) {
		t.Errorf("[ERROR]: Expected k3, got %s", new.getKey(1))
	}
	if !bytes.Equal(new.getVal(1), []byte("val3")) {
		t.Errorf("[ERROR]: Expected val3, got %s", new.getVal(1))
	}
}

func TestLeafUpsertUpdateSingle(t *testing.T) {
	oldKvs := [][][]byte{
		{[]byte("k2"), []byte("val2")},
	}
	old := createLeafNode(1, oldKvs)

	new := BNode(make([]byte, BTREE_PAGE_SIZE))
	new.leafUpsert(old, []byte("k2"))

	if new.nkeys() != 1 {
		t.Errorf("[ERROR]: Expected 1 key, got %d", new.nkeys())
	}

	if !bytes.Equal(new.getKey(0), []byte("k2")) {
		t.Errorf("[ERROR]: Expected k2, got %s", new.getKey(0))
	}
	if !bytes.Equal(new.getVal(0), []byte("val2")) {
		t.Errorf("[ERROR]: Expected val2, got %s", new.getVal(0))
	}
}

func TestLeafUpsertUpdateFirst(t *testing.T) {
	oldKvs := [][][]byte{
		{[]byte("k1"), []byte("val1")},
		{[]byte("k3"), []byte("val3")},
	}
	old := createLeafNode(1, oldKvs)

	new := BNode(make([]byte, BTREE_PAGE_SIZE))
	new.leafUpsert(old, []byte("k1"))

	if new.nkeys() != 2 {
		t.Errorf("[ERROR]: Expected 2 keys, got %d", new.nkeys())
	}

	if !bytes.Equal(new.getKey(0), []byte("k1")) {
		t.Errorf("[ERROR]: Expected k1, got %s", new.getKey(0))
	}
	if !bytes.Equal(new.getVal(0), []byte("val1")) {
		t.Errorf("[ERROR]: Expected val1, got %s", new.getVal(0))
	}
}
