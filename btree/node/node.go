package node

type V struct {
	Block uint32
	Offset uint32
}

type BTreeNode interface {
	Get(key uint64) (V, bool)
	GetRange(start, end uint64, res []V) []V
	Traverse(res []V) []V
	Set(key uint64, val V) (bool, error)
	Delete(key uint64) (bool, error)
	GetParent() *InternalNode
	SetParent(parent *InternalNode) error
	GetNewRoot() BTreeNode
	Print(level int)
	Verify() (uint64, uint64)
	split() error
	merge() error
}

type InternalNode struct {
	Order    uint64
	Parent   *InternalNode
	Keys     []uint64
	Children []BTreeNode
}

type LeafNode struct {
	Order  uint64
	Parent *InternalNode
	Next   *LeafNode
	Prev   *LeafNode
	Keys   []uint64
	Vals   []V
}


