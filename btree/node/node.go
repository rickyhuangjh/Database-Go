package node


type BTreeNode interface {
	Get(c NodeCache, key uint64) (uint64, bool)
	GetRange(c NodeCache, start, end uint64, res []uint64) []uint64
	Traverse(c NodeCache, res []uint64) []uint64
	Set(c NodeCache, key uint64, val uint64) (bool, error)
	Delete(c NodeCache, key uint64) (bool, error)
	Print(level int)
	Verify() (uint64, uint64)
	split(c NodeCache) error
	merge(c NodeCache) error
}

type InternalNode struct {
	ID uint64
	ParentID uint64
	Keys     []uint64
	ChildIDs []uint64
}



type LeafNode struct {
	ID uint64
	ParentID uint64
	NextID uint64
	PrevID   uint64
	Keys   []uint64
	Vals   []uint64
}


