package customdb

import "encoding/binary"

const (
	BNODE_NODE = 1 // internal nodes with pointers
	BNODE_LEAF = 2 // leaf nodes with values
)

const (
	BTREE_PAGE_SIZE    = 4096
	BTREE_MAX_KEY_SIZE = 1000
	BTREE_MAX_VAL_SIZE = 3000
)

/// THE NODE FORMAT:
// | type | nkeys |  pointers  |  offsets   | key-values | unused |
// |  2B  |   2B  | nkeys × 8B | nkeys × 2B |     ...    |        |

// where:
// type = node type (leaf or internal)
// nkeys = the number of keys (and the number of child pointers)

// | key-values | -> broken down:
// | key_size | val_size | key | val |
// |    2B    |    2B    | ... | ... |

// each KV pair is prefixed by its size, for internal nodes, the value size is 0

// encoded KV pairs are concatenated. To find the nth KV pair, we have to read all previous
// pairs, unless we store an offset of each KV pair
// e.g leaf node like {"k1": "hi", "k3": "hello"}
// | type | nkeys | pointers | offsets |            key-values           | unused |
// |   2  |   2   | nil nil  |  8 19   | 2 2 "k1" "hi"  2 5 "k3" "hello" |        |
// |  2B  |  2B   |   2×8B   |  2×2B   | 4B + 2B + 2B + 4B + 2B + 5B     |        |

// The offset of the first KV pair is always 0, so its not stored.
// To find the position of the n-th pair, use the offsets[n-1]

// Block Node
type BNode []byte // can be dumped to the disk

// read the fixed-size header
func (node BNode) btype() uint16 {
	return binary.LittleEndian.Uint16(node[0:2])
}

// read the fixed-size header
func (node BNode) nkeys() uint16 {
	return binary.LittleEndian.Uint16(node[2:4])
}

// Write the fixed-size header
func (node BNode) setHeader(btype uint16, nkeys uint16) {
	binary.LittleEndian.PutUint16(node[0:2], btype)
	binary.LittleEndian.PutUint16(node[2:4], nkeys)
}

// read the child pointers array (for internal nodes)
func (node BNode) getPtr(idx uint16) uint64 {
	assert(idx < node.nkeys(), "idx out of bounds")
	pos := 4 + 8*idx
	return binary.LittleEndian.Uint64(node[pos:])
}

// write the child pointers array (for internal nodes)
func (node BNode) setPtr(idx uint16, val uint64) {
	assert(idx < node.nkeys(), "idx out of bounds")
	pos := 4 + 8*idx
	binary.LittleEndian.PutUint64(node[pos:], val)
}

// Read the offsets array to locate the nth key in O(1)
func (node BNode) getOffset(idx uint16) uint16 {
	if idx == 0 {
		return 0
	}

	pos := 4 + 8*node.nkeys() + 2*(idx-1)
	return binary.LittleEndian.Uint16(node[pos:])
}

// Sets the offset
func (node BNode) setOffset(idx uint16, offset uint16) {
	if idx == 0 {
		return
	}
	pos := 4 + 8*node.nkeys() + 2*(idx-1)
	binary.LittleEndian.PutUint16(node[pos+0:], offset)
}

// the offset is adjusted by the data before. So node[node.kvPos(0):] is where the encoded pairs starts
// this returns the position of the nth key using getOffset()
func (node BNode) kvPos(idx uint16) uint16 {
	assert(idx <= node.nkeys(), "idx out of bounds")
	return 4 + 8*node.nkeys() + 2*node.nkeys() + node.getOffset(idx)
}

// Then the KV data is returned as a byte slice, after decoding their sizes.
/*
| key_size | val_size | key | val |
|    2B    |    2B    | ... | ... |
*/
// this gets the nth key data as a slice
func (node BNode) getKey(idx uint16) []byte {
	assert(idx < node.nkeys(), "idx out of bounds")
	pos := node.kvPos(idx)
	klen := binary.LittleEndian.Uint16(node[pos:])
	return node[pos+4:][:klen]
}

// this gets the nth value data as a slice (for leaf nodes)
func (node BNode) getVal(idx uint16) []byte {
	assert(idx < node.nkeys(), "idx out of bounds")
	pos := node.kvPos(idx)
	klen := binary.LittleEndian.Uint16(node[pos+0:])
	vlen := binary.LittleEndian.Uint16(node[pos+2:])
	return node[pos+4+klen:][:vlen]
}

func assert(condition bool, msg string) {
	if !condition {
		panic(msg)
	}
}

func init() {
	node1max := 4 + 1*8 + 1*2 + 4 + BTREE_MAX_KEY_SIZE + BTREE_MAX_VAL_SIZE
	assert(node1max <= BTREE_PAGE_SIZE, "Node size exceeds page size")
}

// Adds a new-key value pair. ASSUMES KEY_VALUE PAIRS ARE SET IN ORDER:
// USES OFFSET OF PREVIOUS KV PAIR
func nodeAppendKV(new BNode, idx uint16, ptr uint64, key []byte, val []byte) {
	// pointers
	new.setPtr(idx, ptr)

	// Key-Value pairs
	pos := new.kvPos(idx) // uses the offset value of the previous key

	// 4-byte KV sizes
	binary.LittleEndian.PutUint16(new[pos+0:], uint16(len(key))) // set the size of the key
	binary.LittleEndian.PutUint16(new[pos+2:], uint16(len(val))) // set the size of the value

	// KV data
	copy(new[pos+4:], key)                  // Set the key
	copy(new[pos+4+uint16(len(key)):], val) // Set the value

	// update the offset value for the next key
	new.setOffset(idx+1, new.getOffset(idx)+4+uint16((len(key)+len(val))))
}
