package node

type BTreeNode interface {
	get(key int32) ([]byte, bool)
	getRange(start, end int32, res [][]byte) [][]byte
	traverse(res [][]byte) [][]byte
	set(key int32, val []byte) (bool, error)
	delete(key int32) (bool, error)
	getParent() *InternalNode
	setParent(parent *InternalNode) error
	split() error
	getNewRoot() BTreeNode
	print(level int)
	verify() (int32, int32)
}

type InternalNode struct {
	Order    uint32
	Parent   *InternalNode
	Keys     []int32
	Children []BTreeNode
}

type LeafNode struct {
	Order  uint32
	Parent *InternalNode
	Next   *LeafNode
	Prev   *LeafNode
	Keys   []int32
	Vals   [][]byte
}
